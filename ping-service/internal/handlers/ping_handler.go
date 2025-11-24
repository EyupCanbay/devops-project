package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"ping-service/internal/config"
	"ping-service/internal/middleware"

	"github.com/labstack/echo/v4"
)

type PingHandler struct {
	Config *config.Config
}

func NewPingHandler(cfg *config.Config) *PingHandler {
	return &PingHandler{Config: cfg}
}

// StartSimulation : /start
func (h *PingHandler) StartSimulation(c echo.Context) error {
	go h.runTrafficLoop()
	return c.String(http.StatusOK, "starting traffic loop")
}

// Callback : /callback
func (h *PingHandler) Callback(c echo.Context) error {
	fmt.Println("[Ping Service] Request received from Pong service.")
	return c.JSON(http.StatusOK, map[string]string{"status": "ack"})
}

func (h *PingHandler) runTrafficLoop() {
	fmt.Println("[LOOP] Starting traffic loop for 30 seconds.")
	timer := time.NewTimer(time.Second * 30)
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for {
		select {
		case <-timer.C:
			fmt.Println("[LOOP] Time over")
			return
		case <-ticker.C:
			h.sendPing()
		}
	}
}

func (h *PingHandler) sendPing() {
	start := time.Now()

	resp, err := http.Post(h.Config.PongURL, "application/json", bytes.NewBuffer([]byte("{}")))
	
	duration := time.Since(start).Seconds()

	middleware.DependencyDuration.WithLabelValues("pong-service", "POST").Observe(duration)

	if err != nil {
		fmt.Printf("[LOOP] Error sending ping %v\n", err)
		return
	}
	defer resp.Body.Close()
}