package sync

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ImTheCurse/ConflowCI/pkg/config"
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
)

func NewTaskExecutor(conn *ssh.Client, cfg config.ValidatedConfig, task config.TaskConsumerJobs) (*TaskExecutor, error) {
	files := []string{}
	var err error
	if task.File == nil {
		files, err = getFilesByRegex(conn, task.Pattern, buildPath)
		if err != nil {
			return nil, err
		}
	} else {
		files = task.File
	}
	ch := make(chan string, len(files))
	for _, cmd := range task.Commands {
		for _, file := range files {
			expandedCmd := strings.ReplaceAll(cmd, "{file}", file)
			ch <- expandedCmd
		}
	}
	close(ch)
	return &TaskExecutor{
		TaskID:      uuid.New(),
		State:       StartingTask,
		RunsOn:      getTasksMachine(cfg, task),
		Files:       files,
		CmdQueue:    ch,
		OutputQueue: make(chan string, 10_000),
		ErrorQueue:  make(chan error, 10_000),
	}, err

}

func getTasksMachine(cfg config.ValidatedConfig, task config.TaskConsumerJobs) []config.EndpointInfo {
	res := []config.EndpointInfo{}
	for _, ep := range cfg.Endpoints {
		if slices.Contains(task.RunsOn, ep.Name) {
			res = append(res, ep)
		}
	}
	return res
}

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
