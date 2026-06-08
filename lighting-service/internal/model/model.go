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
}

type DeviceStatus struct {
	DeviceID    string    `json:"device_id"`
	DeviceType  string    `json:"device_type"`
	Status      string    `json:"status"`
	Brightness  float64   `json:"brightness,omitempty"`
	LastUpdate  time.Time `json:"last_update"`
}
