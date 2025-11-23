package config

import "os"

type Config struct {
	PongURL  string
	PingPort string
}

func Load() *Config {
	return &Config{
		PongURL:  getEnv("PONG_SERVICE_URL", "http://localhost:8081/receive-ping"),
		PingPort: getEnv("PORT", ":8080"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
