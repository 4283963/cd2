package com.tunnel.emergency.mqtt;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Configuration;

import java.util.Map;

@Data
@Configuration
@ConfigurationProperties(prefix = "mqtt")
public class MqttConfig {

    private String broker;
    private String clientId;
    private String username;
    private String password;
    private Map<String, String> topics;
    private Map<String, String> publishTopics;
}
