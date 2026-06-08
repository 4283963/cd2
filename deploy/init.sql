CREATE DATABASE IF NOT EXISTS tunnel DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE tunnel;

CREATE TABLE IF NOT EXISTS lighting_segments (
    id VARCHAR(64) NOT NULL PRIMARY KEY,
    tunnel_id VARCHAR(64) NOT NULL,
    name VARCHAR(128) NOT NULL,
    brightness DOUBLE NOT NULL DEFAULT 0,
    mode INT NOT NULL DEFAULT 0,
    length DOUBLE NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_tunnel_id (tunnel_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS emergency_alerts (
    emergency_id VARCHAR(64) NOT NULL PRIMARY KEY,
    tunnel_id VARCHAR(64) NOT NULL,
    type INT NOT NULL DEFAULT 0,
    location VARCHAR(256),
    description VARCHAR(1024),
    severity INT DEFAULT 3,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP NULL,
    INDEX idx_tunnel_id (tunnel_id),
    INDEX idx_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS escape_lights (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    tunnel_id VARCHAR(64) NOT NULL,
    segment_id VARCHAR(64) NOT NULL,
    device_id VARCHAR(128) UNIQUE,
    status INT NOT NULL DEFAULT 0,
    direction VARCHAR(32),
    brightness INT DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_tunnel_segment (tunnel_id, segment_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS led_marks (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    tunnel_id VARCHAR(64) NOT NULL,
    segment_id VARCHAR(64) NOT NULL,
    device_id VARCHAR(128) UNIQUE,
    color VARCHAR(16),
    is_on BOOLEAN DEFAULT TRUE,
    pattern VARCHAR(32),
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_tunnel_segment (tunnel_id, segment_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT IGNORE INTO lighting_segments (id, tunnel_id, name, brightness, mode, length) VALUES
('S01', 'T001', '入口段', 100, 0, 50),
('S02', 'T001', '过渡段1', 80, 0, 50),
('S03', 'T001', '过渡段2', 60, 0, 50),
('S04', 'T001', '中间段', 40, 0, 200),
('S05', 'T001', '出口段', 80, 0, 50);
