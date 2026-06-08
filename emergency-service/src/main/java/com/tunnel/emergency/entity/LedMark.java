package com.tunnel.emergency.entity;

import jakarta.persistence.*;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;
import lombok.Builder;

import java.time.LocalDateTime;

@Data
@Entity
@Table(name = "led_marks")
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class LedMark {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "tunnel_id", length = 64, nullable = false)
    private String tunnelId;

    @Column(name = "segment_id", length = 64, nullable = false)
    private String segmentId;

    @Column(name = "device_id", length = 128, unique = true)
    private String deviceId;

    @Column(name = "color", length = 16)
    private String color;

    @Column(name = "is_on")
    private Boolean on;

    @Column(name = "pattern", length = 32)
    private String pattern;

    @Column(name = "updated_at")
    private LocalDateTime updatedAt;

    @PrePersist
    @PreUpdate
    protected void onUpdate() {
        updatedAt = LocalDateTime.now();
        if (on == null) {
            on = true;
        }
    }
}
