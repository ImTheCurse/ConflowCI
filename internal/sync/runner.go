package sync

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/ImTheCurse/ConflowCI/internal/mq"
	mqpb "github.com/ImTheCurse/ConflowCI/internal/mq/pb"
	"github.com/ImTheCurse/ConflowCI/pkg/grpc"
)

// RunTaskOnAllMachines distributes tasks across all endpoints
func (te *TaskExecutor) RunTaskOnAllMachines() error {
	uri := os.Getenv("CONFLOW_MQ_URI")

	te.State = RunningTask
	logger.Printf("%s: with id: %s", te.State.String(), te.TaskID)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(len(te.Cmds))

	params := mq.ConsumerParams{
		QueueRoutingInfo: map[string]string{
			mq.QueueNameCmd:    mq.RoutingKeyCmdQueue,
			mq.QueueNameOutput: mq.RoutingKeyOutputQueue,
			mq.QueueNameError:  mq.RoutingKeyErrorOutputQueue,
		},
	}

	var consumersReady sync.WaitGroup
	consumersReady.Add(len(te.RunsOn))
	// start all consumer goroutines
	for _, ep := range te.RunsOn {
		logger.Printf("Creating consumer for endpoint: %s", ep.Name)

		go func() {
			conn, err := grpc.CreateNewClientConnection(ep.GetEndpointURL())
			if err != nil {
				logger.Printf("Error creating gRPC connection: %v", err)
				return
			}

			client := mqpb.NewConsumerServicerClient(conn)
			req := mqpb.ConsumerCommandRequest{
				MqUrl:    uri,
				Exchange: mq.ExchangeName,
				Params:   params.QueueRoutingInfo,
				Tag:      ep.Name,
			}
			stream, err := client.StartConsumer(ctx, &req)
			if err != nil {
				logger.Printf("Error starting consumer: %v", err)
				return
			}

			consumersReady.Done()
			logger.Printf("Consumer ready for endpoint: %s", ep.Name)

			// we recieve the client's stream right away so client.StartConsumer isn't blocking.
			for {
				msg, err := stream.Recv()
				if err != nil {
					logger.Printf("Error receiving message: %v", err)
					return
				}

				if msg.FinishedCommand != nil {
					logger.Println("Command finished")
					wg.Done()
				}

				if msg.Error != nil {
					logger.Printf("Error received from remote machine: %v", msg.Error)
				}
				logger.Printf("Output received from remote machine: %v", msg.Output)
			}
		}()
	}

	consumersReady.Wait()
	time.Sleep(3 * time.Second)
	for i, cmd := range te.Cmds {
		p, err := mq.NewPublisher(uri, mq.ExchangeName)
		if err != nil {
			logger.Printf("Error creating publisher: %v", err)
			continue
		}
		defer p.Close()

		p.Publish(ctx, mq.RoutingKeyCmdQueue, []byte(cmd))
		logger.Printf("Published command: %s", cmd)
		logger.Printf("%v commands remaining to publish.", len(te.Cmds)-i-1)
	}

	wg.Wait()
	cancel()

	logger.Println("Getting command outputs and errors...")

	outputConsumer, err := mq.NewConsumer(uri, mq.ExchangeName, params, "output-consumer")
	if err != nil {
		logger.Printf("Error creating output consumer: %v", err)
		return err
	}
	defer outputConsumer.Close()

	errorConsumer, err := mq.NewConsumer(uri, mq.ExchangeName, params, "error-consumer")
	if err != nil {
		logger.Printf("Error creating output consumer: %v", err)
		return err
	}
	defer errorConsumer.Close()

	done := make(chan struct{})

	var cmdResWg sync.WaitGroup
	cmdResWg.Add(len(te.Cmds))
	go func() {
		// wait until all commands outputs/errors finish reading from queue
		cmdResWg.Wait()
		close(done)
	}()
	var errorsRes, outputsRes []string = []string{}, []string{}
	var errorsErr, outputsError error = nil, nil

	go errorConsumer.ConsumeQueueContents(&cmdResWg, done, mq.QueueNameError, &errorsRes, &errorsErr)
	go outputConsumer.ConsumeQueueContents(&cmdResWg, done, mq.QueueNameOutput, &outputsRes, &outputsError)

	cmdResWg.Wait()

	if errorsErr != nil {
		logger.Printf("Error consuming error queue contents: %v", err)
		return err
	}
	if outputsError != nil {
		logger.Printf("Error consuming output queue contents: %v", err)
		return err
	}

	if len(errorsRes) > 0 {
		te.State = CompleteTaskWithErrors
		logger.Printf("%s: with id: %s", te.State.String(), te.TaskID)
	}

	te.Outputs = outputsRes
	te.Errors = errorsRes
	return err
}
