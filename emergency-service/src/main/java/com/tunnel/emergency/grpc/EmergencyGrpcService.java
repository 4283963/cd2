package com.tunnel.emergency.grpc;

import com.tunnel.emergency.entity.EmergencyAlert;
import com.tunnel.emergency.entity.EmergencyType;
import com.tunnel.emergency.entity.EscapeLight;
import com.tunnel.emergency.entity.EscapeLightStatus;
import com.tunnel.emergency.service.EmergencyDispatchService;
import io.grpc.stub.StreamObserver;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.server.service.GrpcService;

import java.util.List;

@Slf4j
@GrpcService
@RequiredArgsConstructor
public class EmergencyGrpcService extends com.tunnel.proto.EmergencyServiceGrpc.EmergencyServiceImplBase {

    private final EmergencyDispatchService dispatchService;
    private final LightingGrpcClient lightingGrpcClient;

    @Override
    public void reportEmergency(com.tunnel.proto.EmergencyAlert request,
                                StreamObserver<com.tunnel.proto.EmergencyStatusResponse> responseObserver) {
        log.info("Received emergency report via gRPC - Tunnel: {}, Type: {}",
                request.getTunnelId(), request.getType());

        try {
            EmergencyType emergencyType = convertEmergencyType(request.getType());

            EmergencyAlert alert = dispatchService.reportEmergency(
                    request.getTunnelId(),
                    emergencyType,
                    request.getLocation(),
                    request.getDescription(),
                    request.getSeverity()
            );

            if (lightingGrpcClient.isAvailable()) {
                lightingGrpcClient.setEmergencyLighting(
                        request.getTunnelId(),
                        request.getType(),
                        request.getLocation()
                );
            }

            com.tunnel.proto.EmergencyStatusResponse response = buildStatusResponse(request.getTunnelId(), alert);
            responseObserver.onNext(response);
            responseObserver.onCompleted();

            log.info("Emergency processed via gRPC - ID: {}", alert.getEmergencyId());
        } catch (Exception e) {
            log.error("Error processing emergency via gRPC: {}", e.getMessage(), e);
            responseObserver.onError(e);
        }
    }

    @Override
    public void getEmergencyStatus(com.tunnel.proto.EmergencyStatusRequest request,
                                   StreamObserver<com.tunnel.proto.EmergencyStatusResponse> responseObserver) {
        try {
            EmergencyAlert alert = dispatchService.getActiveEmergency(request.getTunnelId());
            boolean hasEmergency = alert != null;

            com.tunnel.proto.EmergencyStatusResponse response = buildStatusResponse(request.getTunnelId(), alert);
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("Error getting emergency status via gRPC: {}", e.getMessage(), e);
            responseObserver.onError(e);
        }
    }

    @Override
    public void clearEmergency(com.tunnel.proto.EmergencyStatusRequest request,
                               StreamObserver<com.tunnel.proto.EmergencyStatusResponse> responseObserver) {
        log.info("Received clear emergency request via gRPC - Tunnel: {}", request.getTunnelId());

        try {
            EmergencyAlert alert = dispatchService.clearEmergency(request.getTunnelId());

            com.tunnel.proto.EmergencyStatusResponse response = buildStatusResponse(request.getTunnelId(), alert);
            responseObserver.onNext(response);
            responseObserver.onCompleted();

            log.info("Emergency cleared via gRPC - Tunnel: {}", request.getTunnelId());
        } catch (Exception e) {
            log.error("Error clearing emergency via gRPC: {}", e.getMessage(), e);
            responseObserver.onError(e);
        }
    }

    private com.tunnel.proto.EmergencyStatusResponse buildStatusResponse(String tunnelId, EmergencyAlert alert) {
        com.tunnel.proto.EmergencyStatusResponse.Builder responseBuilder =
                com.tunnel.proto.EmergencyStatusResponse.newBuilder();

        boolean hasEmergency = alert != null && alert.getActive();
        responseBuilder.setHasEmergency(hasEmergency);

        if (hasEmergency && alert != null) {
            com.tunnel.proto.EmergencyAlert.Builder alertBuilder = com.tunnel.proto.EmergencyAlert.newBuilder()
                    .setEmergencyId(alert.getEmergencyId())
                    .setTunnelId(alert.getTunnelId())
                    .setType(convertProtoEmergencyType(alert.getType()))
                    .setLocation(alert.getLocation() != null ? alert.getLocation() : "")
                    .setDescription(alert.getDescription() != null ? alert.getDescription() : "")
                    .setTimestamp(alert.getCreatedAt() != null ?
                            java.sql.Timestamp.valueOf(alert.getCreatedAt()).getTime() : 0)
                    .setSeverity(alert.getSeverity() != null ? alert.getSeverity() : 0);

            responseBuilder.setCurrentAlert(alertBuilder.build());
        }

        List<EscapeLight> escapeLights = dispatchService.getEscapeLights(tunnelId);
        for (EscapeLight light : escapeLights) {
            com.tunnel.proto.EscapeLightControl escapeControl = com.tunnel.proto.EscapeLightControl.newBuilder()
                    .setTunnelId(tunnelId)
                    .setSegmentId(light.getSegmentId())
                    .setStatus(convertEscapeLightStatus(light.getStatus()))
                    .setDirection(light.getDirection() != null ? light.getDirection() : "")
                    .setTimestamp(light.getUpdatedAt() != null ?
                            java.sql.Timestamp.valueOf(light.getUpdatedAt()).getTime() : 0)
                    .build();
            responseBuilder.addEscapeLights(escapeControl);
        }

        return responseBuilder.build();
    }

    private EmergencyType convertEmergencyType(com.tunnel.proto.EmergencyType type) {
        return switch (type) {
            case EMERGENCY_ACCIDENT -> EmergencyType.ACCIDENT;
            case EMERGENCY_FIRE -> EmergencyType.FIRE;
            case EMERGENCY_FAULT -> EmergencyType.FAULT;
            default -> EmergencyType.NONE;
        };
    }

    private com.tunnel.proto.EmergencyType convertProtoEmergencyType(EmergencyType type) {
        return switch (type) {
            case ACCIDENT -> com.tunnel.proto.EmergencyType.EMERGENCY_ACCIDENT;
            case FIRE -> com.tunnel.proto.EmergencyType.EMERGENCY_FIRE;
            case FAULT -> com.tunnel.proto.EmergencyType.EMERGENCY_FAULT;
            default -> com.tunnel.proto.EmergencyType.EMERGENCY_NONE;
        };
    }

    private com.tunnel.proto.EscapeLightStatus convertEscapeLightStatus(EscapeLightStatus status) {
        return switch (status) {
            case ON -> com.tunnel.proto.EscapeLightStatus.ESCAPE_ON;
            case BLINK -> com.tunnel.proto.EscapeLightStatus.ESCAPE_BLINK;
            default -> com.tunnel.proto.EscapeLightStatus.ESCAPE_OFF;
        };
    }
}
