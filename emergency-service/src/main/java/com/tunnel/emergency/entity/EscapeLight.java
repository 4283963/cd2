package com.tunnel.emergency.entity;

import jakarta.persistence.*;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;
import lombok.Builder;

import java.time.LocalDateTime;

@Data
@Entity
@Table(name = "escape_lights")
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class EscapeLight {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "tunnel_id", length = 64, nullable = false)
    private String tunnelId;

    @Column(name = "segment_id", length = 64, nullable = false)
    private String segmentId;

    @Column(name = "device_id", length = 128, unique = true)
    private String deviceId;

    @Enumerated(EnumType.ORDINAL)
    @Column(name = "status", nullable = false)
    private EscapeLightStatus status;

    @Column(name = "direction", length = 32)
    private String direction;

    @Column(name = "brightness")
    private Integer brightness;

    @Column(name = "updated_at")
    private LocalDateTime updatedAt;

    @PrePersist
    @PreUpdate
    protected void onUpdate() {
        updatedAt = LocalDateTime.now();
    }
}
