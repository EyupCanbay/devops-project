package handlers

import (
	"bytes"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"pong-service/internal/config"
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
	_, err := http.Post(h.Config.PingURL, "application/json", bytes.NewBuffer([]byte(`{}`)))
	if err != nil {
		fmt.Println("[ERROR] Callback failed %v\n", err)
	}
}
