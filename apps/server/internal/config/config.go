package config

import (
	"os"
	"time"
)

type AppServerConfig struct {
	Environment               string
	HTTPAddress               string
	AppServerPublicURL        string
	WebStaticDir              string
	RecorderWorkerInternalURL string
	PostgresDSN               string
	MinIOInternalURL          string
	MinIOPublicURL            string
	MinIOBucket               string
	MinIOAccessKey            string
	MinIOSecretKey            string
	SFUWebSocketPublicBaseURL string
	TURNPublicURL             string
	TURNInternalURL           string
	TURNUsername              string
	TURNPassword              string
}

type RecorderWorkerConfig struct {
	Environment                 string
	HTTPAddress                 string
	AppServerInternalURL        string
	SFUWebSocketInternalBaseURL string
	TURNInternalURL             string
	TURNUsername                string
	TURNPassword                string
	PollInterval                time.Duration
	RecordingChunkDuration      time.Duration
	PostgresDSN                 string
	MinIOInternalURL            string
	MinIOBucket                 string
	MinIOAccessKey              string
	MinIOSecretKey              string
}

func LoadAppServerConfig() AppServerConfig {
	legacySFUWebSocketBaseURL := getEnv("SFU_WS_BASE_URL", "ws://localhost:8080")
	legacyTURNURL := getEnv("TURN_URL", "turn:127.0.0.1:3478?transport=udp")
	return AppServerConfig{
		Environment:               getEnv("APP_ENV", "development"),
		HTTPAddress:               getEnv("APP_SERVER_HTTP_ADDR", ":8080"),
		AppServerPublicURL:        getEnv("APP_SERVER_PUBLIC_URL", "http://localhost:8080"),
		WebStaticDir:              getEnv("WEB_STATIC_DIR", ""),
		RecorderWorkerInternalURL: getEnvWithFallback("RECORDER_WORKER_INTERNAL_URL", "RECORDER_WORKER_URL", "http://localhost:8082"),
		PostgresDSN:               buildPostgresDSN(),
		MinIOInternalURL:          getEnvWithFallback("MINIO_INTERNAL_URL", "MINIO_ENDPOINT", "http://localhost:9000"),
		MinIOPublicURL:            getEnv("MINIO_PUBLIC_URL", ""),
		MinIOBucket:               getEnv("MINIO_BUCKET", "robot-center"),
		MinIOAccessKey:            getEnv("MINIO_ROOT_USER", "minioadmin"),
		MinIOSecretKey:            getEnv("MINIO_ROOT_PASSWORD", "minioadmin"),
		SFUWebSocketPublicBaseURL: getEnv("SFU_WS_PUBLIC_BASE_URL", legacySFUWebSocketBaseURL),
		TURNPublicURL:             getEnv("TURN_PUBLIC_URL", legacyTURNURL),
		TURNInternalURL:           getEnv("TURN_INTERNAL_URL", legacyTURNURL),
		TURNUsername:              getEnv("TURN_USERNAME", "robot"),
		TURNPassword:              getEnv("TURN_PASSWORD", "robot-pass"),
	}
}

func LoadRecorderWorkerConfig() RecorderWorkerConfig {
	legacySFUWebSocketBaseURL := getEnv("SFU_WS_BASE_URL", "ws://localhost:8080")
	legacyTURNURL := getEnv("TURN_URL", "turn:127.0.0.1:3478?transport=udp")
	return RecorderWorkerConfig{
		Environment:                 getEnv("APP_ENV", "development"),
		HTTPAddress:                 getEnv("RECORDER_WORKER_HTTP_ADDR", ":8082"),
		AppServerInternalURL:        getEnvWithFallback("APP_SERVER_INTERNAL_URL", "APP_SERVER_PUBLIC_URL", "http://localhost:8080"),
		SFUWebSocketInternalBaseURL: getEnv("SFU_WS_INTERNAL_BASE_URL", legacySFUWebSocketBaseURL),
		TURNInternalURL:             getEnv("TURN_INTERNAL_URL", legacyTURNURL),
		TURNUsername:                getEnv("TURN_USERNAME", "robot"),
		TURNPassword:                getEnv("TURN_PASSWORD", "robot-pass"),
		PollInterval:                getDurationEnv("RECORDER_WORKER_POLL_INTERVAL", 5*time.Second),
		RecordingChunkDuration:      getDurationEnv("RECORDING_CHUNK_DURATION", 10*time.Minute),
		PostgresDSN:                 buildPostgresDSN(),
		MinIOInternalURL:            getEnvWithFallback("MINIO_INTERNAL_URL", "MINIO_ENDPOINT", "http://localhost:9000"),
		MinIOBucket:                 getEnv("MINIO_BUCKET", "robot-center"),
		MinIOAccessKey:              getEnv("MINIO_ROOT_USER", "minioadmin"),
		MinIOSecretKey:              getEnv("MINIO_ROOT_PASSWORD", "minioadmin"),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvWithFallback(key string, legacyKey string, fallback string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	return getEnv(legacyKey, fallback)
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return duration
}

func buildPostgresDSN() string {
	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	database := getEnv("POSTGRES_DB", "robot_center")
	user := getEnv("POSTGRES_USER", "robot_center")
	password := getEnv("POSTGRES_PASSWORD", "robot_center")
	return "postgres://" + user + ":" + password + "@" + host + ":" + port + "/" + database + "?sslmode=disable"
}
