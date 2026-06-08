package com.tunnel.emergency.entity;

import jakarta.persistence.*;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;
import lombok.Builder;

import java.time.LocalDateTime;

@Data
@Entity
@Table(name = "emergency_alerts")
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class EmergencyAlert {

    @Id
    @Column(name = "emergency_id", length = 64)
    private String emergencyId;

    @Column(name = "tunnel_id", length = 64, nullable = false)
    private String tunnelId;

    @Enumerated(EnumType.ORDINAL)
    @Column(name = "type", nullable = false)
    private EmergencyType type;

    @Column(name = "location", length = 256)
    private String location;

    @Column(name = "description", length = 1024)
    private String description;

    @Column(name = "severity")
    private Integer severity;

    @Column(name = "is_active")
    private Boolean active;

    @Column(name = "created_at")
    private LocalDateTime createdAt;

    @Column(name = "resolved_at")
    private LocalDateTime resolvedAt;

    @PrePersist
    protected void onCreate() {
        createdAt = LocalDateTime.now();
        if (active == null) {
            active = true;
        }
    }
}
