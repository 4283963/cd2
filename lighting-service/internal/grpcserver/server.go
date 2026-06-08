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
	pb "lighting-service/proto"
)

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
	pb.UnimplementedLightingServiceServer
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
		grpc.MaxRecvMsgSize(10*1024*1024),
		grpc.MaxSendMsgSize(10*1024*1024),
	)

	pb.RegisterLightingServiceServer(s.server, s)

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

func (s *LightingServer) GetLightingStatus(ctx context.Context, req *pb.LightingStatusRequest) (*pb.LightingStatusResponse, error) {
	tunnelID := req.GetTunnelId()
	if tunnelID == "" {
		return nil, status.Error(codes.InvalidArgument, "tunnel_id is required")
	}

	segments, ok := s.dimmer.GetSegments(tunnelID)
	if !ok {
		s.dimmer.InitializeSegments(tunnelID)
		segments, _ = s.dimmer.GetSegments(tunnelID)
	}

	mode, _ := s.dimmer.GetCurrentMode(tunnelID)

	pbSegments := make([]*pb.LightingSegment, len(segments))
	for i, seg := range segments {
		pbSegments[i] = toProtoLightingSegment(seg)
	}

	return &pb.LightingStatusResponse{
		Segments:    pbSegments,
		CurrentMode: toProtoLightingMode(mode),
		Timestamp:   time.Now().Unix(),
	}, nil
}

func (s *LightingServer) SetEmergencyLighting(ctx context.Context, req *pb.LightingControlRequest) (*pb.LightingControlResponse, error) {
	tunnelID := req.GetTunnelId()
	if tunnelID == "" {
		return nil, status.Error(codes.InvalidArgument, "tunnel_id is required")
	}

	emergencyType := fromProtoEmergencyType(req.GetEmergencyType())

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

	pbSegments := make([]*pb.LightingSegment, len(segments))
	for i, seg := range segments {
		pbSegments[i] = toProtoLightingSegment(seg)
	}

	return &pb.LightingControlResponse{
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

func toProtoLightingSegment(seg *model.LightingSegment) *pb.LightingSegment {
	return &pb.LightingSegment{
		SegmentId:  seg.ID,
		Brightness: seg.Brightness,
		Mode:       toProtoLightingMode(seg.Mode),
		UpdatedAt:  seg.UpdatedAt.Unix(),
	}
}

func toProtoLightingMode(mode model.LightingMode) pb.LightingMode {
	switch mode {
	case model.ModeDay:
		return pb.LightingMode_MODE_DAY
	case model.ModeNight:
		return pb.LightingMode_MODE_NIGHT
	case model.ModeCloudy:
		return pb.LightingMode_MODE_CLOUDY
	case model.ModeEmergency:
		return pb.LightingMode_MODE_EMERGENCY
	default:
		return pb.LightingMode_MODE_NORMAL
	}
}

func fromProtoEmergencyType(et pb.EmergencyType) model.EmergencyType {
	switch et {
	case pb.EmergencyType_EMERGENCY_ACCIDENT:
		return model.EmergencyAccident
	case pb.EmergencyType_EMERGENCY_FIRE:
		return model.EmergencyFire
	case pb.EmergencyType_EMERGENCY_FAULT:
		return model.EmergencyFault
	default:
		return model.EmergencyNone
	}
}
