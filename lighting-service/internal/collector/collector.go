package collector

import (
	"log"
	"sync"
	"time"

	"lighting-service/internal/model"
)

type DataCollector struct {
	mu                sync.RWMutex
	trafficData       map[string]model.TrafficData
	lightIntensityData map[string]model.LightIntensityData
	trafficHistory    map[string][]model.TrafficData
	lightHistory      map[string][]model.LightIntensityData
	maxHistorySize    int
}

func NewDataCollector() *DataCollector {
	return &DataCollector{
		trafficData:       make(map[string]model.TrafficData),
		lightIntensityData: make(map[string]model.LightIntensityData),
		trafficHistory:    make(map[string][]model.TrafficData),
		lightHistory:      make(map[string][]model.LightIntensityData),
		maxHistorySize:    100,
	}
}

func (dc *DataCollector) Start(trafficCh <-chan model.TrafficData, lightCh <-chan model.LightIntensityData) {
	go func() {
		for data := range trafficCh {
			dc.collectTraffic(data)
		}
	}()

	go func() {
		for data := range lightCh {
			dc.collectLightIntensity(data)
		}
	}()

	log.Println("Data collector started")
}

func (dc *DataCollector) collectTraffic(data model.TrafficData) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.trafficData[data.TunnelID] = data

	history := dc.trafficHistory[data.TunnelID]
	history = append(history, data)
	if len(history) > dc.maxHistorySize {
		history = history[len(history)-dc.maxHistorySize:]
	}
	dc.trafficHistory[data.TunnelID] = history

	log.Printf("Collected traffic data - Tunnel: %s, Vehicles: %d, Speed: %.1f km/h",
		data.TunnelID, data.VehicleCount, data.AvgSpeed)
}

func (dc *DataCollector) collectLightIntensity(data model.LightIntensityData) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.lightIntensityData[data.TunnelID] = data

	history := dc.lightHistory[data.TunnelID]
	history = append(history, data)
	if len(history) > dc.maxHistorySize {
		history = history[len(history)-dc.maxHistorySize:]
	}
	dc.lightHistory[data.TunnelID] = history

	log.Printf("Collected light intensity data - Tunnel: %s, Outside: %.1f lux, Inside: %.1f lux",
		data.TunnelID, data.OutsideLux, data.InsideLux)
}

func (dc *DataCollector) GetTrafficData(tunnelID string) (model.TrafficData, bool) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	data, ok := dc.trafficData[tunnelID]
	return data, ok
}

func (dc *DataCollector) GetLightIntensityData(tunnelID string) (model.LightIntensityData, bool) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	data, ok := dc.lightIntensityData[tunnelID]
	return data, ok
}

func (dc *DataCollector) GetAverageTraffic(tunnelID string, duration time.Duration) (int, float64) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	history, ok := dc.trafficHistory[tunnelID]
	if !ok || len(history) == 0 {
		return 0, 0
	}

	cutoff := time.Now().Add(-duration)
	var totalVehicles int
	var totalSpeed float64
	var count int

	for _, data := range history {
		if data.Timestamp.After(cutoff) {
			totalVehicles += data.VehicleCount
			totalSpeed += data.AvgSpeed
			count++
		}
	}

	if count == 0 {
		return 0, 0
	}

	return totalVehicles / count, totalSpeed / float64(count)
}

func (dc *DataCollector) GetAverageLightIntensity(tunnelID string, duration time.Duration) (float64, float64) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	history, ok := dc.lightHistory[tunnelID]
	if !ok || len(history) == 0 {
		return 0, 0
	}

	cutoff := time.Now().Add(-duration)
	var totalOutside, totalInside float64
	var count int

	for _, data := range history {
		if data.Timestamp.After(cutoff) {
			totalOutside += data.OutsideLux
			totalInside += data.InsideLux
			count++
		}
	}

	if count == 0 {
		return 0, 0
	}

	return totalOutside / float64(count), totalInside / float64(count)
}

func (dc *DataCollector) GetAllTunnels() []string {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	tunnels := make(map[string]bool)
	for id := range dc.trafficData {
		tunnels[id] = true
	}
	for id := range dc.lightIntensityData {
		tunnels[id] = true
	}

	result := make([]string, 0, len(tunnels))
	for id := range tunnels {
		result = append(result, id)
	}
	return result
}
