package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Config struct {
	PongURL string
	Port    string
}

func getConfig() Config {
	return Config{
		PongURL: getEnv("PONG_SERVICE_URL", "http://localhost:8081/receive-ping"),
		Port:    getEnv("PORT", ":8080"),
	}
}

func main() {
	e := echo.New()
	cfg := getConfig()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/start", func(c echo.Context) error {
		go startTrafficSimulation(cfg.PongURL)
		return c.String(http.StatusOK, "Traffic smulation starting..! (it will last 30 second )")
	})

	e.POST("/callback", func(c echo.Context) error {
		fmt.Println("[Ping Service] from pong 'ı am here' response received.")
		return c.JSON(http.StatusOK, map[string]string{"status": "ack"})
	})

	e.Logger.Fatal(e.Start(cfg.Port))
}

func startTrafficSimulation(targetURL string) {
	fmt.Println("[Loop] 30 second loop starting...")

	timer := time.NewTimer(30 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timer.C:
			fmt.Println("[Loop] time over. simulation finished...")
			return
		case <-ticker.C:
			sendPing(targetURL)
		}
	}
}

func sendPing(url string) {
	fmt.Println("[Ping Service] sending request to pong...")
	resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(`{}`)))
	if err != nil {
		fmt.Printf("[Error] request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
