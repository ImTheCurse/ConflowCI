package mq

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func CreateMessageQueueContainer() (testcontainers.Container, string, error) {
	ctx := context.Background()

	// Start RabbitMQ container
	req := testcontainers.ContainerRequest{
		Image:        "rabbitmq:3.12-management",
		ExposedPorts: []string{"5672/tcp", "14351/tcp"},
		WaitingFor: wait.ForListeningPort("5672/tcp").
			WithStartupTimeout(30 * time.Second),
	}
	rmqC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		return nil, "", fmt.Errorf("failed to start container: %v", err)
	}

	host, err := rmqC.Host(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get container host: %v", err)
	}
	port, err := rmqC.MappedPort(ctx, "5672/tcp")
	if err != nil {
		return nil, "", fmt.Errorf("failed to get mapped port: %v", err)
	}
	amqpURL := "amqp://guest:guest@" + host + ":" + port.Port() + "/"
	return rmqC, amqpURL, nil

}
