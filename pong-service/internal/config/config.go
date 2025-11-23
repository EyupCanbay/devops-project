package config

import "os"

type Config struct {
	PingURL  string
	PongPort string
}

func LoadConfig() *Config {
	return &Config{
		PingURL:  getEnv("PING_SERVICE_URL", "http://localhost:8080/callback"),
		PongPort: getEnv("PONGPORT", ":8081"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
