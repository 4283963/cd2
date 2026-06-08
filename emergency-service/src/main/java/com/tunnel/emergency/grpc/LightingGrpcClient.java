package com.tunnel.emergency.grpc;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;
import java.util.concurrent.TimeUnit;

@Slf4j
@Component
public class LightingGrpcClient {

    @Value("${grpc.client.lighting-service.address:localhost:50051}")
    private String lightingServiceAddress;

    private ManagedChannel channel;
    private com.tunnel.proto.LightingServiceGrpc.LightingServiceBlockingStub blockingStub;
    private com.tunnel.proto.LightingServiceGrpc.LightingServiceStub asyncStub;

    @PostConstruct
    public void init() {
        try {
            channel = ManagedChannelBuilder.forTarget(lightingServiceAddress)
                    .usePlaintext()
                    .keepAliveWithoutCalls(true)
                    .build();

            blockingStub = com.tunnel.proto.LightingServiceGrpc.newBlockingStub(channel);
            asyncStub = com.tunnel.proto.LightingServiceGrpc.newStub(channel);

            log.info("gRPC client initialized for lighting service at {}", lightingServiceAddress);
        } catch (Exception e) {
            log.warn("Failed to initialize gRPC client: {}", e.getMessage());
        }
    }

    public com.tunnel.proto.LightingStatusResponse getLightingStatus(String tunnelId) {
        if (blockingStub == null) {
            log.warn("gRPC client not initialized");
            return null;
        }

        try {
            com.tunnel.proto.LightingStatusRequest request = com.tunnel.proto.LightingStatusRequest.newBuilder()
                    .setTunnelId(tunnelId)
                    .build();

            return blockingStub.getLightingStatus(request);
        } catch (Exception e) {
            log.error("Failed to get lighting status: {}", e.getMessage());
            return null;
        }
    }

    public com.tunnel.proto.LightingControlResponse setEmergencyLighting(String tunnelId,
                                                                          com.tunnel.proto.EmergencyType emergencyType,
                                                                          String emergencySegment) {
        if (blockingStub == null) {
            log.warn("gRPC client not initialized");
            return null;
        }

        try {
            com.tunnel.proto.LightingControlRequest request = com.tunnel.proto.LightingControlRequest.newBuilder()
                    .setTunnelId(tunnelId)
                    .setEmergencyType(emergencyType)
                    .setEmergencySegment(emergencySegment != null ? emergencySegment : "")
                    .setTimestamp(System.currentTimeMillis())
                    .build();

            return blockingStub.setEmergencyLighting(request);
        } catch (Exception e) {
            log.error("Failed to set emergency lighting: {}", e.getMessage());
            return null;
        }
    }

    @PreDestroy
    public void shutdown() {
        if (channel != null) {
            try {
                channel.shutdown().awaitTermination(5, TimeUnit.SECONDS);
                log.info("gRPC client shutdown complete");
            } catch (InterruptedException e) {
                log.warn("gRPC client shutdown interrupted");
            }
        }
    }

    public boolean isAvailable() {
        return channel != null && !channel.isShutdown();
    }
}
