package mq

import (
	"context"
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
	return p.channel.PublishWithContext(
		ctx,
		p.exchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "text/plain",
			DeliveryMode: amqp.Persistent,
			Body:         body,
			Timestamp:    time.Now(),
		},
	)
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
