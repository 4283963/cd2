package model

import (
	"time"
)

type LightingMode int

const (
	ModeNormal LightingMode = iota
	ModeDay
	ModeNight
	ModeCloudy
	ModeEmergency
)

func (m LightingMode) String() string {
	switch m {
	case ModeDay:
		return "DAY"
	case ModeNight:
		return "NIGHT"
	case ModeCloudy:
		return "CLOUDY"
	case ModeEmergency:
		return "EMERGENCY"
	default:
		return "NORMAL"
	}
}

type LightingSegment struct {
	ID        string       `json:"id" gorm:"column:id;primaryKey"`
	TunnelID  string       `json:"tunnel_id" gorm:"column:tunnel_id;index"`
	Name      string       `json:"name" gorm:"column:name"`
	Brightness  float64    `json:"brightness" gorm:"column:brightness"`
	Mode      LightingMode `json:"mode" gorm:"column:mode"`
	Length    float64      `json:"length" gorm:"column:length"`
	UpdatedAt time.Time    `json:"updated_at" gorm:"column:updated_at"`
}

func (LightingSegment) TableName() string {
	return "lighting_segments"
}

type TrafficData struct {
	TunnelID     string    `json:"tunnel_id"`
	VehicleCount int       `json:"vehicle_count"`
	AvgSpeed     float64   `json:"avg_speed"`
	Timestamp    time.Time `json:"timestamp"`
}

type LightIntensityData struct {
	TunnelID    string    `json:"tunnel_id"`
	OutsideLux  float64   `json:"outside_lux"`
	InsideLux   float64   `json:"inside_lux"`
	Timestamp   time.Time `json:"timestamp"`
}

type EmergencyType int

const (
	EmergencyNone EmergencyType = iota
	EmergencyAccident
	EmergencyFire
	EmergencyFault
)

func (e EmergencyType) String() string {
	switch e {
	case EmergencyAccident:
		return "ACCIDENT"
	case EmergencyFire:
		return "FIRE"
	case EmergencyFault:
		return "FAULT"
	default:
		return "NONE"
	}
}

type DimmerConfig struct {
	Segments                []LightingSegment
	DayFactor               float64
	NightFactor             float64
	CloudyFactor            float64
	EmergencyFactor         float64
	TrafficDensityFactor    float64
	MinBrightness           float64
	MaxBrightness           float64
	EntranceSegmentID       string
	WeatherBrightnessBoost  float64
	SevereWeatherBoost      float64
}

type DeviceStatus struct {
	DeviceID    string    `json:"device_id"`
	DeviceType  string    `json:"device_type"`
	Status      string    `json:"status"`
	Brightness  float64   `json:"brightness,omitempty"`
	LastUpdate  time.Time `json:"last_update"`
}

type WeatherCondition int

const (
	WeatherClear WeatherCondition = iota
	WeatherCloudy
	WeatherLightRain
	WeatherHeavyRain
	WeatherStorm
	WeatherLightSnow
	WeatherHeavySnow
	WeatherBlizzard
)

func (w WeatherCondition) String() string {
	switch w {
	case WeatherCloudy:
		return "CLOUDY"
	case WeatherLightRain:
		return "LIGHT_RAIN"
	case WeatherHeavyRain:
		return "HEAVY_RAIN"
	case WeatherStorm:
		return "STORM"
	case WeatherLightSnow:
		return "LIGHT_SNOW"
	case WeatherHeavySnow:
		return "HEAVY_SNOW"
	case WeatherBlizzard:
		return "BLIZZARD"
	default:
		return "CLEAR"
	}
}

func WeatherConditionFromString(s string) WeatherCondition {
	switch s {
	case "CLOUDY":
		return WeatherCloudy
	case "LIGHT_RAIN":
		return WeatherLightRain
	case "HEAVY_RAIN", "RAIN":
		return WeatherHeavyRain
	case "STORM":
		return WeatherStorm
	case "LIGHT_SNOW":
		return WeatherLightSnow
	case "HEAVY_SNOW", "SNOW":
		return WeatherHeavySnow
	case "BLIZZARD":
		return WeatherBlizzard
	default:
		return WeatherClear
	}
}

func (w WeatherCondition) IsSevere() bool {
	return w >= WeatherHeavyRain
}

type WeatherAlert struct {
	TunnelID    string           `json:"tunnel_id"`
	Condition   WeatherCondition `json:"condition"`
	Severity    int32            `json:"severity"`
	Description string           `json:"description"`
	ExpectedDuration int64       `json:"expected_duration"`
	Timestamp   time.Time        `json:"timestamp"`
}
