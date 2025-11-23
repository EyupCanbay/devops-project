package main

import (
	"pong-service/internal/config"
	"pong-service/internal/handlers"
	"pong-service/internal/middleware"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {
	cfg := config.LoadConfig()

	middleware.InitMetrics()

	e := echo.New()
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(middleware.PrometheusMiddleware)

	h := handlers.NewPongHandler(cfg)

	e.GET("/metrics", middleware.MetricsHandler())
	e.POST("/receive-ping", h.ReceivePing)

	e.Logger.Fatal(e.Start(cfg.PongPort))
}
