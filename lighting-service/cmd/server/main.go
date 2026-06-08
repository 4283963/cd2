package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"

	"lighting-service/internal/collector"
	"lighting-service/internal/dimmer"
	"lighting-service/internal/grpcserver"
	"lighting-service/internal/mqttclient"
	"lighting-service/internal/model"
	"lighting-service/internal/storage"
)

type Config struct {
	Server struct {
		Name     string `yaml:"name"`
		GrpcPort int    `yaml:"grpc_port"`
		HttpPort int    `yaml:"http_port"`
	} `yaml:"server"`
	MQTT struct {
		Broker   string            `yaml:"broker"`
		ClientID string            `yaml:"client_id"`
		Username string            `yaml:"username"`
		Password string            `yaml:"password"`
		Topics   map[string]string `yaml:"topics"`
	} `yaml:"mqtt"`
	Redis struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
		CacheTTL int    `yaml:"cache_ttl"`
	} `yaml:"redis"`
	MySQL struct {
		Host         string `yaml:"host"`
		Port         int    `yaml:"port"`
		Database     string `yaml:"database"`
		Username     string `yaml:"username"`
		Password     string `yaml:"password"`
		MaxOpenConns int    `yaml:"max_open_conns"`
		MaxIdleConns int    `yaml:"max_idle_conns"`
	} `yaml:"mysql"`
	Dimmer struct {
		EntranceSegmentID    string  `yaml:"entrance_segment_id"`
		WeatherBrightnessBoost float64 `yaml:"weather_brightness_boost"`
		SevereWeatherBoost   float64 `yaml:"severe_weather_boost"`
		Segments []struct {
			ID            string  `yaml:"id"`
			Name          string  `yaml:"name"`
			BaseBrightness float64 `yaml:"base_brightness"`
			Length        float64 `yaml:"length"`
		} `yaml:"segments"`
		DayFactor            float64 `yaml:"day_brightness_factor"`
		NightFactor          float64 `yaml:"night_brightness_factor"`
		CloudyFactor         float64 `yaml:"cloudy_brightness_factor"`
		EmergencyFactor      float64 `yaml:"emergency_brightness_factor"`
		TrafficDensityFactor float64 `yaml:"traffic_density_factor"`
		MinBrightness        float64 `yaml:"min_brightness"`
		MaxBrightness        float64 `yaml:"max_brightness"`
	} `yaml:"dimmer"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config failed: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse config failed: %w", err)
	}

	return &config, nil
}

func main() {
	configPath := "config/config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Load config failed: %v", err)
	}
	log.Println("Config loaded successfully")

	storageCfg := storage.Config{
		RedisHost:     config.Redis.Host,
		RedisPort:     config.Redis.Port,
		RedisPassword: config.Redis.Password,
		RedisDB:       config.Redis.DB,
		CacheTTL:      config.Redis.CacheTTL,
		MySQLHost:     config.MySQL.Host,
		MySQLPort:     config.MySQL.Port,
		MySQLDatabase: config.MySQL.Database,
		MySQLUsername: config.MySQL.Username,
		MySQLPassword: config.MySQL.Password,
		MaxOpenConns:  config.MySQL.MaxOpenConns,
		MaxIdleConns:  config.MySQL.MaxIdleConns,
	}

	store, err := storage.NewStorage(storageCfg)
	if err != nil {
		log.Printf("Warning: storage init failed: %v", err)
	}
	defer store.Close()

	dataCollector := collector.NewDataCollector()

	segments := make([]model.LightingSegment, len(config.Dimmer.Segments))
	for i, seg := range config.Dimmer.Segments {
		segments[i] = model.LightingSegment{
			ID:         seg.ID,
			Name:       seg.Name,
			Brightness: seg.BaseBrightness,
			Length:     seg.Length,
		}
	}

	dimmerConfig := &model.DimmerConfig{
		Segments:             segments,
		DayFactor:            config.Dimmer.DayFactor,
		NightFactor:          config.Dimmer.NightFactor,
		CloudyFactor:         config.Dimmer.CloudyFactor,
		EmergencyFactor:      config.Dimmer.EmergencyFactor,
		TrafficDensityFactor: config.Dimmer.TrafficDensityFactor,
		MinBrightness:        config.Dimmer.MinBrightness,
		MaxBrightness:        config.Dimmer.MaxBrightness,
		EntranceSegmentID:    config.Dimmer.EntranceSegmentID,
		WeatherBrightnessBoost: config.Dimmer.WeatherBrightnessBoost,
		SevereWeatherBoost:   config.Dimmer.SevereWeatherBoost,
	}

	dimmerService := dimmer.NewDimmerService(dimmerConfig, dataCollector)
	dimmerService.InitializeSegments("T001")

	mqttCfg := mqttclient.Config{
		Broker:   config.MQTT.Broker,
		ClientID: config.MQTT.ClientID,
		Username: config.MQTT.Username,
		Password: config.MQTT.Password,
		Topics:   config.MQTT.Topics,
	}

	mqttClient := mqttclient.NewClient(mqttCfg)
	if err := mqttClient.Connect(); err != nil {
		log.Printf("Warning: mqtt connect failed: %v", err)
	}
	defer mqttClient.Disconnect()

	dataCollector.Start(mqttClient.TrafficChan(), mqttClient.LightChan())

	go func() {
		for emergencyType := range mqttClient.EmergencyChan() {
			dimmerService.SetEmergencyMode("T001", emergencyType)
			if store != nil {
				store.SetEmergencyMode("T001", emergencyType)
			}
			segments, _ := dimmerService.CalculateBrightness("T001")
			if store != nil {
				store.CacheLightingSegments("T001", segments)
			}
		}
	}()

	go func() {
		for weatherAlert := range mqttClient.WeatherChan() {
			tunnelID := weatherAlert.TunnelID
			if tunnelID == "" {
				tunnelID = "T001"
			}

			if _, ok := dimmerService.GetSegments(tunnelID); !ok {
				dimmerService.InitializeSegments(tunnelID)
			}

			dimmerService.SetWeatherCondition(tunnelID, weatherAlert.Condition)

			segments, err := dimmerService.CalculateBrightness(tunnelID)
			if err != nil {
				log.Printf("Calculate brightness after weather alert failed: %v", err)
				continue
			}

			if store != nil {
				if err := store.CacheLightingSegments(tunnelID, segments); err != nil {
					log.Printf("Cache lighting segments after weather alert failed: %v", err)
				}
			}

			log.Printf("Weather alert processed for tunnel %s: %s", tunnelID, weatherAlert.Condition.String())
		}
	}()

	go startBrightnessCalculator(dimmerService, store, dataCollector)

	grpcServer := grpcserver.NewLightingServer(dimmerService, store)
	go func() {
		port := fmt.Sprintf(":%d", config.Server.GrpcPort)
		if err := grpcServer.Start(port); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	log.Printf("Lighting service started - gRPC on port %d", config.Server.GrpcPort)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh

	log.Printf("Received signal %v, shutting down...", sig)
	grpcServer.Stop()
	log.Println("Server stopped")
}

func startBrightnessCalculator(dimmerService *dimmer.DimmerService, store *storage.Storage, collector *collector.DataCollector) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		tunnels := collector.GetAllTunnels()
		for _, tunnelID := range tunnels {
			if _, ok := dimmerService.GetSegments(tunnelID); !ok {
				dimmerService.InitializeSegments(tunnelID)
			}

			segments, err := dimmerService.CalculateBrightness(tunnelID)
			if err != nil {
				log.Printf("Calculate brightness for tunnel %s failed: %v", tunnelID, err)
				continue
			}

			if store != nil {
				if err := store.CacheLightingSegments(tunnelID, segments); err != nil {
					log.Printf("Cache lighting segments failed: %v", err)
				}
			}

			log.Printf("Updated brightness for tunnel %s, %d segments", tunnelID, len(segments))
		}
	}
}
