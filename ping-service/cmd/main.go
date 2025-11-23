package main

import (
	"ping-service/internal/config"
	"ping-service/internal/handlers"
	"ping-service/internal/middleware"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {
	cfg := config.Load()

	middleware.InitMetrics()

	e := echo.New()
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(middleware.PrometheusMiddleware)

	h := handlers.NewPingHandler(cfg)

	e.GET("/metrics", middleware.MetricsHandler())
	e.GET("/start", h.StartSimulation)
	e.POST("/callback", h.Callback)

	e.Logger.Fatal(e.Start(cfg.PingPort))

}
