package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/vidarnilsson/vinlaro-chat/internal/model"
)

type Consumer struct {
	client *kgo.Client
}

func NewConsumer(brokers []string, topic, groupID string) (*Consumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(topic),
	)
	if err != nil {
		return nil, fmt.Errorf("kafka consumer: %w", err)
	}
	return &Consumer{client: client}, nil
}

// Consume polls Kafka in a loop, calling handler for each MessageEvent.
// Returns when ctx is cancelled.
func (c *Consumer) Consume(ctx context.Context, handler func(model.MessageEvent)) {
	for {
		fetches := c.client.PollFetches(ctx)
		if fetches.IsClientClosed() || ctx.Err() != nil {
			return
		}

		fetches.EachError(func(t string, p int32, err error) {
			log.Printf("kafka fetch error topic=%s partition=%d: %v", t, p, err)
		})

		fetches.EachRecord(func(record *kgo.Record) {
			var event model.MessageEvent
			if err := json.Unmarshal(record.Value, &event); err != nil {
				log.Printf("kafka: failed to unmarshal message event: %v", err)
				return
			}
			handler(event)
		})
	}
}

func (c *Consumer) Close() {
	c.client.Close()
}
