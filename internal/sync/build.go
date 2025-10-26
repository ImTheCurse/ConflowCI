package sync

import (
	"bytes"
	"context"
	"errors"
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
	googleGrpc "google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
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

func (s *WorkerBuilderServer) BuildRepository(ctx context.Context, cfg *syncPB.WorkerConfig) (
	*syncPB.WorkerBuildOutput, error) {
	repoWithBranch := fmt.Sprintf("%s-%s", cfg.Req.Name, cfg.Req.BranchName)
	path := filepath.Join("..", repoWithBranch)
	dir := filepath.Join(os.ExpandEnv(BuildPath), cfg.Req.Name)

	err := s.syncRepository(cfg)
	if err != nil {
		e := fmt.Sprintf("Error syncing repository: %s", err.Error())
		return &syncPB.WorkerBuildOutput{WorkerName: cfg.WorkerName, Error: &syncPB.WorkerBuildError{Error: e}}, err
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
		return &syncPB.WorkerBuildOutput{WorkerName: cfg.WorkerName, Error: &syncPB.WorkerBuildError{Error: e}}, err
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
		return &syncPB.WorkerBuildOutput{
			WorkerName: cfg.WorkerName, Output: string(b),
			Error: &syncPB.WorkerBuildError{Error: e},
		}, err
	}
	return &syncPB.WorkerBuildOutput{WorkerName: cfg.WorkerName, Output: string(b)}, nil
}

func (s *WorkerBuilderServer) RemoveRepositoryWorkspace(ctx context.Context, cfg *syncPB.WorkerConfig) (*emptypb.Empty, error) {
	dir := filepath.Join(os.ExpandEnv(BuildPath), cfg.Req.Name)
	branchName := strings.Split(cfg.Req.BranchRef, ":")[1]
	wrkTreeReq := providerPB.WorkTreeRequest{
		Name:            cfg.WorkerName,
		RepoDir:         dir,
		BranchName:      branchName,
		WorktreeRelPath: "../" + cfg.Req.Name + "-" + cfg.Req.BranchName,
	}
	resp, err := s.provider.RemoveWorkTree(ctx, &wrkTreeReq)
	if err != nil {
		e := GetProtoWorkerError("Error removing work tree", err, resp)
		return nil, errors.New(e)
	}
	return nil, nil
}

func (wb *WorkersBuilder) RemoveAllRepositoryWorkspaces() []error {
	dir := filepath.Join(os.ExpandEnv(BuildPath), wb.Name)
	errs := []error{}
	var wg sync.WaitGroup
	var mu sync.Mutex
	wg.Add(len(wb.RunsOn))

	for _, ep := range wb.RunsOn {
		addr := formatAddress(ep)
		go func() {
			defer logger.Println("wg done in @RemoveAllRepositoryWorkspaces")
			defer wg.Done()
			conn, err := grpc.CreateNewClientConnection(addr)
			if err != nil {
				e := GetProtoWorkerError("Error creating new client connection", err, nil)
				ConcurrentAppendToArray(&mu, errors.New(e), &errs)
			}
			defer conn.Close()

			workerCfg, s := wb.getWorkerConfig(conn, ep.Name, dir)
			ctx := context.Background()
			_, err = s.RemoveRepositoryWorkspace(ctx, workerCfg)
			ConcurrentAppendToArray(&mu, err, &errs)
		}()
	}
	wg.Wait()
	return errs

}

func (wb *WorkersBuilder) getWorkerConfig(conn *googleGrpc.ClientConn, workerName, dir string) (
	*syncPB.WorkerConfig, WorkerBuilderServer) {
	providerClient := providerPB.NewRepositoryProviderClient(conn)
	s := WorkerBuilderServer{provider: providerClient}
	workerCfg := syncPB.WorkerConfig{
		WorkerName: workerName,
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
	return &workerCfg, s
}

func formatAddress(ep config.EndpointInfo) string {
	if ep.Port == 0 {
		return ep.Host
	}
	return fmt.Sprintf("%s:%d", ep.Host, ep.Port)
}

func (wb *WorkersBuilder) BuildAllEndpoints() []*syncPB.WorkerBuildOutput {
	outputs := []*syncPB.WorkerBuildOutput{}
	dir := filepath.Join(os.ExpandEnv(BuildPath), wb.Name)

	var wg sync.WaitGroup
	var mu sync.Mutex
	wg.Add(len(wb.RunsOn))

	for _, ep := range wb.RunsOn {
		addr := formatAddress(ep)
		go func() {
			defer logger.Println("wg done in @BuildAllEndpoints")
			defer wg.Done()
			conn, err := grpc.CreateNewClientConnection(addr)
			if err != nil {
				e := GetProtoWorkerError("Error creating new client connection", err, nil)
				ConcurrentAppendToArray(
					&mu,
					&syncPB.WorkerBuildOutput{WorkerName: ep.Name, Error: &syncPB.WorkerBuildError{Error: e}},
					&outputs,
				)
			}
			defer conn.Close()

			workerCfg, s := wb.getWorkerConfig(conn, ep.Name, dir)
			ctx := context.Background()
			output, err := s.BuildRepository(ctx, workerCfg)
			if err != nil {
				e := GetProtoWorkerError("Error Building repository", err, nil)
				ConcurrentAppendToArray(
					&mu,
					&syncPB.WorkerBuildOutput{WorkerName: ep.Name, Error: &syncPB.WorkerBuildError{Error: e}},
					&outputs,
				)
			}
			ConcurrentAppendToArray(&mu, output, &outputs)
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
