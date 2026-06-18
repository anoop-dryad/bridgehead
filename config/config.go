package config

import (
	"os"
)

type Config struct {
	HTTP    HTTP
	DB      DB
	Redis   Redis
	MQTT    MQTT
	Kinesis Kinesis
	SQS     SQS
}

type HTTP struct {
	Addr         string
	IsProduction bool
}

type DB struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
}

type Redis struct {
	Addr     string
	Password string
}

type MQTT struct {
	BrokerURL string
	ClientID  string
	Topic     string
}

type Kinesis struct {
	StreamName string
	Region     string
}

type SQS struct {
	QueueURL string
	Region   string
}

func Load() Config {
	return Config{
		HTTP: HTTP{
			Addr:         getEnv("HTTP_ADDR", ":8080"),
			IsProduction: getEnv("ENV", "development") == "production",
		},
		DB: DB{
			DSN:          mustEnv("DB_DSN"),
			MaxOpenConns: 25,
			MaxIdleConns: 5,
		},
		Redis: Redis{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
		},
		MQTT: MQTT{
			BrokerURL: mustEnv("MQTT_BROKER_URL"),
			ClientID:  getEnv("MQTT_CLIENT_ID", "downlink-service"),
			Topic:     mustEnv("MQTT_TOPIC"),
		},
		Kinesis: Kinesis{
			StreamName: mustEnv("KINESIS_STREAM_NAME"),
			Region:     getEnv("AWS_REGION", "eu-central-1"),
		},
		SQS: SQS{
			QueueURL: mustEnv("SQS_QUEUE_URL"),
			Region:   getEnv("AWS_REGION", "eu-central-1"),
		},
	}
}

// mustEnv panics at startup if a required variable is missing.
// better to crash immediately than fail silently in production.
func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic("required environment variable not set: " + key)
	}
	return val
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
