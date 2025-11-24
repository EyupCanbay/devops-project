package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"pong-service/internal/config"
	"pong-service/internal/middleware"

	"github.com/labstack/echo/v4"
)

type PongHandler struct {
	Config *config.Config
}

func NewPongHandler(cfg *config.Config) *PongHandler {
	return &PongHandler{Config: cfg}
}

func (h *PongHandler) ReceivePing(c echo.Context) error {
	fmt.Println("[Pong Service] Ping request received")

	go h.sendAckBack()

	return c.String(http.StatusOK, "Pong")
}

func (h *PongHandler) sendAckBack() {
	start := time.Now()

	_, err := http.Post(h.Config.PingURL, "application/json", bytes.NewBuffer([]byte(`{}`)))

	duration := time.Since(start).Seconds()

	middleware.DependencyDuration.WithLabelValues("ping-service", "POST").Observe(duration)

	if err != nil {
		fmt.Printf("[ERROR] Callback failed %v\n", err)
	}
}