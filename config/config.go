package config

import (
	"log/slog"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppPort      string
	AppHost      string
	KafkaTopic   string
	KafkaGroup   string
	KafkaBrokers []string
	JWTSecret    string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func Load() *Config {
	viper.SetConfigName("config") // config.yaml
	viper.AddConfigPath("./config")
	viper.SetConfigType("yaml")

	// Значения по умолчанию
	viper.SetDefault("AppPort", "8080")
	viper.SetDefault("AppHost", "localhost")
	viper.SetDefault("KafkaTopic", "realtime")
	viper.SetDefault("KafkaGroup", "realtime-consumer")
	viper.SetDefault("KafkaBrokers", []string{"localhost:9092"})
	viper.SetDefault("JWTSecret", "supersecret")
	viper.SetDefault("ReadTimeout", "15s")
	viper.SetDefault("WriteTimeout", "15s")

	if err := viper.ReadInConfig(); err != nil {
		slog.Warn("No config file found, using defaults", slog.String("error", err.Error()))
	} else {
		slog.Info("Config file loaded", slog.String("path", viper.ConfigFileUsed()))
	}

	cfg := &Config{
		AppPort:      viper.GetString("AppPort"),
		AppHost:      viper.GetString("AppHost"),
		KafkaTopic:   viper.GetString("KafkaTopic"),
		KafkaGroup:   viper.GetString("KafkaGroup"),
		KafkaBrokers: viper.GetStringSlice("KafkaBrokers"),
		JWTSecret:    viper.GetString("JWTSecret"),
		ReadTimeout:  viper.GetDuration("ReadTimeout"),
		WriteTimeout: viper.GetDuration("WriteTimeout"),
	}

	slog.Info("Config loaded",
		slog.String("app_host", cfg.AppHost),
		slog.String("app_port", cfg.AppPort),
		slog.String("kafka_topic", cfg.KafkaTopic),
		slog.String("kafka_group", cfg.KafkaGroup),
		slog.Any("kafka_brokers", cfg.KafkaBrokers),
		slog.Duration("read_timeout", cfg.ReadTimeout),
		slog.Duration("write_timeout", cfg.WriteTimeout),
	)

	return cfg
}
