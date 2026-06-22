package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/vidarnilsson/vinlaro-chat/internal/model"
)

type Producer struct {
	client *kgo.Client
	topic  string
}

func NewProducer(brokers []string, topic string) (*Producer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
	)
	if err != nil {
		return nil, fmt.Errorf("kafka producer: %w", err)
	}

	if err := ensureTopic(client, topic); err != nil {
		client.Close()
		return nil, err
	}

	return &Producer{client: client, topic: topic}, nil
}

// ensureTopic creates the topic if it doesn't already exist.
func ensureTopic(client *kgo.Client, topic string) error {
	admin := kadm.NewClient(client)
	resp, err := admin.CreateTopics(context.Background(), 1, 1, nil, topic)
	if err != nil {
		return fmt.Errorf("kafka create topic: %w", err)
	}
	for _, r := range resp {
		if r.Err != nil && r.Err != kerr.TopicAlreadyExists {
			return fmt.Errorf("kafka create topic %q: %w", r.Topic, r.Err)
		}
	}
	return nil
}

func (p *Producer) Publish(ctx context.Context, event model.MessageEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal message event: %w", err)
	}

	record := &kgo.Record{
		Topic: p.topic,
		Key:   []byte(event.ChannelID),
		Value: payload,
	}

	if err := p.client.ProduceSync(ctx, record).FirstErr(); err != nil {
		return fmt.Errorf("kafka produce: %w", err)
	}
	return nil
}

func (p *Producer) Close() {
	p.client.Close()
}
