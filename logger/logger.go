package logger

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"
	"log/slog"
	"os"
	"time"
)

func InitLogger() {
	_ = os.MkdirAll("logs", os.ModePerm)

	logWriter := &lumberjack.Logger{
		Filename:   "var/log/realtime/service.log",
		MaxAge:     1,
		MaxSize:    0,
		MaxBackups: 0,
		Compress:   true,
	}

	logger := slog.New(slog.NewJSONHandler(logWriter, nil))
	slog.SetDefault(logger)
}

func SlogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		slog.Info("HTTP request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("client_ip", c.ClientIP()),
		)
	}
}
