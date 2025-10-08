package mq

import (
	"fmt"
)

type ConnectionError struct {
	msg string
}

func (e ConnectionError) Error() string {
	return fmt.Sprintf("Message queue: connection error: %s", e.msg)
}

type ChannelError struct {
	msg string
}

func (e ChannelError) Error() string {
	return fmt.Sprintf("Message queue: channel error: %s", e.msg)
}

type ExchangeError struct {
	msg string
}

func (e ExchangeError) Error() string {
	return fmt.Sprintf("Message queue: exchange error: %s", e.msg)
}

type QueueError struct {
	msg string
}

func (e QueueError) Error() string {
	return fmt.Sprintf("Message queue: queue error: %s", e.msg)
}

type BindingError struct {
	msg string
}

func (e BindingError) Error() string {
	return fmt.Sprintf("Message queue: binding error: %s", e.msg)
}
