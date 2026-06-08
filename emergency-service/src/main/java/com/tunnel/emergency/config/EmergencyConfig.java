package com.tunnel.emergency.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Configuration;

import java.util.List;

@Data
@Configuration
@ConfigurationProperties(prefix = "emergency")
public class EmergencyConfig {

    private EscapeLightConfig escapeLight;
    private LedMarkConfig ledMark;
    private List<SegmentConfig> segments;

    @Data
    public static class EscapeLightConfig {
        private int blinkInterval;
        private int defaultBrightness;
    }

    @Data
    public static class LedMarkConfig {
        private String emergencyColor;
        private String normalColor;
    }

    @Data
    public static class SegmentConfig {
        private String id;
        private String name;
    }
}
