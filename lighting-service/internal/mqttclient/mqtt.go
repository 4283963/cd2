package mqttclient

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"lighting-service/internal/model"
)

type Config struct {
	Broker   string
	ClientID string
	Username string
	Password string
	Topics   map[string]string
}

type Client struct {
	client       mqtt.Client
	config       Config
	trafficChan  chan model.TrafficData
	lightChan    chan model.LightIntensityData
	emergencyChan chan model.EmergencyType
}

func NewClient(config Config) *Client {
	return &Client{
		config:       config,
		trafficChan:  make(chan model.TrafficData, 100),
		lightChan:    make(chan model.LightIntensityData, 100),
		emergencyChan: make(chan model.EmergencyType, 10),
	}
}

func (c *Client) Connect() error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(c.config.Broker)
	opts.SetClientID(c.config.ClientID)
	if c.config.Username != "" {
		opts.SetUsername(c.config.Username)
		opts.SetPassword(c.config.Password)
	}
	opts.SetAutoReconnect(true)
	opts.SetReconnectingHandler(func(client mqtt.Client, opts *mqtt.ClientOptions) {
		log.Println("MQTT reconnecting...")
	})
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Println("MQTT connected")
		c.subscribeTopics()
	})

	c.client = mqtt.NewClient(opts)

	token := c.client.Connect()
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("mqtt connect failed: %w", token.Error())
	}

	return nil
}

func (c *Client) subscribeTopics() {
	trafficTopic := c.config.Topics["traffic"]
	token := c.client.Subscribe(trafficTopic, 0, c.handleTrafficData)
	token.Wait()
	log.Printf("Subscribed to topic: %s", trafficTopic)

	lightTopic := c.config.Topics["light_intensity"]
	token = c.client.Subscribe(lightTopic, 0, c.handleLightIntensityData)
	token.Wait()
	log.Printf("Subscribed to topic: %s", lightTopic)

	emergencyTopic := c.config.Topics["emergency"]
	token = c.client.Subscribe(emergencyTopic, 1, c.handleEmergencyData)
	token.Wait()
	log.Printf("Subscribed to topic: %s", emergencyTopic)
}

func (c *Client) handleTrafficData(client mqtt.Client, msg mqtt.Message) {
	var data model.TrafficData
	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		log.Printf("Failed to parse traffic data: %v", err)
		return
	}
	data.Timestamp = time.Now()
	select {
	case c.trafficChan <- data:
	default:
		log.Println("Traffic data channel full, dropping message")
	}
}

func (c *Client) handleLightIntensityData(client mqtt.Client, msg mqtt.Message) {
	var data model.LightIntensityData
	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		log.Printf("Failed to parse light intensity data: %v", err)
		return
	}
	data.Timestamp = time.Now()
	select {
	case c.lightChan <- data:
	default:
		log.Println("Light intensity data channel full, dropping message")
	}
}

func (c *Client) handleEmergencyData(client mqtt.Client, msg mqtt.Message) {
	var emergency struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(msg.Payload(), &emergency); err != nil {
		log.Printf("Failed to parse emergency data: %v", err)
		return
	}

	var etype model.EmergencyType
	switch emergency.Type {
	case "ACCIDENT":
		etype = model.EmergencyAccident
	case "FIRE":
		etype = model.EmergencyFire
	case "FAULT":
		etype = model.EmergencyFault
	default:
		etype = model.EmergencyNone
	}

	select {
	case c.emergencyChan <- etype:
	default:
		log.Println("Emergency data channel full, dropping message")
	}
}

func (c *Client) TrafficChan() <-chan model.TrafficData {
	return c.trafficChan
}

func (c *Client) LightChan() <-chan model.LightIntensityData {
	return c.lightChan
}

func (c *Client) EmergencyChan() <-chan model.EmergencyType {
	return c.emergencyChan
}

func (c *Client) Publish(topic string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	token := c.client.Publish(topic, 0, false, data)
	token.Wait()
	return token.Error()
}

func (c *Client) Disconnect() {
	if c.client != nil {
		c.client.Disconnect(250)
	}
	close(c.trafficChan)
	close(c.lightChan)
	close(c.emergencyChan)
}
