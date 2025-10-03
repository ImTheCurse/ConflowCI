package mq

import (
	"context"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"golang.org/x/crypto/ssh"
)

// NewConsumer sets up a consumer bound to an exchange and queue
func NewConsumer(amqpURL string, exchangeName string, params ConsumerParams, tag string) (*Consumer, error) {
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

	// we create a publisher in order to send the output / error back to the message queue.
	p, err := NewPublisher(amqpURL, exchangeName)
	if err != nil {
		return nil, err
	}
	logger.Println("Created mq.Consumer Instance.")
	return &Consumer{
		conn:         conn,
		channel:      ch,
		publisher:    p,
		exchangeName: exchangeName,
		tag:          tag,
	}, nil
}

// ConsumeCommand starts consuming
func (c *Consumer) ConsumeCommand(ctx context.Context, wg *sync.WaitGroup,
	client *ssh.Client, handler func([]byte, *ssh.Session) (string, error)) error {
	logger.Println("Consuming command queue.")
	msgs, err := c.channel.Consume(
		QueueNameCmd,
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
			s, err := client.NewSession()
			if err != nil {
				_ = d.Nack(false, true) // requeue on error
				continue
			}
			defer s.Close()
			o, err := handler(d.Body, s) // error here is a cmd error
			if err != nil {
				err = c.publisher.Publish(ctx, RoutingKeyErrorOutputQueue, []byte(o)) // send message to error queue if the cmd failed.
				if err != nil {
					logger.Printf("failed to publish error message: %v, with error: %v", o, err)
				}
			} else {
				err = c.publisher.Publish(ctx, RoutingKeyOutputQueue, []byte(o)) // send output with no error to output queue.
				if err != nil {
					logger.Printf("failed to publish message: %v, with error: %v", o, err)
				}
			}
			wg.Done()
			_ = d.Ack(false)
		case <-ctx.Done():
			return nil
		}
	}
}

func (c *Consumer) ConsumeQueueContents(wg *sync.WaitGroup, done chan struct{}, queueName string, buf *[]string, e *error) {
	msgs, err := c.channel.Consume(
		queueName,
		c.tag,
		false, // manual ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		*e = err
		return
	}

	for {
		select {
		case d, ok := <-msgs:
			if ok {
				*buf = append(*buf, string(d.Body))
				wg.Done()
				_ = d.Ack(false)
			}
		case <-done:
			return
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
	c.publisher.Close()
}
