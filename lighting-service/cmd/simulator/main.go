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
	client   mqtt.Client
	tunnelID string
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
		client:   client,
		tunnelID: tunnelID,
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

func main() {
	simulator := NewTrafficSimulator("tcp://localhost:1883", "T001")

	go simulator.PublishTrafficData()
	go simulator.PublishLightIntensityData()

	select {}
}
