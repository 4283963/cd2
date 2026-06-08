package com.tunnel.emergency.mqtt;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.tunnel.emergency.entity.EmergencyType;
import com.tunnel.emergency.entity.EscapeLight;
import com.tunnel.emergency.entity.LedMark;
import com.tunnel.emergency.service.EmergencyDispatchService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.integration.annotation.ServiceActivator;
import org.springframework.integration.channel.DirectChannel;
import org.springframework.integration.core.MessageProducer;
import org.springframework.integration.mqtt.core.DefaultMqttPahoClientFactory;
import org.springframework.integration.mqtt.core.MqttPahoClientFactory;
import org.springframework.integration.mqtt.inbound.MqttPahoMessageDrivenChannelAdapter;
import org.springframework.integration.mqtt.outbound.MqttPahoMessageHandler;
import org.springframework.integration.mqtt.support.DefaultPahoMessageConverter;
import org.springframework.messaging.MessageChannel;
import org.springframework.messaging.MessageHandler;

import java.util.Map;

@Slf4j
@Configuration
@RequiredArgsConstructor
public class MqttConfiguration {

    private final MqttConfig mqttConfig;
    private final EmergencyDispatchService dispatchService;
    private final ObjectMapper objectMapper;

    @Bean
    public MqttPahoClientFactory mqttClientFactory() {
        DefaultMqttPahoClientFactory factory = new DefaultMqttPahoClientFactory();
        var options = new org.eclipse.paho.client.mqttv3.MqttConnectOptions();
        options.setServerURIs(new String[]{mqttConfig.getBroker()});
        if (mqttConfig.getUsername() != null && !mqttConfig.getUsername().isEmpty()) {
            options.setUserName(mqttConfig.getUsername());
            options.setPassword(mqttConfig.getPassword().toCharArray());
        }
        options.setCleanSession(true);
        options.setConnectionTimeout(10);
        options.setKeepAliveInterval(30);
        factory.setConnectionOptions(options);
        return factory;
    }

    @Bean
    public MessageChannel mqttInputChannel() {
        return new DirectChannel();
    }

    @Bean
    public MessageProducer mqttInbound() {
        String[] topics = mqttConfig.getTopics().values().toArray(new String[0]);
        MqttPahoMessageDrivenChannelAdapter adapter =
                new MqttPahoMessageDrivenChannelAdapter(
                        mqttConfig.getClientId() + "-in",
                        mqttClientFactory(),
                        topics);
        adapter.setCompletionTimeout(5000);
        adapter.setConverter(new DefaultPahoMessageConverter());
        adapter.setQos(1);
        adapter.setOutputChannel(mqttInputChannel());
        return adapter;
    }

    @Bean
    @ServiceActivator(inputChannel = "mqttInputChannel")
    public MessageHandler mqttMessageHandler() {
        return message -> {
            String topic = (String) message.getHeaders().get("mqtt_receivedTopic");
            String payload = message.getPayload().toString();

            log.debug("Received MQTT message - Topic: {}, Payload: {}", topic, payload);

            try {
                handleMessage(topic, payload);
            } catch (Exception e) {
                log.error("Error handling MQTT message: {}", e.getMessage(), e);
            }
        };
    }

    private void handleMessage(String topic, String payload) throws Exception {
        Map<String, String> topics = mqttConfig.getTopics();

        if (topic.matches(topics.get("emergency-alert").replace("+", "[^/]+"))) {
            handleEmergencyAlert(payload);
        } else if (topic.contains("escape_light")) {
            handleEscapeLightStatus(payload);
        }
    }

    @SuppressWarnings("unchecked")
    private void handleEmergencyAlert(String payload) throws Exception {
        Map<String, Object> data = objectMapper.readValue(payload, Map.class);

        String tunnelId = (String) data.get("tunnelId");
        if (tunnelId == null) {
            tunnelId = "T001";
        }

        String typeStr = (String) data.get("type");
        EmergencyType type = EmergencyType.NONE;
        if (typeStr != null) {
            try {
                type = EmergencyType.valueOf(typeStr.toUpperCase());
            } catch (IllegalArgumentException ignored) {
            }
        }

        String location = (String) data.get("location");
        String description = (String) data.get("description");
        Integer severity = data.get("severity") != null ? ((Number) data.get("severity")).intValue() : null;

        if (type != EmergencyType.NONE) {
            dispatchService.reportEmergency(tunnelId, type, location, description, severity);
        }
    }

    private void handleEscapeLightStatus(String payload) {
        log.debug("Escape light status update: {}", payload);
    }

    @Bean
    public MessageChannel mqttOutboundChannel() {
        return new DirectChannel();
    }

    @Bean
    @ServiceActivator(inputChannel = "mqttOutboundChannel")
    public MessageHandler mqttOutbound() {
        MqttPahoMessageHandler messageHandler = new MqttPahoMessageHandler(
                mqttConfig.getClientId() + "-out", mqttClientFactory());
        messageHandler.setAsync(true);
        messageHandler.setDefaultTopic("tunnel/default");
        return messageHandler;
    }
}
