package mq

import (
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

var logger = log.New(os.Stdout, "[Message Queue]: ", log.Lshortfile|log.LstdFlags)

var RoutingKeyCmdQueue string = "route-cmd-queue"
var RoutingKeyOutputQueue string = "route-output-queue"
var RoutingKeyErrorOutputQueue string = "route-error-queue"

var QueueNameCmd string = "cmd-queue"
var QueueNameOutput string = "output-queue"
var QueueNameError string = "error-queue"

var ExchangeName string = "x-conflow"

type Publisher struct {
	conn    *amqp.Connection // Connection to RabbitMQ server
	channel *amqp.Channel    // Channel for publishing messages

	queueName    string
	exchangeName string
}

type Consumer struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	publisher *Publisher

	exchangeName string
	tag          string // Consumer tag for message acknowledgment
}

type ConsumerParams struct {
	QueueRoutingInfo map[string]string // Stores key: queue name, val: routing key
}
