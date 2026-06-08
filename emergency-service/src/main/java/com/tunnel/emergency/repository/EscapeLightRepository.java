package com.tunnel.emergency.repository;

import com.tunnel.emergency.entity.EscapeLight;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface EscapeLightRepository extends JpaRepository<EscapeLight, Long> {

    List<EscapeLight> findByTunnelId(String tunnelId);

    List<EscapeLight> findByTunnelIdAndSegmentId(String tunnelId, String segmentId);

    EscapeLight findByDeviceId(String deviceId);
}
