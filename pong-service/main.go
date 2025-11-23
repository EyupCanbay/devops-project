package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Config struct {
	PingCallbackURL string
	Port            string
}

func getConfig() Config {
	return Config{
		PingCallbackURL: getEnv("PING_SERVICE_URL", "http://localhost:8080/callback"),
		Port:            getEnv("PORT", ":8081"),
	}
}

func main() {
	e := echo.New()
	cfg := getConfig()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.POST("/receive-ping", func(c echo.Context) error {
		fmt.Println("[Pong Service] Ping request received!")
		go sendAckBack(cfg.PingCallbackURL)

		return c.String(http.StatusOK, "Pong!")
	})

	e.Logger.Fatal(e.Start(cfg.Port))
}

func sendAckBack(url string) {
	fmt.Println("[Pong Service] Sending callback back to Ping...")
	_, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(`{}`)))
	if err != nil {
		fmt.Printf("[Error] Callback failed: %v\n", err)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
