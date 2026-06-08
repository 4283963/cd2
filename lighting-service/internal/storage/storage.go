package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"lighting-service/internal/model"
)

type Config struct {
	RedisHost     string
	RedisPort     int
	RedisPassword string
	RedisDB       int
	CacheTTL      int
	MySQLHost     string
	MySQLPort     int
	MySQLDatabase string
	MySQLUsername string
	MySQLPassword string
	MaxOpenConns  int
	MaxIdleConns  int
}

type Storage struct {
	redisClient *redis.Client
	mysqlDB     *sql.DB
	config      Config
	ctx         context.Context
}

func NewStorage(config Config) (*Storage, error) {
	s := &Storage{
		config: config,
		ctx:    context.Background(),
	}

	if err := s.initRedis(); err != nil {
		return nil, fmt.Errorf("init redis failed: %w", err)
	}

	if err := s.initMySQL(); err != nil {
		log.Printf("Warning: init mysql failed: %v", err)
	}

	return s, nil
}

func (s *Storage) initRedis() error {
	s.redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", s.config.RedisHost, s.config.RedisPort),
		Password: s.config.RedisPassword,
		DB:       s.config.RedisDB,
	})

	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()

	if err := s.redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	log.Println("Redis connected successfully")
	return nil
}

func (s *Storage) initMySQL() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		s.config.MySQLUsername,
		s.config.MySQLPassword,
		s.config.MySQLHost,
		s.config.MySQLPort,
		s.config.MySQLDatabase,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("mysql open failed: %w", err)
	}

	db.SetMaxOpenConns(s.config.MaxOpenConns)
	db.SetMaxIdleConns(s.config.MaxIdleConns)
	db.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("mysql ping failed: %w", err)
	}

	s.mysqlDB = db
	log.Println("MySQL connected successfully")
	return nil
}

func (s *Storage) CacheDeviceStatus(deviceID string, status model.DeviceStatus) error {
	data, err := json.Marshal(status)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("device:%s", deviceID)
	ttl := time.Duration(s.config.CacheTTL) * time.Second

	return s.redisClient.Set(s.ctx, key, data, ttl).Err()
}

func (s *Storage) GetDeviceStatus(deviceID string) (*model.DeviceStatus, error) {
	key := fmt.Sprintf("device:%s", deviceID)
	data, err := s.redisClient.Get(s.ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var status model.DeviceStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, err
	}

	return &status, nil
}

func (s *Storage) CacheLightingSegments(tunnelID string, segments []*model.LightingSegment) error {
	data, err := json.Marshal(segments)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("lighting:segments:%s", tunnelID)
	ttl := time.Duration(s.config.CacheTTL) * time.Second

	return s.redisClient.Set(s.ctx, key, data, ttl).Err()
}

func (s *Storage) GetLightingSegments(tunnelID string) ([]*model.LightingSegment, error) {
	key := fmt.Sprintf("lighting:segments:%s", tunnelID)
	data, err := s.redisClient.Get(s.ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var segments []*model.LightingSegment
	if err := json.Unmarshal(data, &segments); err != nil {
		return nil, err
	}

	return segments, nil
}

func (s *Storage) SaveLightingSegment(segment *model.LightingSegment) error {
	if s.mysqlDB == nil {
		return fmt.Errorf("mysql not connected")
	}

	query := `INSERT INTO lighting_segments (id, tunnel_id, name, brightness, mode, length, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE brightness=VALUES(brightness), mode=VALUES(mode), updated_at=VALUES(updated_at)`

	_, err := s.mysqlDB.Exec(query,
		segment.ID,
		segment.TunnelID,
		segment.Name,
		segment.Brightness,
		segment.Mode,
		segment.Length,
		segment.UpdatedAt,
	)

	return err
}

func (s *Storage) GetLightingSegmentsFromDB(tunnelID string) ([]*model.LightingSegment, error) {
	if s.mysqlDB == nil {
		return nil, fmt.Errorf("mysql not connected")
	}

	query := `SELECT id, tunnel_id, name, brightness, mode, length, updated_at
		FROM lighting_segments WHERE tunnel_id = ?`

	rows, err := s.mysqlDB.Query(query, tunnelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var segments []*model.LightingSegment
	for rows.Next() {
		seg := &model.LightingSegment{}
		err := rows.Scan(&seg.ID, &seg.TunnelID, &seg.Name, &seg.Brightness, &seg.Mode, &seg.Length, &seg.UpdatedAt)
		if err != nil {
			return nil, err
		}
		segments = append(segments, seg)
	}

	return segments, nil
}

func (s *Storage) SetEmergencyMode(tunnelID string, emergencyType model.EmergencyType) error {
	key := fmt.Sprintf("emergency:%s", tunnelID)
	value := map[string]interface{}{
		"type":       emergencyType.String(),
		"updated_at": time.Now().Unix(),
	}
	return s.redisClient.HSet(s.ctx, key, value).Err()
}

func (s *Storage) ClearEmergencyMode(tunnelID string) error {
	key := fmt.Sprintf("emergency:%s", tunnelID)
	return s.redisClient.Del(s.ctx, key).Err()
}

func (s *Storage) GetEmergencyStatus(tunnelID string) (bool, model.EmergencyType, error) {
	key := fmt.Sprintf("emergency:%s", tunnelID)
	exists, err := s.redisClient.Exists(s.ctx, key).Result()
	if err != nil {
		return false, model.EmergencyNone, err
	}
	if exists == 0 {
		return false, model.EmergencyNone, nil
	}

	typeStr, err := s.redisClient.HGet(s.ctx, key, "type").Result()
	if err != nil {
		return true, model.EmergencyNone, err
	}

	var etype model.EmergencyType
	switch typeStr {
	case "ACCIDENT":
		etype = model.EmergencyAccident
	case "FIRE":
		etype = model.EmergencyFire
	case "FAULT":
		etype = model.EmergencyFault
	default:
		etype = model.EmergencyNone
	}

	return true, etype, nil
}

func (s *Storage) Close() error {
	if s.redisClient != nil {
		s.redisClient.Close()
	}
	if s.mysqlDB != nil {
		s.mysqlDB.Close()
	}
	return nil
}
