package com.tunnel.emergency.repository;

import com.tunnel.emergency.entity.LedMark;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface LedMarkRepository extends JpaRepository<LedMark, Long> {

    List<LedMark> findByTunnelId(String tunnelId);

    List<LedMark> findByTunnelIdAndSegmentId(String tunnelId, String segmentId);

    LedMark findByDeviceId(String deviceId);
}
