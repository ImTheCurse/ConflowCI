package mq

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestMessageQueue(t *testing.T) {
	ctx := context.Background()
	logger.Printf("Creating Container RabbitMQ")
	c, connURI, err := CreateMessageQueueContainer()
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer c.Terminate(ctx)
	logger.Printf("Container RabbitMQ created")

	logger.Printf("Creating Publisher")

	// retry connection a few times
	var p *Publisher
	for _ = range 5 {
		p, err = NewPublisher(connURI, ExchangeName)
		if err == nil {
			break
		}
		logger.Printf("Message queue not ready, retrying... (%v)", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		t.Fatalf("Failed to create publisher: %v", err)
	}

	logger.Printf("Creating Consumer")
	params := ConsumerParams{
		QueueRoutingInfo: map[string]string{
			QueueNameCmd:    RoutingKeyCmdQueue,
			QueueNameOutput: RoutingKeyOutputQueue,
			QueueNameError:  RoutingKeyErrorOutputQueue,
		},
	}
	consumer, err := NewConsumer(connURI, ExchangeName, params, "test-consumer")
	if err != nil {
		t.Errorf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	m := "hello-world!"
	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCtx()

	done := make(chan error, 1)
	handler := func(msg []byte) error {
		recievedMsg := string(msg)
		if recievedMsg == m {
			logger.Printf("Received message: %s", recievedMsg)
			done <- nil
			return nil
		}
		done <- fmt.Errorf("Expected message: %s got: %s", m, recievedMsg)
		return nil
	}
	logger.Printf("Starting Consumption")
	go consumer.Consume(ctx, QueueNameOutput, handler)
	if err := p.Publish(ctx, RoutingKeyOutputQueue, []byte(m)); err != nil {
		t.Errorf("Failed to publish message: %v", err)
	}
	logger.Printf("Published message")
	// Wait for handler or timeout
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Failed to receive message: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting for message")
	}
}
