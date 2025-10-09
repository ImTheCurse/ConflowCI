package sync

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/ImTheCurse/ConflowCI/internal/provider"
	"github.com/ImTheCurse/ConflowCI/pkg/config"
	"github.com/ImTheCurse/ConflowCI/pkg/ssh"
	"github.com/google/uuid"
	"github.com/pelletier/go-toml"
	goSSH "golang.org/x/crypto/ssh"
)

func NewWorkerBuilder(cfg config.ValidatedConfig, remote, branch string) *WorkersBuilder {
	return &WorkersBuilder{
		Name:       cfg.Pipeline.Build.Name,
		BuildID:    uuid.New(),
		State:      StartingBuild,
		RunsOn:     cfg.Endpoints,
		Steps:      cfg.Pipeline.Build.BuildSteps,
		CloneURL:   cfg.GetCloneURL(),
		Remote:     remote,
		BranchName: branch,
	}
}

func (wb *WorkersBuilder) BuildRepository(rb provider.RepositoryReader) []WorkerBuildOutput {
	repoWithBranch := fmt.Sprintf("%s-%s", wb.Name, wb.BranchName)
	path := filepath.Join("..", repoWithBranch)
	dir := filepath.Join(BuildPath, wb.Name)
	outputs := []WorkerBuildOutput{}
	for _, ep := range wb.RunsOn {
		sshCfg := ssh.SSHConnConfig{
			Username:       ep.User,
			PrivateKeyPath: ep.PrivateKeyPath,
		}
		cfg, err := sshCfg.BuildConfig()
		if err != nil {
			e := fmt.Errorf("Error building ssh config: %w", err)
			outputs = append(outputs, WorkerBuildOutput{WorkerName: ep.Name, Output: "", Error: e})
			continue
		}

		conn, err := ssh.NewSSHConn(ep, cfg)
		if err != nil {
			e := fmt.Errorf("Error connecting to ssh: %w", err)
			outputs = append(outputs, WorkerBuildOutput{WorkerName: ep.Name, Output: "", Error: e})
			continue
		}
		// ensure connection is closed at each iteration
		func() {
			defer conn.Close()
			err = wb.syncRepository(conn, rb)
			if err != nil {
				e := fmt.Errorf("Error syncing repository: %w", err)
				outputs = append(outputs, WorkerBuildOutput{WorkerName: ep.Name, Output: "", Error: e})
				return
			}

			err = rb.CreateWorkTree(conn, dir, path)
			if err != nil {
				e := fmt.Errorf("Error creating worktree for repository: %w", err)
				outputs = append(outputs, WorkerBuildOutput{WorkerName: ep.Name, Output: "", Error: e})
				return
			}

			cdToWrkTree := "cd " + filepath.Join(dir, path)
			buildCmd := strings.Join(wb.Steps, "&&")
			cmd := cdToWrkTree + " && " + buildCmd

			s, err := conn.NewSession()
			if err != nil {
				e := fmt.Errorf("Error creating ssh session at buildRepository: %w", err)
				outputs = append(outputs, WorkerBuildOutput{WorkerName: ep.Name, Output: "", Error: e})
				return
			}
			defer s.Close()

			b, err := s.CombinedOutput(cmd)
			if err != nil {
				e := fmt.Errorf("Error trying to run build: %w", err)
				outputs = append(outputs, WorkerBuildOutput{WorkerName: ep.Name, Output: string(b), Error: e})
				return
			}
			err = rb.RemoveWorkTree(conn, dir, path)
			if err != nil {
				e := fmt.Errorf("Error removing worktree from repository: %w", err)
				outputs = append(outputs, WorkerBuildOutput{WorkerName: ep.Name, Output: "", Error: e})
				return
			}
			outputs = append(outputs, WorkerBuildOutput{WorkerName: ep.Name, Output: string(b), Error: nil})
		}()
	}
	return outputs
}

// SyncRepository syncs the repository to the latest commit of specified branch.
func (wb *WorkersBuilder) syncRepository(conn *goSSH.Client, rb provider.RepositoryReader) error {

	path := filepath.Join(BuildPath, wb.Name)
	logger.Printf("Syncing repository %s", path)
	isMetadataExist := wb.checkMetadatFileExist(conn)
	if isMetadataExist == false {
		_, err := rb.Clone(conn, BuildPath)
		if err != nil {
			return err
		}
		_, err = rb.Fetch(conn, path)
		if err != nil {
			return err
		}
	} else {
		_, err := rb.Fetch(conn, path)
		if err != nil {
			return err
		}
	}

	// create or update metadata file in build directory.
	err := wb.CreateMetadataFile(conn)
	if err != nil {
		return err
	}

	return nil
}

func (wb *WorkersBuilder) checkMetadatFileExist(conn *goSSH.Client) bool {
	logger.Printf("Checking .conflowci.toml metadata file exist...")
	s, err := conn.NewSession()
	if err != nil {
		return false
	}
	defer s.Close()
	path := filepath.Join(BuildPath, wb.Name, ".conflowci.toml")

	err = s.Run("cat " + path)
	if err != nil {
		return false
	}
	return true
}

// CreateMetadataFile creates a metadata file for the build.
// it performs a checksum and provide other relevant details like source of the repository, time of
// build and creation.
// it is the function user responsibilty to close the ssh client connection.
func (wb *WorkersBuilder) CreateMetadataFile(conn *goSSH.Client) error {
	logger.Printf("Creating or updating metadata file...")
	path := filepath.Join(BuildPath, wb.Name)

	s, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer s.Close()

	logger.Printf("creating metadata file in path %s", path)

	cmd := fmt.Sprintf(`mkdir -p %s && find %s -type f \
  ! -path "*/.git/*" \
  ! -path "*/.conflowci.toml" \
  -exec sha256sum {} + | sort | sha256sum
`, path, path)

	b, err := s.CombinedOutput(cmd)
	if err != nil {
		logger.Printf("checksum output: %v", string(b))
		return CheckSumError{message: err.Error()}
	}
	hash := strings.Split(string(b), " ")[0]

	metadata := BuildMetadata{
		Repository: RepositoryMetadata{
			Name:    wb.Name,
			Source:  wb.CloneURL,
			Version: config.ConflowVersion,
		},
		State: StateMetadata{
			ClonedAt:  time.Now().Format(time.RFC3339),
			LastBuild: time.Now().Format(time.RFC3339),
			Checksum:  hash,
		},
	}

	metadataPath := filepath.Join(path, ".conflowci.toml")

	var buf bytes.Buffer
	err = toml.NewEncoder(&buf).Encode(metadata)
	if err != nil {
		return MetadataEncodeError{message: err.Error()}
	}

	metadataWriteSession, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer s.Close()

	cmd = fmt.Sprintf(`echo '%s' > %s`, buf.String(), metadataPath)
	_, err = metadataWriteSession.CombinedOutput(cmd)
	if err != nil {
		return err
	}
	return nil
}
