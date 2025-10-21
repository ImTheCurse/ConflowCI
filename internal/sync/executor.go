package sync

import (
	"context"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	pb "github.com/ImTheCurse/ConflowCI/internal/sync/pb"
	"github.com/ImTheCurse/ConflowCI/pkg/config"
	"github.com/ImTheCurse/ConflowCI/pkg/grpc"
	"github.com/google/uuid"
)

// Creates a new task executor, the task executor is responsible for executing tasks on a remote machine
// it dispatches each cmd with file/pattern to a remote machine in a concurrent way using the RunTaskOnAllMachines func.
// there is no guarantee that the commands will be executed in the order they were dispatched.
func NewTaskExecutor(cfg config.ValidatedConfig, task config.TaskConsumerJobs, wsName string) (*TaskExecutor, error) {
	ctx := context.Background()
	files := []string{}
	var err error
	if task.File == nil {
		finder := pb.TaskFileFinder{
			Pattern:  task.Pattern,
			BuildDir: BuildPath,
		}
		endpoint := cfg.Endpoints[0]
		conn, err := grpc.CreateNewClientConnection(endpoint.GetEndpointURL())
		if err != nil {
			return nil, err
		}
		client := pb.NewFileExtractorClient(conn)

		f, err := client.GetFilesByRegex(ctx, &finder)
		if err != nil {
			return nil, err
		}
		files = f.Files
	} else {
		for _, file := range task.File {
			filesWithPath := filepath.Join(BuildPath, wsName, file)
			// filesWithPath := fmt.Sprintf("%s/%s", BuildPath, file)
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

func GetFilesByRegex(ctx context.Context, finder *pb.TaskFileFinder) (*pb.FileList, error) {
	cmd := exec.Command(
		"find",
		finder.BuildDir,
		"-regextype", "posix-extended",
		"-regex", finder.Pattern,
	)
	cmd.Dir = finder.BuildDir
	out, err := cmd.CombinedOutput()

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
	return &pb.FileList{Files: files}, nil
}
