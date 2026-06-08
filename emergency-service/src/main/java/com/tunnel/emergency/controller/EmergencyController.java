package com.tunnel.emergency.controller;

import com.tunnel.emergency.entity.EmergencyAlert;
import com.tunnel.emergency.entity.EmergencyType;
import com.tunnel.emergency.entity.EscapeLight;
import com.tunnel.emergency.entity.LedMark;
import com.tunnel.emergency.service.EmergencyDispatchService;
import lombok.RequiredArgsConstructor;
import lombok.Data;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/emergency")
@RequiredArgsConstructor
public class EmergencyController {

    private final EmergencyDispatchService dispatchService;

    @PostMapping("/report")
    public ResponseEntity<EmergencyAlert> reportEmergency(@RequestBody EmergencyReportRequest request) {
        EmergencyAlert alert = dispatchService.reportEmergency(
                request.getTunnelId(),
                request.getType(),
                request.getLocation(),
                request.getDescription(),
                request.getSeverity()
        );
        return ResponseEntity.ok(alert);
    }

    @PostMapping("/clear/{tunnelId}")
    public ResponseEntity<EmergencyAlert> clearEmergency(@PathVariable String tunnelId) {
        EmergencyAlert alert = dispatchService.clearEmergency(tunnelId);
        if (alert == null) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(alert);
    }

    @GetMapping("/status/{tunnelId}")
    public ResponseEntity<Map<String, Object>> getEmergencyStatus(@PathVariable String tunnelId) {
        boolean hasEmergency = dispatchService.hasActiveEmergency(tunnelId);
        EmergencyAlert alert = dispatchService.getActiveEmergency(tunnelId);
        List<EscapeLight> escapeLights = dispatchService.getEscapeLights(tunnelId);
        List<LedMark> ledMarks = dispatchService.getLedMarks(tunnelId);

        return ResponseEntity.ok(Map.of(
                "hasEmergency", hasEmergency,
                "alert", alert != null ? alert : "none",
                "escapeLights", escapeLights,
                "ledMarks", ledMarks
        ));
    }

    @GetMapping("/escape-lights/{tunnelId}")
    public ResponseEntity<List<EscapeLight>> getEscapeLights(@PathVariable String tunnelId) {
        return ResponseEntity.ok(dispatchService.getEscapeLights(tunnelId));
    }

    @GetMapping("/led-marks/{tunnelId}")
    public ResponseEntity<List<LedMark>> getLedMarks(@PathVariable String tunnelId) {
        return ResponseEntity.ok(dispatchService.getLedMarks(tunnelId));
    }

    @Data
    public static class EmergencyReportRequest {
        private String tunnelId;
        private EmergencyType type;
        private String location;
        private String description;
        private Integer severity;
    }
}
