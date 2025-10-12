package mq

import (
	"fmt"

	pb "github.com/ImTheCurse/ConflowCI/internal/mq/pb"
)

var ConsumerInstance *Consumer = nil

func getConsumer(amqpURL string, exchangeName string, params ConsumerParams, tag string) (*Consumer, error) {
	if ConsumerInstance == nil {
		c, err := NewConsumer(amqpURL, exchangeName, params, tag)
		if err != nil {
			return nil, err
		}
		ConsumerInstance = c
	}
	return ConsumerInstance, nil
}

func (s *ConsumerServer) StartConsumer(req *pb.ConsumerCommandRequest,
	stream pb.ConsumerServicer_StartConsumerServer) error {

	params := ConsumerParams{
		QueueRoutingInfo: req.Params,
	}
	consumer, err := getConsumer(req.MqUrl, req.Exchange, params, req.Tag)
	if err != nil {
		e := fmt.Sprintf("Failed to get message queue by consumer, got err : %s", err.Error())
		stream.Send(&pb.ConsumerCommandResponse{Error: &pb.ConsumerError{Reason: e}})
		return fmt.Errorf(e)
	}
	err = consumer.ConsumeCommand(stream.Context(), stream)
	if err != nil {
		e := fmt.Sprintf("Failed to consume commands, got: %s", err)
		stream.Send(&pb.ConsumerCommandResponse{Error: &pb.ConsumerError{Reason: e}})
		return fmt.Errorf(e)
	}
	return nil
}
