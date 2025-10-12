package mq

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// NewPublisher creates a publisher that publishes to an exchange
func NewPublisher(amqpURL, exchangeName string) (*Publisher, error) {
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, ConnectionError{msg: err.Error()}
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, ChannelError{msg: err.Error()}
	}

	// Declare exchange
	err = ch.ExchangeDeclare(
		exchangeName,
		// direct exchange - binds routing key directly to a queue. you can read more here:
		// https://www.rabbitmq.com/tutorials/tutorial-four-go#direct-exchange
		"direct",
		// non-durable, we want to fail our test-pipeline if the message queue is down. we dont save the jobs.
		// we expect the pipeline the restart in entirety.
		// TODO: maybe in the future we could support long-term lived jobs and pipeline continuation.
		false,
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, ExchangeError{msg: err.Error()}
	}
	return &Publisher{conn: conn, channel: ch, exchangeName: exchangeName}, nil
}

// Publish sends a message to the exchange with the routing key
func (p *Publisher) Publish(ctx context.Context, routingKey string, body []byte) error {
	return p.PublishWithRetry(ctx, routingKey, body)

}

// publishWithRetry handles retries for unroutable messages
func (p *Publisher) PublishWithRetry(ctx context.Context, routingKey string, body []byte) error {
	const maxRetries = 10
	const initialBackoff = 500 * time.Millisecond

	// Channel to receive returned messages
	returns := p.channel.NotifyReturn(make(chan amqp.Return, 100))

	var lastErr error
	backoff := initialBackoff

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Publish with mandatory=true
		err := p.channel.PublishWithContext(
			ctx,
			p.exchangeName,
			routingKey,
			true,
			false,
			amqp.Publishing{
				ContentType:  "text/plain",
				DeliveryMode: amqp.Persistent,
				Body:         body,
				Timestamp:    time.Now(),
			},
		)
		if err != nil {
			lastErr = err
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		// Check if the message was returned as unroutable
		select {
		case ret := <-returns:
			logger.Printf("Message unroutable, retrying: routingKey=%s, body=%s. requeueing...", routingKey, string(ret.Body))
			lastErr = fmt.Errorf("message unroutable")
			time.Sleep(backoff)
			backoff *= 2
		case <-time.After(1 * time.Millisecond):
			// Message accepted, return success
			return nil
		}
	}

	return fmt.Errorf("failed to publish message after %d attempts: last error: %v", maxRetries, lastErr)
}

// Close cleans up resources
func (p *Publisher) Close() {
	if p.channel != nil {
		_ = p.channel.Close()
	}
	if p.conn != nil {
		_ = p.conn.Close()
	}
}
