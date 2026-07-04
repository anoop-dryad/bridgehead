package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	App     App
	HTTP    HTTP
	DB      DB
	Redis   Redis
	MQTT    MQTT
	Kinesis Kinesis
	SQS     SQS
}

type App struct {
	IsProduction bool
	Name         string
	Version      string
}

type HTTP struct {
	Addr string
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
	Gateway GatewayMQTT
	TTN     TTNMQTT
}

type GatewayMQTT struct {
	BrokerURL string
	ClientID  string
}

type TTNMQTT struct {
	BrokerURL string
	Apps      []TTNApp
}

type TTNApp struct {
	AppID    string
	Username string
	Password string
}

type Kinesis struct {
	StreamName string
	Region     string
	DSN        string
}

type SQS struct {
	QueueURL string
	Region   string
}

func Load() Config {
	return Config{
		App: App{
			IsProduction: getEnv("ENV", "development") == "production",
			Name:         getEnv("APP_NAME", "bridgehead"),
			Version:      getEnv("APP_VERSION", "dev"),
		},
		HTTP: HTTP{
			Addr: getEnv("HTTP_ADDR", ":8080"),
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
			Gateway: GatewayMQTT{
				BrokerURL: mustEnv("GATEWAY_MQTT_BROKER_URL"),
				ClientID:  getEnv("GATEWAY_MQTT_CLIENT_ID", "bridgehead-gw"),
			},
			TTN: TTNMQTT{
				BrokerURL: mustEnv("TTN_MQTT_BROKER_URL"),
				Apps:      loadTTNApps(),
			},
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

func loadTTNApps() []TTNApp {
	appIDs := strings.Split(mustEnv("TTN_APP_IDS"), ",")
	var apps []TTNApp
	for _, id := range appIDs {
		id = strings.TrimSpace(id)
		apps = append(apps, TTNApp{
			AppID:    id,
			Username: mustEnv(fmt.Sprintf("TTN_APP_%s_USERNAME", id)),
			Password: mustEnv(fmt.Sprintf("TTN_APP_%s_PASSWORD", id)),
		})
	}
	return apps
}
