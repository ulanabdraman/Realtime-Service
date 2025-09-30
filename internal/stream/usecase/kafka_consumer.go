package usecase

import (
	"context"
	"encoding/json"
	"log"

	"RealtimeService/internal/domains/connection/model"
	"RealtimeService/internal/server/hub"

	"github.com/IBM/sarama"
)

type KafkaConsumer struct {
	hub   *hub.Hub
	topic string
	group string
	addrs []string
}

func NewKafkaConsumer(h *hub.Hub, topic, group string, brokers []string) *KafkaConsumer {
	return &KafkaConsumer{
		hub:   h,
		topic: topic,
		group: group,
		addrs: brokers,
	}
}

func (kc *KafkaConsumer) Start(ctx context.Context) error {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	group, err := sarama.NewConsumerGroup(kc.addrs, kc.group, config)
	if err != nil {
		return err
	}

	handler := &consumerGroupHandler{
		hub: kc.hub,
	}

	go func() {
		for {
			if err := group.Consume(ctx, []string{kc.topic}, handler); err != nil {
				log.Printf("Kafka consume error: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	return nil
}

// consumerGroupHandler implements sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	hub *hub.Hub
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var records []model.Data

		// Пробуем распарсить как массив
		if err := json.Unmarshal(msg.Value, &records); err != nil {
			// Если не массив — пробуем как одиночный объект
			var single model.Data
			if err2 := json.Unmarshal(msg.Value, &single); err2 != nil {
				log.Printf("❌ Invalid Kafka message: %v\nRaw: %s", err2, string(msg.Value))
				continue
			}

			// Оборачиваем в массив
			records = []model.Data{single}
		}

		// Обработка каждого record
		for _, record := range records {
			unitID := int(record.ID)
			h.hub.Broadcast(unitID, record)
		}

		// Подтверждаем, что сообщение прочитано
		sess.MarkMessage(msg, "")
	}
	return nil
}
