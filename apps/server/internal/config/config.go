package config

import (
	"os"
	"time"
)

type AppServerConfig struct {
	Environment       string
	HTTPAddress       string
	PublicURL         string
	WebStaticDir      string
	RecorderWorkerURL string
	PostgresDSN       string
	MinIOEndpoint     string
	MinIOBucket       string
	SFUWebSocketURL   string
	TURNURL           string
	TURNUsername      string
	TURNPassword      string
}

type RecorderWorkerConfig struct {
	Environment            string
	HTTPAddress            string
	AppServerURL           string
	SFUWebSocketURL        string
	TURNURL                string
	TURNUsername           string
	TURNPassword           string
	PollInterval           time.Duration
	RecordingChunkDuration time.Duration
	PostgresDSN            string
	MinIOEndpoint          string
	MinIOBucket            string
	MinIOAccessKey         string
	MinIOSecretKey         string
}

func LoadAppServerConfig() AppServerConfig {
	return AppServerConfig{
		Environment:       getEnv("APP_ENV", "development"),
		HTTPAddress:       getEnv("APP_SERVER_HTTP_ADDR", ":8080"),
		PublicURL:         getEnv("APP_SERVER_PUBLIC_URL", "http://localhost:8080"),
		WebStaticDir:      getEnv("WEB_STATIC_DIR", ""),
		RecorderWorkerURL: getEnv("RECORDER_WORKER_URL", "http://localhost:8082"),
		PostgresDSN:       buildPostgresDSN(),
		MinIOEndpoint:     getEnv("MINIO_ENDPOINT", "http://localhost:9000"),
		MinIOBucket:       getEnv("MINIO_BUCKET", "robot-center"),
		SFUWebSocketURL:   getEnv("SFU_WS_URL", "ws://localhost:8080/sfu/ws"),
		TURNURL:           getEnv("TURN_URL", "turn:127.0.0.1:3478?transport=udp"),
		TURNUsername:      getEnv("TURN_USERNAME", "robot"),
		TURNPassword:      getEnv("TURN_PASSWORD", "robot-pass"),
	}
}

func LoadRecorderWorkerConfig() RecorderWorkerConfig {
	return RecorderWorkerConfig{
		Environment:            getEnv("APP_ENV", "development"),
		HTTPAddress:            getEnv("RECORDER_WORKER_HTTP_ADDR", ":8082"),
		AppServerURL:           getEnv("APP_SERVER_PUBLIC_URL", "http://localhost:8080"),
		SFUWebSocketURL:        getEnv("SFU_WS_URL", "ws://localhost:8080/sfu/ws"),
		TURNURL:                getEnv("TURN_URL", "turn:127.0.0.1:3478?transport=udp"),
		TURNUsername:           getEnv("TURN_USERNAME", "robot"),
		TURNPassword:           getEnv("TURN_PASSWORD", "robot-pass"),
		PollInterval:           getDurationEnv("RECORDER_WORKER_POLL_INTERVAL", 5*time.Second),
		RecordingChunkDuration: getDurationEnv("RECORDING_CHUNK_DURATION", 10*time.Minute),
		PostgresDSN:            buildPostgresDSN(),
		MinIOEndpoint:          getEnv("MINIO_ENDPOINT", "http://localhost:9000"),
		MinIOBucket:            getEnv("MINIO_BUCKET", "robot-center"),
		MinIOAccessKey:         getEnv("MINIO_ROOT_USER", "minioadmin"),
		MinIOSecretKey:         getEnv("MINIO_ROOT_PASSWORD", "minioadmin"),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
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
