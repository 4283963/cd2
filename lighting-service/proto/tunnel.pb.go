package tunnel

import (
	"fmt"
)

type LightingMode int32

const (
	LightingMode_MODE_NORMAL    LightingMode = 0
	LightingMode_MODE_DAY       LightingMode = 1
	LightingMode_MODE_NIGHT     LightingMode = 2
	LightingMode_MODE_CLOUDY    LightingMode = 3
	LightingMode_MODE_EMERGENCY LightingMode = 4
)

type EmergencyType int32

const (
	EmergencyType_EMERGENCY_NONE     EmergencyType = 0
	EmergencyType_EMERGENCY_ACCIDENT EmergencyType = 1
	EmergencyType_EMERGENCY_FIRE     EmergencyType = 2
	EmergencyType_EMERGENCY_FAULT    EmergencyType = 3
)

type EscapeLightStatus int32

const (
	EscapeLightStatus_ESCAPE_OFF   EscapeLightStatus = 0
	EscapeLightStatus_ESCAPE_ON    EscapeLightStatus = 1
	EscapeLightStatus_ESCAPE_BLINK EscapeLightStatus = 2
)

type LightingSegment struct {
	SegmentId  string
	Brightness float64
	Mode       LightingMode
	UpdatedAt  int64
}

type TrafficData struct {
	TunnelId     string
	VehicleCount int32
	AvgSpeed     float64
	Timestamp    int64
}

type LightIntensityData struct {
	TunnelId   string
	OutsideLux float64
	InsideLux  float64
	Timestamp  int64
}

type LightingControlRequest struct {
	TunnelId         string
	EmergencyType    EmergencyType
	EmergencySegment string
	Timestamp        int64
}

type LightingControlResponse struct {
	Success  bool
	Message  string
	Segments []*LightingSegment
}

type LightingStatusRequest struct {
	TunnelId string
}

type LightingStatusResponse struct {
	Segments    []*LightingSegment
	CurrentMode LightingMode
	Timestamp   int64
}

type EmergencyAlert struct {
	EmergencyId string
	TunnelId    string
	Type        EmergencyType
	Location    string
	Description string
	Timestamp   int64
	Severity    int32
}

type EscapeLightControl struct {
	TunnelId  string
	SegmentId string
	Status    EscapeLightStatus
	Direction string
	Timestamp int64
}

type EmergencyStatusRequest struct {
	TunnelId string
}

type EmergencyStatusResponse struct {
	HasEmergency bool
	CurrentAlert *EmergencyAlert
	EscapeLights []*EscapeLightControl
}

var LightingMode_name = map[int32]string{
	0: "MODE_NORMAL",
	1: "MODE_DAY",
	2: "MODE_NIGHT",
	3: "MODE_CLOUDY",
	4: "MODE_EMERGENCY",
}

var LightingMode_value = map[string]int32{
	"MODE_NORMAL":    0,
	"MODE_DAY":       1,
	"MODE_NIGHT":     2,
	"MODE_CLOUDY":    3,
	"MODE_EMERGENCY": 4,
}

func (x LightingMode) String() string {
	return LightingMode_name[int32(x)]
}

var EmergencyType_name = map[int32]string{
	0: "EMERGENCY_NONE",
	1: "EMERGENCY_ACCIDENT",
	2: "EMERGENCY_FIRE",
	3: "EMERGENCY_FAULT",
}

var EmergencyType_value = map[string]int32{
	"EMERGENCY_NONE":     0,
	"EMERGENCY_ACCIDENT": 1,
	"EMERGENCY_FIRE":     2,
	"EMERGENCY_FAULT":    3,
}

func (x EmergencyType) String() string {
	return EmergencyType_name[int32(x)]
}

var EscapeLightStatus_name = map[int32]string{
	0: "ESCAPE_OFF",
	1: "ESCAPE_ON",
	2: "ESCAPE_BLINK",
}

var EscapeLightStatus_value = map[string]int32{
	"ESCAPE_OFF":   0,
	"ESCAPE_ON":    1,
	"ESCAPE_BLINK": 2,
}

func (x EscapeLightStatus) String() string {
	return EscapeLightStatus_name[int32(x)]
}

func (m *LightingSegment) GetSegmentId() string {
	if m != nil {
		return m.SegmentId
	}
	return ""
}

func (m *LightingSegment) GetBrightness() float64 {
	if m != nil {
		return m.Brightness
	}
	return 0
}

func (m *LightingSegment) GetMode() LightingMode {
	if m != nil {
		return m.Mode
	}
	return LightingMode_MODE_NORMAL
}

func (m *LightingSegment) GetUpdatedAt() int64 {
	if m != nil {
		return m.UpdatedAt
	}
	return 0
}

func (m *EmergencyAlert) GetEmergencyId() string {
	if m != nil {
		return m.EmergencyId
	}
	return ""
}

func (m *EmergencyAlert) GetTunnelId() string {
	if m != nil {
		return m.TunnelId
	}
	return ""
}

func (m *EmergencyAlert) GetType() EmergencyType {
	if m != nil {
		return m.Type
	}
	return EmergencyType_EMERGENCY_NONE
}

func (m *EmergencyAlert) GetLocation() string {
	if m != nil {
		return m.Location
	}
	return ""
}

func (m *EmergencyAlert) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *EmergencyAlert) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *EmergencyAlert) GetSeverity() int32 {
	if m != nil {
		return m.Severity
	}
	return 0
}

type LightingServiceServer interface {
	GetLightingStatus(interface { /* context.Context */
	}, *LightingStatusRequest) (*LightingStatusResponse, error)
	SetEmergencyLighting(interface { /* context.Context */
	}, *LightingControlRequest) (*LightingControlResponse, error)
}

type LightingServiceClient interface {
	GetLightingStatus(interface { /* context.Context */
	}, *LightingStatusRequest) (*LightingStatusResponse, error)
	SetEmergencyLighting(interface { /* context.Context */
	}, *LightingControlRequest) (*LightingControlResponse, error)
}

type UnimplementedLightingServiceServer struct{}

func (UnimplementedLightingServiceServer) GetLightingStatus(interface{}, *LightingStatusRequest) (*LightingStatusResponse, error) {
	return nil, fmt.Errorf("method not implemented")
}

func (UnimplementedLightingServiceServer) SetEmergencyLighting(interface{}, *LightingControlRequest) (*LightingControlResponse, error) {
	return nil, fmt.Errorf("method not implemented")
}

type EmergencyServiceServer interface {
	ReportEmergency(interface { /* context.Context */
	}, *EmergencyAlert) (*EmergencyStatusResponse, error)
	GetEmergencyStatus(interface { /* context.Context */
	}, *EmergencyStatusRequest) (*EmergencyStatusResponse, error)
	ClearEmergency(interface { /* context.Context */
	}, *EmergencyStatusRequest) (*EmergencyStatusResponse, error)
}

type UnimplementedEmergencyServiceServer struct{}

func (UnimplementedEmergencyServiceServer) ReportEmergency(interface{}, *EmergencyAlert) (*EmergencyStatusResponse, error) {
	return nil, fmt.Errorf("method not implemented")
}

func (UnimplementedEmergencyServiceServer) GetEmergencyStatus(interface{}, *EmergencyStatusRequest) (*EmergencyStatusResponse, error) {
	return nil, fmt.Errorf("method not implemented")
}

func (UnimplementedEmergencyServiceServer) ClearEmergency(interface{}, *EmergencyStatusRequest) (*EmergencyStatusResponse, error) {
	return nil, fmt.Errorf("method not implemented")
}

func RegisterLightingServiceServer(s *grpcServerMock, srv LightingServiceServer) {
}

func RegisterEmergencyServiceServer(s *grpcServerMock, srv EmergencyServiceServer) {
}

type grpcServerMock struct{}
