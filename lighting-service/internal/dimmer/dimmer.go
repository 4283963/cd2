package dimmer

import (
	"log"
	"math"
	"sync"
	"time"

	"lighting-service/internal/model"
)

type DimmerService struct {
	mu             sync.RWMutex
	config         *model.DimmerConfig
	segments       map[string][]*model.LightingSegment
	currentMode    map[string]model.LightingMode
	emergencyMode  map[string]bool
	weatherCondition map[string]model.WeatherCondition
	collector      DataCollectorInterface
}

type DataCollectorInterface interface {
	GetAverageTraffic(tunnelID string, duration time.Duration) (int, float64)
	GetAverageLightIntensity(tunnelID string, duration time.Duration) (float64, float64)
}

func NewDimmerService(config *model.DimmerConfig, collector DataCollectorInterface) *DimmerService {
	ds := &DimmerService{
		config:           config,
		segments:         make(map[string][]*model.LightingSegment),
		currentMode:      make(map[string]model.LightingMode),
		emergencyMode:    make(map[string]bool),
		weatherCondition: make(map[string]model.WeatherCondition),
		collector:        collector,
	}
	return ds
}

func (ds *DimmerService) InitializeSegments(tunnelID string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	segments := make([]*model.LightingSegment, len(ds.config.Segments))
	for i, seg := range ds.config.Segments {
		segments[i] = &model.LightingSegment{
			ID:         seg.ID,
			TunnelID:   tunnelID,
			Name:       seg.Name,
			Brightness: seg.Brightness,
			Mode:       model.ModeNormal,
			Length:     seg.Length,
			UpdatedAt:  time.Now(),
		}
	}

	ds.segments[tunnelID] = segments
	ds.currentMode[tunnelID] = model.ModeNormal
	ds.emergencyMode[tunnelID] = false
	ds.weatherCondition[tunnelID] = model.WeatherClear

	log.Printf("Initialized %d lighting segments for tunnel %s", len(segments), tunnelID)
}

func (ds *DimmerService) CalculateBrightness(tunnelID string) ([]*model.LightingSegment, error) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	segments, ok := ds.segments[tunnelID]
	if !ok {
		return nil, nil
	}

	if ds.emergencyMode[tunnelID] {
		return ds.calculateEmergencyBrightness(tunnelID, segments)
	}

	avgOutsideLux, _ := ds.collector.GetAverageLightIntensity(tunnelID, 5*time.Minute)
	avgVehicles, _ := ds.collector.GetAverageTraffic(tunnelID, 5*time.Minute)

	mode := ds.determineMode(avgOutsideLux)
	ds.currentMode[tunnelID] = mode

	factor := ds.getModeFactor(mode)
	trafficFactor := 1.0 + float64(avgVehicles)*ds.config.TrafficDensityFactor/100.0

	weatherCondition := ds.weatherCondition[tunnelID]
	weatherBoost := ds.getWeatherBoost(weatherCondition)

	for _, seg := range segments {
		baseBrightness := ds.getSegmentBaseBrightness(seg.ID)
		brightness := baseBrightness * factor * trafficFactor

		if ds.isEntranceSegment(seg.ID) && weatherBoost > 1.0 {
			brightness = baseBrightness * factor * weatherBoost * trafficFactor
		}

		brightness = math.Min(brightness, ds.config.MaxBrightness)
		brightness = math.Max(brightness, ds.config.MinBrightness)

		seg.Brightness = math.Round(brightness*10) / 10
		seg.Mode = mode
		seg.UpdatedAt = time.Now()
	}

	if weatherBoost > 1.0 {
		log.Printf("Calculated brightness for tunnel %s (mode: %s, weather: %s, boost: %.1fx)",
			tunnelID, mode.String(), weatherCondition.String(), weatherBoost)
	} else {
		log.Printf("Calculated brightness for tunnel %s (mode: %s)", tunnelID, mode.String())
	}

	return segments, nil
}

func (ds *DimmerService) calculateEmergencyBrightness(tunnelID string, segments []*model.LightingSegment) ([]*model.LightingSegment, error) {
	for _, seg := range segments {
		baseBrightness := ds.getSegmentBaseBrightness(seg.ID)
		brightness := baseBrightness * ds.config.EmergencyFactor

		brightness = math.Min(brightness, ds.config.MaxBrightness)

		seg.Brightness = math.Round(brightness*10) / 10
		seg.Mode = model.ModeEmergency
		seg.UpdatedAt = time.Now()
	}

	log.Printf("Emergency brightness calculated for tunnel %s", tunnelID)
	return segments, nil
}

func (ds *DimmerService) determineMode(outsideLux float64) model.LightingMode {
	switch {
	case outsideLux >= 3000:
		return model.ModeDay
	case outsideLux >= 1000:
		return model.ModeCloudy
	case outsideLux > 0:
		return model.ModeNight
	default:
		return model.ModeNormal
	}
}

func (ds *DimmerService) getModeFactor(mode model.LightingMode) float64 {
	switch mode {
	case model.ModeDay:
		return ds.config.DayFactor
	case model.ModeNight:
		return ds.config.NightFactor
	case model.ModeCloudy:
		return ds.config.CloudyFactor
	case model.ModeEmergency:
		return ds.config.EmergencyFactor
	default:
		return 1.0
	}
}

func (ds *DimmerService) getSegmentBaseBrightness(segmentID string) float64 {
	for _, seg := range ds.config.Segments {
		if seg.ID == segmentID {
			return seg.Brightness
		}
	}
	return ds.config.MinBrightness
}

func (ds *DimmerService) getWeatherBoost(condition model.WeatherCondition) float64 {
	switch condition {
	case model.WeatherLightRain, model.WeatherLightSnow:
		return ds.config.WeatherBrightnessBoost
	case model.WeatherHeavyRain, model.WeatherHeavySnow:
		return ds.config.SevereWeatherBoost
	case model.WeatherStorm, model.WeatherBlizzard:
		return ds.config.SevereWeatherBoost * 1.2
	default:
		return 1.0
	}
}

func (ds *DimmerService) isEntranceSegment(segmentID string) bool {
	return segmentID == ds.config.EntranceSegmentID
}

func (ds *DimmerService) SetWeatherCondition(tunnelID string, condition model.WeatherCondition) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	oldCondition := ds.weatherCondition[tunnelID]
	ds.weatherCondition[tunnelID] = condition

	if oldCondition != condition {
		if condition.IsSevere() {
			log.Printf("Severe weather alert for tunnel %s: %s - entrance brightness boosted",
				tunnelID, condition.String())
		} else if oldCondition.IsSevere() {
			log.Printf("Weather cleared for tunnel %s: %s - entrance brightness restored",
				tunnelID, condition.String())
		} else {
			log.Printf("Weather updated for tunnel %s: %s", tunnelID, condition.String())
		}
	}
}

func (ds *DimmerService) GetWeatherCondition(tunnelID string) model.WeatherCondition {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.weatherCondition[tunnelID]
}

func (ds *DimmerService) SetEmergencyMode(tunnelID string, emergencyType model.EmergencyType) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.emergencyMode[tunnelID] = true
	log.Printf("Emergency mode activated for tunnel %s, type: %s", tunnelID, emergencyType.String())
}

func (ds *DimmerService) ClearEmergencyMode(tunnelID string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.emergencyMode[tunnelID] = false
	log.Printf("Emergency mode cleared for tunnel %s", tunnelID)
}

func (ds *DimmerService) GetSegments(tunnelID string) ([]*model.LightingSegment, bool) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	segments, ok := ds.segments[tunnelID]
	if !ok {
		return nil, false
	}

	result := make([]*model.LightingSegment, len(segments))
	for i, seg := range segments {
		s := *seg
		result[i] = &s
	}
	return result, true
}

func (ds *DimmerService) GetCurrentMode(tunnelID string) (model.LightingMode, bool) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	mode, ok := ds.currentMode[tunnelID]
	return mode, ok
}

func (ds *DimmerService) IsEmergencyMode(tunnelID string) bool {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.emergencyMode[tunnelID]
}
