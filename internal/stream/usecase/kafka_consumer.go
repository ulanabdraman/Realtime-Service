package usecase

import (
	"context"
	"encoding/json"
	"log/slog"

	"RealtimeService/internal/domains/connection/model"
	"RealtimeService/internal/server/hub"

	"github.com/IBM/sarama"
)

type KafkaConsumer struct {
	hub    *hub.Hub
	topic  string
	group  string
	addrs  []string
	logger *slog.Logger
}

func NewKafkaConsumer(h *hub.Hub, topic, group string, brokers []string) *KafkaConsumer {
	logger := slog.With(
		slog.String("object", "stream"),
		slog.String("layer", "kafka_consumer"),
	)

	return &KafkaConsumer{
		hub:    h,
		topic:  topic,
		group:  group,
		addrs:  brokers,
		logger: logger,
	}
}

func (kc *KafkaConsumer) Start(ctx context.Context) error {
	kc.logger.Info("Starting Kafka consumer",
		slog.String("topic", kc.topic),
		slog.String("group", kc.group),
		slog.Any("brokers", kc.addrs),
	)

	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	group, err := sarama.NewConsumerGroup(kc.addrs, kc.group, config)
	if err != nil {
		kc.logger.Error("Failed to create Kafka consumer group", slog.String("error", err.Error()))
		return err
	}

	handler := &consumerGroupHandler{
		hub:    kc.hub,
		logger: kc.logger,
	}

	go func() {
		for {
			if err := group.Consume(ctx, []string{kc.topic}, handler); err != nil {
				kc.logger.Error("Kafka consume error", slog.String("error", err.Error()))
			}
			if ctx.Err() != nil {
				kc.logger.Info("Kafka consumer shutting down")
				return
			}
		}
	}()

	return nil
}

type consumerGroupHandler struct {
	hub    *hub.Hub
	logger *slog.Logger
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var records []model.Data

		if err := json.Unmarshal(msg.Value, &records); err != nil {
			// Пробуем как одиночный объект
			var single model.Data
			if err2 := json.Unmarshal(msg.Value, &single); err2 != nil {
				h.logger.Error("Invalid Kafka message",
					slog.String("error", err2.Error()),
					slog.String("raw", truncate(msg.Value, 300)),
				)
				continue
			}
			records = []model.Data{single}
		}

		for _, record := range records {
			unitID := int(record.ID)
			h.hub.Broadcast(unitID, record)

			h.logger.Debug("Broadcasted Kafka message",
				slog.Int("unit_id", unitID),
				slog.Any("short", truncateStruct(record, 3)),
			)
		}

		sess.MarkMessage(msg, "")
	}
	return nil
}

func truncate(data []byte, max int) string {
	if len(data) <= max {
		return string(data)
	}
	return string(data[:max]) + "..."
}

func truncateStruct(data model.Data, maxParams int) map[string]interface{} {
	// Вернёт сокращённый вид
	short := map[string]interface{}{
		"id":  data.ID,
		"pos": data.Pos,
	}
	// Подрежем params
	if len(data.Params) > 0 {
		shortParams := make(map[string]interface{})
		count := 0
		for k, v := range data.Params {
			if count >= maxParams {
				shortParams["..."] = "truncated"
				break
			}
			shortParams[k] = v
			count++
		}
		short["params"] = shortParams
	}
	return short
}
