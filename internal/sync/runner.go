package sync

import (
	"github.com/ImTheCurse/ConflowCI/pkg/config"
	"github.com/ImTheCurse/ConflowCI/pkg/ssh"
	goSSH "golang.org/x/crypto/ssh"
)

func (te *TaskExecutor) RunTaskOnAllMachines() []error {
	te.State = RunningTask
	logger.Printf("%s: with id: %s", te.State.String(), te.TaskID)
	done := make(chan bool, len(te.RunsOn))

	for _, ep := range te.RunsOn {
		go commandConsumer(&ep, &te.State, te.CmdQueue, te.OutputQueue, te.ErrorQueue, done)
	}

	// Waiting for all commands consumers to finish execution
	for i := 0; i < len(te.RunsOn); i++ {
		<-done
	}
	close(te.OutputQueue)
	close(te.ErrorQueue)

	if len(te.ErrorQueue) != 0 {
		errors := []error{}
		for e := range te.ErrorQueue {
			errors = append(errors, e)
		}
		return errors
	}
	return []error{}
}

func commandConsumer(ep *config.EndpointInfo, taskState *TaskState,
	cmdQueue chan string, outputQueue chan<- string, errorQueue chan<- error, done chan<- bool) {

	logger.Printf("Starting command consumer for endpoint: %s - %s", ep.Name, ep.Host)
	for cmd := range cmdQueue {
		sshCfg := ssh.SSHConnConfig{
			Username:       ep.User,
			PrivateKeyPath: ep.PrivateKeyPath,
		}

		cfg, err := sshCfg.BuildConfig()
		if err != nil {
			// this isn't a great implementation since
			// it could cause a race condition.
			// TODO: implement this using message queue acks.
			logger.Printf("Error: %v |  %s - %s, adding command back to queue | command: %s", err, ep.Name, ep.Host, cmd)
			cmdQueue <- cmd
			break
		}

		conn, err := ssh.NewSSHConn(*ep, cfg)
		if err != nil {
			logger.Printf("Failed to create SSH connection: %v", err)
			break
		}
		defer conn.Close()
		s, err := conn.NewSession()
		if err != nil {
			logger.Printf("Failed to create session for command: %s", cmd)
			break
		}
		defer s.Close()
		o, err := runCommandOnRemoteMachine(s, ep, cmd)

		if err != nil {
			logger.Printf("Error running command: %v", err)
			errorQueue <- err
			// no need to worry about race conditions since we only
			// set the taskState to ErrorInTask.
			// other states are only applied after all goroutines are finished.
			*taskState = ErrorInTask
		}
		outputQueue <- o
		logger.Printf("finshed execution on %s - %s | command: %s", ep.Name, ep.Host, cmd)
	}
	done <- true
}

func runCommandOnRemoteMachine(s *goSSH.Session, ep *config.EndpointInfo, cmd string) (string, error) {
	b, err := s.CombinedOutput(cmd)
	output := string(b)
	return output, err
}
