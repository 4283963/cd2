package com.tunnel.emergency.mqtt;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.tunnel.emergency.entity.EscapeLight;
import com.tunnel.emergency.entity.LedMark;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.integration.mqtt.outbound.MqttPahoMessageHandler;
import org.springframework.messaging.Message;
import org.springframework.messaging.support.MessageBuilder;
import org.springframework.stereotype.Service;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Slf4j
@Service
@RequiredArgsConstructor
public class MqttPublisher {

    private final MqttConfig mqttConfig;
    private final MqttPahoMessageHandler mqttOutbound;
    private final ObjectMapper objectMapper;

    public void publishEscapeLightControl(String tunnelId, String segmentId, EscapeLight light) {
        try {
            String topicTemplate = mqttConfig.getPublishTopics().get("escape-light-control");
            String topic = topicTemplate.replace("{tunnelId}", tunnelId).replace("{segmentId}", segmentId);

            Map<String, Object> payload = new HashMap<>();
            payload.put("deviceId", light.getDeviceId());
            payload.put("status", light.getStatus().name());
            payload.put("brightness", light.getBrightness());
            payload.put("direction", light.getDirection());
            payload.put("timestamp", System.currentTimeMillis());

            String jsonPayload = objectMapper.writeValueAsString(payload);
            publish(topic, jsonPayload);

            log.debug("Published escape light control to {}: {}", topic, jsonPayload);
        } catch (Exception e) {
            log.error("Failed to publish escape light control: {}", e.getMessage(), e);
        }
    }

    public void publishLedMarkControl(String tunnelId, String segmentId, LedMark mark) {
        try {
            String topicTemplate = mqttConfig.getPublishTopics().get("led-mark-control");
            String topic = topicTemplate.replace("{tunnelId}", tunnelId).replace("{segmentId}", segmentId);

            Map<String, Object> payload = new HashMap<>();
            payload.put("deviceId", mark.getDeviceId());
            payload.put("color", mark.getColor());
            payload.put("on", mark.getOn());
            payload.put("pattern", mark.getPattern());
            payload.put("timestamp", System.currentTimeMillis());

            String jsonPayload = objectMapper.writeValueAsString(payload);
            publish(topic, jsonPayload);

            log.debug("Published LED mark control to {}: {}", topic, jsonPayload);
        } catch (Exception e) {
            log.error("Failed to publish LED mark control: {}", e.getMessage(), e);
        }
    }

    public void publishEmergencyAlert(String tunnelId, Map<String, Object> alertData) {
        try {
            String topic = "tunnel/" + tunnelId + "/emergency";
            String jsonPayload = objectMapper.writeValueAsString(alertData);
            publish(topic, jsonPayload);

            log.info("Published emergency alert to {}", topic);
        } catch (Exception e) {
            log.error("Failed to publish emergency alert: {}", e.getMessage(), e);
        }
    }

    private void publish(String topic, String payload) {
        Message<String> message = MessageBuilder.withPayload(payload)
                .setHeader("mqtt_topic", topic)
                .build();
        mqttOutbound.handleMessage(message);
    }

    public void publishBulkEscapeLights(String tunnelId, List<EscapeLight> lights) {
        Map<String, List<EscapeLight>> bySegment = new HashMap<>();
        for (EscapeLight light : lights) {
            bySegment.computeIfAbsent(light.getSegmentId(), k -> new java.util.ArrayList<>()).add(light);
        }

        for (Map.Entry<String, List<EscapeLight>> entry : bySegment.entrySet()) {
            for (EscapeLight light : entry.getValue()) {
                publishEscapeLightControl(tunnelId, entry.getKey(), light);
            }
        }
    }

    public void publishBulkLedMarks(String tunnelId, List<LedMark> marks) {
        Map<String, List<LedMark>> bySegment = new HashMap<>();
        for (LedMark mark : marks) {
            bySegment.computeIfAbsent(mark.getSegmentId(), k -> new java.util.ArrayList<>()).add(mark);
        }

        for (Map.Entry<String, List<LedMark>> entry : bySegment.entrySet()) {
            for (LedMark mark : entry.getValue()) {
                publishLedMarkControl(tunnelId, entry.getKey(), mark);
            }
        }
    }
}
