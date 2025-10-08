package sync

import (
	"context"
	"os"
	"sync"

	"github.com/ImTheCurse/ConflowCI/internal/mq"
	"github.com/ImTheCurse/ConflowCI/pkg/ssh"
	goSSH "golang.org/x/crypto/ssh"
)

// RunTaskOnAllMachines distributes tasks across all endpoints
func (te *TaskExecutor) RunTaskOnAllMachines() error {
	uri := os.Getenv("CONFLOW_MQ_URI")

	te.State = RunningTask
	logger.Printf("%s: with id: %s", te.State.String(), te.TaskID)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(len(te.Cmds))

	params := mq.ConsumerParams{
		QueueRoutingInfo: map[string]string{
			mq.QueueNameCmd:    mq.RoutingKeyCmdQueue,
			mq.QueueNameOutput: mq.RoutingKeyOutputQueue,
			mq.QueueNameError:  mq.RoutingKeyErrorOutputQueue,
		},
	}

	// start all consumer goroutines
	for _, ep := range te.RunsOn {
		logger.Printf("Creating consumer for endpoint: %s", ep.Name)

		consumer, err := mq.NewConsumer(uri, mq.ExchangeName, params, ep.Name)
		if err != nil {
			logger.Printf("Error creating consumer: %v", err)
			continue
		}
		defer consumer.Close()

		sshCfg := ssh.SSHConnConfig{
			Username:       ep.User,
			PrivateKeyPath: ep.PrivateKeyPath,
		}

		cfg, err := sshCfg.BuildConfig()
		if err != nil {
			logger.Printf("Error building SSH config: %v", err)
			continue
		}

		conn, err := ssh.NewSSHConn(ep, cfg)
		if err != nil {
			logger.Printf("Failed to create SSH connection: %v", err)
			break
		}
		defer conn.Close()

		go consumer.ConsumeCommand(ctx, &wg, conn, runCommandOnRemoteMachine)
	}

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

// runCommandOnRemoteMachine executes a command on the endpoint's machine.
func runCommandOnRemoteMachine(cmd []byte, s *goSSH.Session) (string, error) {
	b, err := s.CombinedOutput(string(cmd))
	output := string(b)
	logger.Printf("Executed command: %s. got output: %s", string(cmd), output)
	return output, err
}
