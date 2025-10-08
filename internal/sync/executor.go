package sync

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ImTheCurse/ConflowCI/pkg/config"
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
)

// Creates a new task executor, the task executor is responsible for executing tasks on a remote machine
// it dispatches each cmd with file/pattern to a remote machine in a concurrent way using the RunTaskOnAllMachines func.
// there is no guarantee that the commands will be executed in the order they were dispatched.
func NewTaskExecutor(conn *ssh.Client, cfg config.ValidatedConfig, task config.TaskConsumerJobs) (*TaskExecutor, error) {
	files := []string{}
	var err error
	if task.File == nil {
		files, err = getFilesByRegex(conn, task.Pattern, BuildPath)
		if err != nil {
			return nil, err
		}
	} else {
		for _, file := range task.File {
			filesWithPath := fmt.Sprintf("%s/%s", BuildPath, file)
			files = append(files, filesWithPath)
		}
	}
	cmds := []string{}
	for _, cmd := range task.Commands {
		for _, file := range files {
			expandedCmd := strings.ReplaceAll(cmd, "{file}", file)
			cmds = append(cmds, expandedCmd)
		}
	}
	logger.Println("Added task commands.")
	logger.Println("Create TaskExecutor.")
	return &TaskExecutor{
		TaskID:  uuid.New(),
		State:   StartingTask,
		RunsOn:  getTasksMachine(cfg, task),
		Files:   files,
		Cmds:    cmds,
		Outputs: []string{},
		Errors:  []string{},
	}, err

}

// getTasksMachine returns the list of endpoints that the task should be executed on.
func getTasksMachine(cfg config.ValidatedConfig, task config.TaskConsumerJobs) []config.EndpointInfo {
	res := []config.EndpointInfo{}
	for _, ep := range cfg.Endpoints {
		if slices.Contains(task.RunsOn, ep.Name) {
			res = append(res, ep)
		}
	}
	return res
}

// getFilesByRegex returns the list of files that match the given regex expression.
// it uses posix-extended flavor of regex.
func getFilesByRegex(conn *ssh.Client, expr string, buildDir string) ([]string, error) {
	s, err := conn.NewSession()
	if err != nil {
		return nil, err
	}
	defer s.Close()

	cmd := fmt.Sprintf("find %s -regextype posix-extended -regex %q", buildDir, expr)
	out, err := s.CombinedOutput(cmd)
	if err != nil {
		logger.Printf("Error running getFilesByRegex, output: %s", string(out))
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	var files []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}
