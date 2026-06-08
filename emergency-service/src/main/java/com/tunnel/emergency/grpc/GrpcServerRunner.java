package com.tunnel.emergency.grpc;

import io.grpc.Server;
import io.grpc.ServerBuilder;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.CommandLineRunner;
import org.springframework.stereotype.Component;

import jakarta.annotation.PreDestroy;
import java.io.IOException;

@Slf4j
@Component
@RequiredArgsConstructor
public class GrpcServerRunner implements CommandLineRunner {

    private final EmergencyGrpcService emergencyGrpcService;

    @Value("${grpc.server.port:50052}")
    private int grpcPort;

    private Server server;

    @Override
    public void run(String... args) throws Exception {
        startServer();
    }

    private void startServer() throws IOException {
        server = ServerBuilder.forPort(grpcPort)
                .addService(emergencyGrpcService)
                .build()
                .start();

        log.info("gRPC server started on port {}", grpcPort);

        Runtime.getRuntime().addShutdownHook(new Thread(() -> {
            log.info("Shutting down gRPC server...");
            GrpcServerRunner.this.stopServer();
            log.info("gRPC server shut down");
        }));
    }

    private void stopServer() {
        if (server != null) {
            server.shutdown();
        }
    }

    @PreDestroy
    public void preDestroy() {
        stopServer();
    }
}
