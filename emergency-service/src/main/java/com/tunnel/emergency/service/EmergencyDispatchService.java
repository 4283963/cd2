package com.tunnel.emergency.service;

import com.tunnel.emergency.config.EmergencyConfig;
import com.tunnel.emergency.entity.*;
import com.tunnel.emergency.repository.EmergencyAlertRepository;
import com.tunnel.emergency.repository.EscapeLightRepository;
import com.tunnel.emergency.repository.LedMarkRepository;
import jakarta.annotation.PostConstruct;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;

@Slf4j
@Service
@RequiredArgsConstructor
public class EmergencyDispatchService {

    private final EmergencyAlertRepository alertRepository;
    private final EscapeLightRepository escapeLightRepository;
    private final LedMarkRepository ledMarkRepository;
    private final EmergencyConfig emergencyConfig;
    private final RedisTemplate<String, Object> redisTemplate;

    private final Map<String, EmergencyAlert> activeEmergencies = new ConcurrentHashMap<>();

    @PostConstruct
    public void init() {
        List<EmergencyAlert> activeAlerts = alertRepository.findByActiveTrueOrderByCreatedAtDesc();
        for (EmergencyAlert alert : activeAlerts) {
            activeEmergencies.put(alert.getTunnelId(), alert);
        }
        log.info("Loaded {} active emergencies from database", activeAlerts.size());
    }

    @Transactional
    public EmergencyAlert reportEmergency(String tunnelId, EmergencyType type, String location,
                                          String description, Integer severity) {
        String emergencyId = "EMG-" + UUID.randomUUID().toString().substring(0, 8).toUpperCase();

        EmergencyAlert alert = EmergencyAlert.builder()
                .emergencyId(emergencyId)
                .tunnelId(tunnelId)
                .type(type)
                .location(location)
                .description(description)
                .severity(severity != null ? severity : 3)
                .active(true)
                .build();

        alert = alertRepository.save(alert);
        activeEmergencies.put(tunnelId, alert);

        cacheEmergencyStatus(tunnelId, alert);

        triggerEmergencyResponse(tunnelId, type);

        log.info("Emergency reported - ID: {}, Tunnel: {}, Type: {}", emergencyId, tunnelId, type);
        return alert;
    }

    private void triggerEmergencyResponse(String tunnelId, EmergencyType type) {
        activateEscapeLights(tunnelId, type);
        activateLedMarks(tunnelId, type);
    }

    @Transactional
    public List<EscapeLight> activateEscapeLights(String tunnelId, EmergencyType type) {
        List<EscapeLight> escapeLights = escapeLightRepository.findByTunnelId(tunnelId);

        if (escapeLights.isEmpty()) {
            escapeLights = initializeEscapeLights(tunnelId);
        }

        EscapeLightStatus status = determineEscapeLightStatus(type);
        String direction = determineEscapeDirection(type);

        for (EscapeLight light : escapeLights) {
            light.setStatus(status);
            light.setDirection(direction);
            light.setBrightness(emergencyConfig.getEscapeLight().getDefaultBrightness());
        }

        escapeLightRepository.saveAll(escapeLights);
        cacheEscapeLights(tunnelId, escapeLights);

        log.info("Activated escape lights for tunnel {}: {} devices", tunnelId, escapeLights.size());
        return escapeLights;
    }

    private List<EscapeLight> initializeEscapeLights(String tunnelId) {
        List<EscapeLight> lights = new ArrayList<>();
        List<EmergencyConfig.SegmentConfig> segments = emergencyConfig.getSegments();

        for (EmergencyConfig.SegmentConfig seg : segments) {
            for (int i = 0; i < 5; i++) {
                EscapeLight light = EscapeLight.builder()
                        .tunnelId(tunnelId)
                        .segmentId(seg.getId())
                        .deviceId(String.format("EL-%s-%s-%02d", tunnelId, seg.getId(), i + 1))
                        .status(EscapeLightStatus.OFF)
                        .brightness(0)
                        .build();
                lights.add(light);
            }
        }

        return escapeLightRepository.saveAll(lights);
    }

    private EscapeLightStatus determineEscapeLightStatus(EmergencyType type) {
        return switch (type) {
            case FIRE -> EscapeLightStatus.BLINK;
            case ACCIDENT, FAULT -> EscapeLightStatus.ON;
            default -> EscapeLightStatus.OFF;
        };
    }

    private String determineEscapeDirection(EmergencyType type) {
        return "BOTH";
    }

    @Transactional
    public List<LedMark> activateLedMarks(String tunnelId, EmergencyType type) {
        List<LedMark> ledMarks = ledMarkRepository.findByTunnelId(tunnelId);

        if (ledMarks.isEmpty()) {
            ledMarks = initializeLedMarks(tunnelId);
        }

        String color = emergencyConfig.getLedMark().getEmergencyColor();
        String pattern = determineLedPattern(type);

        for (LedMark mark : ledMarks) {
            mark.setColor(color);
            mark.setOn(true);
            mark.setPattern(pattern);
        }

        ledMarkRepository.saveAll(ledMarks);
        cacheLedMarks(tunnelId, ledMarks);

        log.info("Activated LED marks for tunnel {}: {} devices", tunnelId, ledMarks.size());
        return ledMarks;
    }

    private List<LedMark> initializeLedMarks(String tunnelId) {
        List<LedMark> marks = new ArrayList<>();
        List<EmergencyConfig.SegmentConfig> segments = emergencyConfig.getSegments();

        for (EmergencyConfig.SegmentConfig seg : segments) {
            for (int i = 0; i < 10; i++) {
                LedMark mark = LedMark.builder()
                        .tunnelId(tunnelId)
                        .segmentId(seg.getId())
                        .deviceId(String.format("LED-%s-%s-%02d", tunnelId, seg.getId(), i + 1))
                        .color(emergencyConfig.getLedMark().getNormalColor())
                        .on(true)
                        .pattern("SOLID")
                        .build();
                marks.add(mark);
            }
        }

        return ledMarkRepository.saveAll(marks);
    }

    private String determineLedPattern(EmergencyType type) {
        return switch (type) {
            case FIRE -> "FLASH";
            case ACCIDENT -> "ARROW_RIGHT";
            case FAULT -> "SLOW_BLINK";
            default -> "SOLID";
        };
    }

    @Transactional
    public EmergencyAlert clearEmergency(String tunnelId) {
        EmergencyAlert alert = activeEmergencies.get(tunnelId);
        if (alert == null) {
            alert = alertRepository.findByTunnelIdAndActiveTrue(tunnelId).orElse(null);
            if (alert == null) {
                return null;
            }
        }

        alert.setActive(false);
        alert.setResolvedAt(LocalDateTime.now());
        alertRepository.save(alert);

        activeEmergencies.remove(tunnelId);
        clearEmergencyCache(tunnelId);

        deactivateEscapeLights(tunnelId);
        deactivateLedMarks(tunnelId);

        log.info("Emergency cleared for tunnel {}", tunnelId);
        return alert;
    }

    @Transactional
    public List<EscapeLight> deactivateEscapeLights(String tunnelId) {
        List<EscapeLight> escapeLights = escapeLightRepository.findByTunnelId(tunnelId);

        for (EscapeLight light : escapeLights) {
            light.setStatus(EscapeLightStatus.OFF);
            light.setBrightness(0);
            light.setDirection(null);
        }

        escapeLightRepository.saveAll(escapeLights);
        cacheEscapeLights(tunnelId, escapeLights);

        log.info("Deactivated escape lights for tunnel {}", tunnelId);
        return escapeLights;
    }

    @Transactional
    public List<LedMark> deactivateLedMarks(String tunnelId) {
        List<LedMark> ledMarks = ledMarkRepository.findByTunnelId(tunnelId);

        for (LedMark mark : ledMarks) {
            mark.setColor(emergencyConfig.getLedMark().getNormalColor());
            mark.setPattern("SOLID");
        }

        ledMarkRepository.saveAll(ledMarks);
        cacheLedMarks(tunnelId, ledMarks);

        log.info("Deactivated LED marks for tunnel {}", tunnelId);
        return ledMarks;
    }

    public EmergencyAlert getActiveEmergency(String tunnelId) {
        return activeEmergencies.get(tunnelId);
    }

    public List<EscapeLight> getEscapeLights(String tunnelId) {
        List<EscapeLight> cached = getCachedEscapeLights(tunnelId);
        if (cached != null && !cached.isEmpty()) {
            return cached;
        }
        return escapeLightRepository.findByTunnelId(tunnelId);
    }

    public List<LedMark> getLedMarks(String tunnelId) {
        List<LedMark> cached = getCachedLedMarks(tunnelId);
        if (cached != null && !cached.isEmpty()) {
            return cached;
        }
        return ledMarkRepository.findByTunnelId(tunnelId);
    }

    public boolean hasActiveEmergency(String tunnelId) {
        return activeEmergencies.containsKey(tunnelId);
    }

    private void cacheEmergencyStatus(String tunnelId, EmergencyAlert alert) {
        String key = "emergency:status:" + tunnelId;
        redisTemplate.opsForValue().set(key, alert);
    }

    private void clearEmergencyCache(String tunnelId) {
        String key = "emergency:status:" + tunnelId;
        redisTemplate.delete(key);
    }

    private void cacheEscapeLights(String tunnelId, List<EscapeLight> lights) {
        String key = "emergency:escapelights:" + tunnelId;
        redisTemplate.opsForValue().set(key, lights);
    }

    @SuppressWarnings("unchecked")
    private List<EscapeLight> getCachedEscapeLights(String tunnelId) {
        String key = "emergency:escapelights:" + tunnelId;
        Object value = redisTemplate.opsForValue().get(key);
        if (value instanceof List) {
            return (List<EscapeLight>) value;
        }
        return null;
    }

    private void cacheLedMarks(String tunnelId, List<LedMark> marks) {
        String key = "emergency:ledmarks:" + tunnelId;
        redisTemplate.opsForValue().set(key, marks);
    }

    @SuppressWarnings("unchecked")
    private List<LedMark> getCachedLedMarks(String tunnelId) {
        String key = "emergency:ledmarks:" + tunnelId;
        Object value = redisTemplate.opsForValue().get(key);
        if (value instanceof List) {
            return (List<LedMark>) value;
        }
        return null;
    }
}
