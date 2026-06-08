package grpcserver

import (
	"context"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"lighting-service/internal/model"
)

type LightingSegmentProto struct {
	SegmentId  string
	Brightness float64
	Mode       int32
	UpdatedAt  int64
}

type LightingStatusResponseProto struct {
	Segments    []*LightingSegmentProto
	CurrentMode int32
	Timestamp   int64
}

type LightingControlResponseProto struct {
	Success  bool
	Message  string
	Segments []*LightingSegmentProto
}

type DimmerServiceInterface interface {
	GetSegments(tunnelID string) ([]*model.LightingSegment, bool)
	GetCurrentMode(tunnelID string) (model.LightingMode, bool)
	IsEmergencyMode(tunnelID string) bool
	SetEmergencyMode(tunnelID string, emergencyType model.EmergencyType)
	ClearEmergencyMode(tunnelID string)
	CalculateBrightness(tunnelID string) ([]*model.LightingSegment, error)
	InitializeSegments(tunnelID string)
}

type StorageInterface interface {
	CacheLightingSegments(tunnelID string, segments []*model.LightingSegment) error
	GetLightingSegments(tunnelID string) ([]*model.LightingSegment, error)
	SetEmergencyMode(tunnelID string, emergencyType model.EmergencyType) error
	ClearEmergencyMode(tunnelID string) error
	SaveLightingSegment(segment *model.LightingSegment) error
}

type LightingServer struct {
	dimmer  DimmerServiceInterface
	storage StorageInterface
	server  *grpc.Server
}

func NewLightingServer(dimmer DimmerServiceInterface, storage StorageInterface) *LightingServer {
	return &LightingServer{
		dimmer:  dimmer,
		storage: storage,
	}
}

func (s *LightingServer) Start(port string) error {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}

	s.server = grpc.NewServer(
		grpc.UnaryInterceptor(s.unaryInterceptor),
	)

	log.Printf("gRPC server started on %s", port)
	return s.server.Serve(lis)
}

func (s *LightingServer) unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	log.Printf("gRPC request - method: %s", info.FullMethod)

	resp, err := handler(ctx, req)

	duration := time.Since(start)
	if err != nil {
		log.Printf("gRPC request failed - method: %s, duration: %v, error: %v", info.FullMethod, duration, err)
	} else {
		log.Printf("gRPC request completed - method: %s, duration: %v", info.FullMethod, duration)
	}

	return resp, err
}

func (s *LightingServer) GetLightingStatus(ctx context.Context, tunnelID string) (*LightingStatusResponseProto, error) {
	if tunnelID == "" {
		return nil, status.Error(codes.InvalidArgument, "tunnel_id is required")
	}

	segments, ok := s.dimmer.GetSegments(tunnelID)
	if !ok {
		s.dimmer.InitializeSegments(tunnelID)
		segments, _ = s.dimmer.GetSegments(tunnelID)
	}

	mode, _ := s.dimmer.GetCurrentMode(tunnelID)

	pbSegments := make([]*LightingSegmentProto, len(segments))
	for i, seg := range segments {
		pbSegments[i] = &LightingSegmentProto{
			SegmentId:  seg.ID,
			Brightness: seg.Brightness,
			Mode:       int32(seg.Mode),
			UpdatedAt:  seg.UpdatedAt.Unix(),
		}
	}

	return &LightingStatusResponseProto{
		Segments:    pbSegments,
		CurrentMode: int32(mode),
		Timestamp:   time.Now().Unix(),
	}, nil
}

func (s *LightingServer) SetEmergencyLighting(ctx context.Context, tunnelID string, emergencyType model.EmergencyType, emergencySegment string) (*LightingControlResponseProto, error) {
	if tunnelID == "" {
		return nil, status.Error(codes.InvalidArgument, "tunnel_id is required")
	}

	s.dimmer.SetEmergencyMode(tunnelID, emergencyType)

	if err := s.storage.SetEmergencyMode(tunnelID, emergencyType); err != nil {
		log.Printf("Failed to cache emergency mode: %v", err)
	}

	segments, err := s.dimmer.CalculateBrightness(tunnelID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "calculate brightness failed: %v", err)
	}

	if err := s.storage.CacheLightingSegments(tunnelID, segments); err != nil {
		log.Printf("Failed to cache lighting segments: %v", err)
	}

	for _, seg := range segments {
		if err := s.storage.SaveLightingSegment(seg); err != nil {
			log.Printf("Failed to save lighting segment %s: %v", seg.ID, err)
		}
	}

	pbSegments := make([]*LightingSegmentProto, len(segments))
	for i, seg := range segments {
		pbSegments[i] = &LightingSegmentProto{
			SegmentId:  seg.ID,
			Brightness: seg.Brightness,
			Mode:       int32(seg.Mode),
			UpdatedAt:  seg.UpdatedAt.Unix(),
		}
	}

	return &LightingControlResponseProto{
		Success:  true,
		Message:  "Emergency lighting activated",
		Segments: pbSegments,
	}, nil
}

func (s *LightingServer) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}
