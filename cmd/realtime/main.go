package main

import (
	"RealtimeService/config"
	"RealtimeService/internal/domains/connection/handler"
	"RealtimeService/internal/domains/connection/repository"
	"RealtimeService/internal/domains/connection/usecase"
	"RealtimeService/internal/server/hub"
	stream "RealtimeService/internal/stream/usecase"
	"RealtimeService/logger"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	// 👇 добавляем pprof
	_ "net/http/pprof"
)

func main() {
	logger.InitLogger()
	slog.Info("Logger initialized")

	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	h := hub.New()

	kafka := stream.NewKafkaConsumer(h, cfg.KafkaTopic, cfg.KafkaGroup, cfg.KafkaBrokers)
	if err := kafka.Start(ctx); err != nil {
		slog.Error("Kafka consumer failed to start", slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Info("Kafka consumer started")

	auth := repository.NewAuthRepository()
	unit := repository.NewUnitRepository()
	redis := repository.NewRedisRepository()

	connUseCase := usecase.New(unit, redis, h)
	wsHandler := handler.NewWebSocketHandler(auth, connUseCase)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(logger.SlogMiddleware())

	// 👇 WebSocket endpoint
	router.GET("/ws", wsHandler.HandleConnection)

	// 👇 Prometheus endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	srv := &http.Server{
		Addr:         cfg.AppHost + ":" + cfg.AppPort,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	// 👇 отдельный HTTP сервер для pprof (чтобы не мешать основному)
	go func() {
		pprofAddr := ":6060"
		slog.Info("pprof available", slog.String("addr", pprofAddr))
		if err := http.ListenAndServe(pprofAddr, nil); err != nil && err != http.ErrServerClosed {
			slog.Error("pprof server failed", slog.String("error", err.Error()))
		}
	}()

	go func() {
		slog.Info("WebSocket server starting", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	<-stop
	slog.Info("Shutdown signal received")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		slog.Error("Graceful shutdown failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("Server gracefully stopped")
}
