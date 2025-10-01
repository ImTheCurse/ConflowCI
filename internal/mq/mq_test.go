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

func TestNewConsumerErrors(t *testing.T) {
	ctx := context.Background()
	logger.Printf("Creating Container RabbitMQ")
	c, connURI, err := CreateMessageQueueContainer()
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer c.Terminate(ctx)
	logger.Printf("Container RabbitMQ created")

	params := ConsumerParams{
		QueueRoutingInfo: map[string]string{
			QueueNameCmd:    RoutingKeyCmdQueue,
			QueueNameOutput: RoutingKeyOutputQueue,
			QueueNameError:  RoutingKeyErrorOutputQueue,
		},
	}
	tests := []struct {
		name        string
		amqpURL     string
		exchange    string
		queueInfo   ConsumerParams
		wantErrType any
	}{
		{
			name:        "Connection error",
			amqpURL:     "amqp://invalid:5672", // invalid URL to force connection failure
			exchange:    "test-ex",
			queueInfo:   params,
			wantErrType: ConnectionError{},
		},
		{
			name:        "Exchange error",
			amqpURL:     connURI,
			exchange:    "", // invalid name triggers exchange declare error
			queueInfo:   params,
			wantErrType: ExchangeError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewConsumer(tt.amqpURL, tt.exchange, tt.queueInfo, "tag")
			if err == nil {
				t.Fatalf("expected error, got nil")
			}

			// check error type using type assertion
			switch tt.wantErrType.(type) {
			case ConnectionError:
				if _, ok := err.(ConnectionError); !ok {
					t.Errorf("expected ConnectionError, got %T", err)
				}
			case ExchangeError:
				if _, ok := err.(ExchangeError); !ok {
					t.Errorf("expected ExchangeError, got %T", err)
				}
			case QueueError:
				if _, ok := err.(QueueError); !ok {
					t.Errorf("expected QueueError, got %T", err)
				}
			case BindingError:
				if _, ok := err.(BindingError); !ok {
					t.Errorf("expected BindingError, got %T", err)
				}
			}
		})
	}
}

func TestNewPublisherErrors(t *testing.T) {
	ctx := context.Background()
	logger.Printf("Creating Container RabbitMQ")
	c, connURI, err := CreateMessageQueueContainer()
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer c.Terminate(ctx)
	logger.Printf("Container RabbitMQ created")
	tests := []struct {
		name        string
		amqpURL     string
		exchange    string
		wantErrType any
	}{
		{
			name:        "Connection error",
			amqpURL:     "amqp://invalid:5672", // invalid URL to force connection failure
			exchange:    "test-ex",
			wantErrType: ConnectionError{},
		},
		{
			name:        "Exchange error",
			amqpURL:     connURI,
			exchange:    "", // invalid exchange name triggers error
			wantErrType: ExchangeError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPublisher(tt.amqpURL, tt.exchange)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}

			// check error type using type assertion
			switch tt.wantErrType.(type) {
			case ConnectionError:
				if _, ok := err.(ConnectionError); !ok {
					t.Errorf("expected ConnectionError, got %T", err)
				}
			case ExchangeError:
				if _, ok := err.(ExchangeError); !ok {
					t.Errorf("expected ExchangeError, got %T", err)
				}
			}
		})
	}
}
