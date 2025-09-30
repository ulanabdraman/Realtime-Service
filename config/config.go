package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppPort      string
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
	viper.SetDefault("KafkaTopic", "realtime")
	viper.SetDefault("KafkaGroup", "realtime-consumer")
	viper.SetDefault("KafkaBrokers", []string{"localhost:9092"})
	viper.SetDefault("JWTSecret", "supersecret")
	viper.SetDefault("ReadTimeout", "15s")
	viper.SetDefault("WriteTimeout", "15s")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("No config file found, using defaults")
	}

	return &Config{
		AppPort:      viper.GetString("AppPort"),
		KafkaTopic:   viper.GetString("KafkaTopic"),
		KafkaGroup:   viper.GetString("KafkaGroup"),
		KafkaBrokers: viper.GetStringSlice("KafkaBrokers"),
		JWTSecret:    viper.GetString("JWTSecret"),
		ReadTimeout:  viper.GetDuration("ReadTimeout"),
		WriteTimeout: viper.GetDuration("WriteTimeout"),
	}
}
