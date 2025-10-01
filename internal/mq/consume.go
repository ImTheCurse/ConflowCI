package mq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

// NewConsumer sets up a consumer bound to an exchange and queue
func NewConsumer(amqpURL string, exchangeName string, params ConsumerParams, tag string) (*Consumer, error) {
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

	for queueName, routingKey := range params.QueueRoutingInfo {
		// Declare queue
		q, err := ch.QueueDeclare(
			queueName,
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			ch.Close()
			conn.Close()
			return nil, QueueError{msg: err.Error()}
		}
		// Bind queue to exchange
		err = ch.QueueBind(
			q.Name,
			routingKey,
			exchangeName,
			false,
			nil,
		)
		if err != nil {
			ch.Close()
			conn.Close()
			return nil, BindingError{msg: err.Error()}
		}
	}

	return &Consumer{
		conn:         conn,
		channel:      ch,
		exchangeName: exchangeName,
		tag:          tag,
	}, nil
}

// Consume starts consuming
func (c *Consumer) Consume(ctx context.Context, queue string, handler func([]byte) error) error {
	msgs, err := c.channel.Consume(
		queue,
		c.tag,
		false, // manual ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for {
		select {
		case d, ok := <-msgs:
			if !ok {
				return nil
			}
			if err := handler(d.Body); err != nil {
				_ = d.Nack(false, true) // requeue on error
				continue
			}
			_ = d.Ack(false)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
func (c *Consumer) Close() {
	if c.channel != nil {
		_ = c.channel.Close()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
}
