package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/pubsub"

	"github.com/joaovitorpedrosomonteiro/low-carbon-identity-service/internal/domain/user"
)

type Publisher struct {
	client *pubsub.Client
	topic  *pubsub.Topic
}

func NewPublisher(ctx context.Context) (*Publisher, error) {
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		projectID = "low-carbon-491109"
	}

	topicID := os.Getenv("PUBSUB_TOPIC_ID")
	if topicID == "" {
		topicID = "identity-events"
	}

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("pubsub.NewClient: %w", err)
	}

	topic := client.Topic(topicID)
	return &Publisher{client: client, topic: topic}, nil
}

func (p *Publisher) Publish(ctx context.Context, event user.DomainEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	result := p.topic.Publish(ctx, &pubsub.Message{
		Data: data,
		Attributes: map[string]string{
			"event_type":    event.EventType,
			"schema_version": event.SchemaVer,
		},
	})

	id, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("pubsub.Publish.Get: %w", err)
	}

	log.Printf("Published event %s (type: %s) with message ID: %s", event.EventID, event.EventType, id)
	return nil
}

func (p *Publisher) Close() error {
	return p.client.Close()
}

type MockPublisher struct{}

func NewMockPublisher() *MockPublisher {
	return &MockPublisher{}
}

func (m *MockPublisher) Publish(ctx context.Context, event user.DomainEvent) error {
	data, _ := json.Marshal(event)
	log.Printf("[MockPublisher] Event: %s", string(data))
	return nil
}
