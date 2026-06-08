package com.tunnel.emergency.repository;

import com.tunnel.emergency.entity.EmergencyAlert;
import com.tunnel.emergency.entity.EmergencyType;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface EmergencyAlertRepository extends JpaRepository<EmergencyAlert, String> {

    Optional<EmergencyAlert> findByTunnelIdAndActiveTrue(String tunnelId);

    List<EmergencyAlert> findByTunnelIdOrderByCreatedAtDesc(String tunnelId);

    List<EmergencyAlert> findByTypeAndActiveTrue(EmergencyType type);

    List<EmergencyAlert> findByActiveTrueOrderByCreatedAtDesc();
}
