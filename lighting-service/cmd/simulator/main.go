package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type TrafficSimulator struct {
	client         mqtt.Client
	tunnelID       string
	currentWeather string
}

func NewTrafficSimulator(broker string, tunnelID string) *TrafficSimulator {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID("traffic-simulator-" + tunnelID)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", token.Error())
	}

	return &TrafficSimulator{
		client:         client,
		tunnelID:       tunnelID,
		currentWeather: "CLEAR",
	}
}

func (ts *TrafficSimulator) PublishTrafficData() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		data := map[string]interface{}{
			"tunnel_id":     ts.tunnelID,
			"vehicle_count": rand.Intn(50) + 10,
			"avg_speed":     float64(rand.Intn(40) + 60),
			"timestamp":     time.Now().Unix(),
		}

		payload, _ := json.Marshal(data)
		topic := fmt.Sprintf("tunnel/%s/traffic", ts.tunnelID)
		ts.client.Publish(topic, 0, false, payload)
		log.Printf("Published traffic data for tunnel %s", ts.tunnelID)
	}
}

func (ts *TrafficSimulator) PublishLightIntensityData() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		hour := time.Now().Hour()
		var outsideLux float64

		switch {
		case hour >= 6 && hour < 8:
			outsideLux = float64(rand.Intn(2000) + 1000)
		case hour >= 8 && hour < 17:
			outsideLux = float64(rand.Intn(3000) + 5000)
		case hour >= 17 && hour < 19:
			outsideLux = float64(rand.Intn(1500) + 500)
		default:
			outsideLux = float64(rand.Intn(50) + 10)
		}

		if ts.currentWeather == "HEAVY_RAIN" || ts.currentWeather == "STORM" {
			outsideLux *= 0.4
		} else if ts.currentWeather == "HEAVY_SNOW" || ts.currentWeather == "BLIZZARD" {
			outsideLux *= 0.5
		} else if ts.currentWeather == "LIGHT_RAIN" || ts.currentWeather == "LIGHT_SNOW" {
			outsideLux *= 0.7
		} else if ts.currentWeather == "CLOUDY" {
			outsideLux *= 0.6
		}

		insideLux := outsideLux * 0.3

		data := map[string]interface{}{
			"tunnel_id":   ts.tunnelID,
			"outside_lux": outsideLux,
			"inside_lux":  insideLux,
			"timestamp":   time.Now().Unix(),
		}

		payload, _ := json.Marshal(data)
		topic := fmt.Sprintf("tunnel/%s/light_intensity", ts.tunnelID)
		ts.client.Publish(topic, 0, false, payload)
		log.Printf("Published light intensity for tunnel %s: %.0f lux", ts.tunnelID, outsideLux)
	}
}

func (ts *TrafficSimulator) SimulateWeather() {
	weatherConditions := []string{
		"CLEAR", "CLOUDY", "LIGHT_RAIN", "HEAVY_RAIN",
		"STORM", "LIGHT_SNOW", "HEAVY_SNOW", "BLIZZARD",
	}

	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if rand.Float32() < 0.3 {
			newWeather := weatherConditions[rand.Intn(len(weatherConditions))]
			if newWeather != ts.currentWeather {
				ts.currentWeather = newWeather
				ts.publishWeatherAlert(newWeather)
			}
		}
	}
}

func (ts *TrafficSimulator) publishWeatherAlert(condition string) {
	severity := int32(1)
	description := ""

	switch condition {
	case "LIGHT_RAIN", "LIGHT_SNOW":
		severity = 1
		description = "轻微降水，请注意行车安全"
	case "HEAVY_RAIN", "HEAVY_SNOW":
		severity = 2
		description = "强降水，隧道入口已提高亮度"
	case "STORM", "BLIZZARD":
		severity = 3
		description = "恶劣天气警报，隧道入口亮度已提升至最高"
	case "CLOUDY":
		severity = 1
		description = "阴天，视线略有影响"
	default:
		severity = 0
		description = "天气转好，亮度恢复正常"
	}

	data := map[string]interface{}{
		"tunnel_id":         ts.tunnelID,
		"condition":         condition,
		"severity":          severity,
		"description":       description,
		"expected_duration": 3600,
		"timestamp":         time.Now().Unix(),
	}

	payload, _ := json.Marshal(data)
	topic := fmt.Sprintf("tunnel/%s/weather", ts.tunnelID)
	ts.client.Publish(topic, 1, false, payload)
	log.Printf("Published weather alert for tunnel %s: %s (severity: %d)", ts.tunnelID, condition, severity)
}

func main() {
	simulator := NewTrafficSimulator("tcp://localhost:1883", "T001")

	go simulator.PublishTrafficData()
	go simulator.PublishLightIntensityData()
	go simulator.SimulateWeather()

	log.Println("Data simulator started - generating traffic, light intensity, and weather data")
	select {}
}
