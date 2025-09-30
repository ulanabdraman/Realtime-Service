package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"RealtimeService/config"
	"RealtimeService/internal/domains/connection/handler"
	"RealtimeService/internal/domains/connection/repository"
	"RealtimeService/internal/domains/connection/usecase"
	"RealtimeService/internal/server/hub"
	stream "RealtimeService/internal/stream/usecase"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Загрузка конфигурации
	cfg := config.Load()

	// 2. Контекст + shutdown-ловушка
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 3. Ловим Ctrl+C или SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// 4. Инициализация хаба
	h := hub.New()

	// 5. Запуск Kafka consumer
	kafka := stream.NewKafkaConsumer(h, cfg.KafkaTopic, cfg.KafkaGroup, cfg.KafkaBrokers)
	if err := kafka.Start(ctx); err != nil {
		log.Fatalf("Kafka error: %v", err)
	}

	// 6. Репозитории
	auth := repository.NewAuthRepository()
	unit := repository.NewUnitRepository()
	redis := repository.NewRedisRepository()

	// 7. UseCase + Handler
	connUseCase := usecase.New(unit, redis, h)
	wsHandler := handler.NewWebSocketHandler(auth, connUseCase)

	// 8. Gin и http.Server
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/ws", wsHandler.HandleConnection)

	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// 9. Запуск в отдельной горутине
	go func() {
		log.Printf("🚀 WebSocket server started at :%s", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// 10. Ждём сигнала завершения
	<-stop
	log.Println("🛑 Caught shutdown signal, exiting...")

	// 11. Shutdown с таймаутом
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("Server forced shutdown: %v", err)
	}

	log.Println("✅ Graceful shutdown complete")
}
