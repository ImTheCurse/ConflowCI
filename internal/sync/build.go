package sync

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	providerPB "github.com/ImTheCurse/ConflowCI/internal/provider/pb"
	syncPB "github.com/ImTheCurse/ConflowCI/internal/sync/pb"
	"github.com/ImTheCurse/ConflowCI/pkg/config"
	"github.com/ImTheCurse/ConflowCI/pkg/grpc"
	"github.com/google/uuid"
	"github.com/pelletier/go-toml"
)

func NewWorkerBuilder(cfg config.ValidatedConfig, remote, branch, branchRef string) *WorkersBuilder {
	return &WorkersBuilder{
		Name:       cfg.Pipeline.Build.Name,
		BuildID:    uuid.New(),
		State:      StartingBuild,
		RunsOn:     cfg.Endpoints,
		Steps:      cfg.Pipeline.Build.BuildSteps,
		CloneURL:   cfg.GetCloneURL(),
		Remote:     remote,
		BranchName: branch,
		Token:      cfg.GetToken(),
		BranchRef:  branchRef,
	}
}

func NewWorkerBuilderServer(client providerPB.RepositoryProviderClient) *WorkerBuilderServer {
	return &WorkerBuilderServer{provider: client}
}

func (s *WorkerBuilderServer) BuildRepo(cfg *syncPB.WorkerConfig) *syncPB.WorkerBuildOutput {
	ctx := context.Background()
	repoWithBranch := fmt.Sprintf("%s-%s", cfg.Req.Name, cfg.Req.BranchName)
	path := filepath.Join("..", repoWithBranch)
	dir := filepath.Join(os.ExpandEnv(BuildPath), cfg.Req.Name)

	err := s.syncRepository(cfg)
	if err != nil {
		e := fmt.Sprintf("Error syncing repository: %s", err.Error())
		return &syncPB.WorkerBuildOutput{WorkerName: cfg.WorkerName, Error: &syncPB.WorkerBuildError{Error: e}}
	}

	branchName := strings.Split(cfg.Req.BranchRef, ":")[1]
	wrkTreeReq := providerPB.WorkTreeRequest{
		Name:            cfg.WorkerName,
		RepoDir:         dir,
		BranchName:      branchName,
		WorktreeRelPath: "../" + cfg.Req.Name + "-" + cfg.Req.BranchName,
	}
	resp, err := s.provider.CreateWorkTree(ctx, &wrkTreeReq)
	if resp.Error != nil || err != nil {
		e := GetProtoWorkerError("Error creating work tree", err, resp)
		return &syncPB.WorkerBuildOutput{WorkerName: cfg.WorkerName, Error: &syncPB.WorkerBuildError{Error: e}}
	}

	cdToWrkTree := "cd " + filepath.Join(dir, path)
	// Trim and filter steps
	var steps []string
	for _, step := range cfg.BuildSteps {
		s := strings.TrimSpace(step)
		if s != "" {
			steps = append(steps, s)
		}
	}

	var cmd string
	if len(steps) != 0 {
		buildCmd := strings.Join(steps, " && ")
		cmd = cdToWrkTree + " && " + buildCmd
	} else {
		cmd = cdToWrkTree
	}
	logger.Printf("Executing command: %s", cmd)

	c := exec.Command("bash", "-c", cmd)
	b, err := c.CombinedOutput()
	if err != nil {
		e := GetProtoWorkerError("Error running commands", err, resp)
		return &syncPB.WorkerBuildOutput{WorkerName: cfg.WorkerName, Output: string(b), Error: &syncPB.WorkerBuildError{Error: e}}
	}
	resp, err = s.provider.RemoveWorkTree(ctx, &wrkTreeReq)
	if err != nil {
		e := GetProtoWorkerError("Error removing work tree", err, resp)
		return &syncPB.WorkerBuildOutput{WorkerName: cfg.WorkerName, Error: &syncPB.WorkerBuildError{Error: e}}
	}
	return &syncPB.WorkerBuildOutput{WorkerName: cfg.WorkerName, Output: string(b)}
}

func (wb *WorkersBuilder) BuildAllEndpoints() []*syncPB.WorkerBuildOutput {
	outputs := []*syncPB.WorkerBuildOutput{}
	dir := filepath.Join(os.ExpandEnv(BuildPath), wb.Name)

	var wg sync.WaitGroup
	var mu sync.Mutex
	wg.Add(len(wb.RunsOn))
	appendToOutput := func(output *syncPB.WorkerBuildOutput) {
		mu.Lock()
		outputs = append(outputs, output)
		mu.Unlock()
	}

	for _, ep := range wb.RunsOn {
		var addr string
		if ep.Port == 0 {
			addr = ep.Host
		} else {
			addr = fmt.Sprintf("%s:%d", ep.Host, ep.Port)
		}
		go func() {
			defer wg.Done()
			conn, err := grpc.CreateNewClientConnection(addr)
			if err != nil {
				e := GetProtoWorkerError("Error creating new client connection", err, nil)
				appendToOutput(&syncPB.WorkerBuildOutput{WorkerName: ep.Name, Error: &syncPB.WorkerBuildError{Error: e}})
			}
			defer conn.Close()

			providerClient := providerPB.NewRepositoryProviderClient(conn)
			s := WorkerBuilderServer{provider: providerClient}
			workerCfg := syncPB.WorkerConfig{
				WorkerName: ep.Name,
				Req: &providerPB.SyncRequest{
					Name:         wb.Name,
					Dir:          dir,
					CloneUrl:     wb.CloneURL,
					BranchName:   wb.BranchName,
					RemoteOrigin: wb.Remote,
					Token:        wb.Token,
					BranchRef:    wb.BranchRef,
				},
				BuildSteps: wb.Steps,
			}
			output := s.BuildRepo(&workerCfg)
			appendToOutput(output)
		}()
	}
	wg.Wait()
	return outputs
}

// SyncRepository syncs the repository to the latest commit of specified branch.
func (s *WorkerBuilderServer) syncRepository(cfg *syncPB.WorkerConfig) error {
	ctx := context.Background()
	path := filepath.Join(os.ExpandEnv(BuildPath), cfg.Req.Name)
	logger.Printf("Syncing repository %s", path)
	isMetadataExist := s.checkMetadatFileExist(cfg.Req.Name)

	if isMetadataExist == false {
		resp, err := s.provider.Clone(ctx, cfg.Req)
		if err != nil {
			return err
		}
		if resp.Error != nil {
			return fmt.Errorf("%s", resp.Error.Reason)
		}
		resp, err = s.provider.Fetch(ctx, cfg.Req)
		if err != nil {
			return err
		}
		if resp.Error != nil {
			return fmt.Errorf("%s", resp.Error.Reason)
		}
	} else {
		resp, err := s.provider.Fetch(ctx, cfg.Req)
		if err != nil {
			return err
		}
		if resp.Error != nil {
			return fmt.Errorf("%s", resp.Error.Reason)
		}
	}

	// create or update metadata file in build directory.
	err := s.createMetadataFile(cfg)
	if err != nil {
		return err
	}

	return nil
}

func (_ *WorkerBuilderServer) checkMetadatFileExist(name string) bool {
	logger.Printf("Checking .conflowci.toml metadata file exist...")
	path := filepath.Join(BuildPath, name, metadataFileName)

	cmd := exec.Command("cat", path)
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}

// CreateMetadataFile creates a metadata file for the build.
// it performs a checksum and provide other relevant details like source of the repository, time of
// build and creation.
// TODO: hash the repo in go and don't rely on linux utilities to do so
// in an attempt to keep cross compability
func (s *WorkerBuilderServer) createMetadataFile(cfg *syncPB.WorkerConfig) error {
	path := filepath.Join(BuildPath, cfg.Req.Name)
	logger.Printf("creating metadata file in path %s", path)

	cmd := fmt.Sprintf(`mkdir -p %s && find %s -type f \
  ! -path "*/.git/*" \
  ! -path "*/.conflowci.toml" \
  -exec sha256sum {} + | sort | sha256sum
`, path, path)

	c := exec.Command("bash", "-c", cmd)
	b, err := c.CombinedOutput()
	if err != nil {
		logger.Printf("checksum output: %v", string(b))
		return CheckSumError{message: err.Error()}
	}
	hash := strings.Split(string(b), " ")[0]

	metadata := BuildMetadata{
		Repository: RepositoryMetadata{
			Name:    cfg.Req.Name,
			Source:  cfg.Req.CloneUrl,
			Version: config.ConflowVersion,
		},
		State: StateMetadata{
			ClonedAt:  time.Now().Format(time.RFC3339),
			LastBuild: time.Now().Format(time.RFC3339),
			Checksum:  hash,
		},
	}

	metadataPath := filepath.Join(path, metadataFileName)

	var buf bytes.Buffer
	err = toml.NewEncoder(&buf).Encode(metadata)
	if err != nil {
		return MetadataEncodeError{message: err.Error()}
	}

	cmd = fmt.Sprintf(`echo '%s' > %s`, buf.String(), metadataPath)
	c = exec.Command("sh", "-c", cmd)
	_, err = c.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}
