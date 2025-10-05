package sync

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/ImTheCurse/ConflowCI/internal/runner/provider"
	"github.com/ImTheCurse/ConflowCI/pkg/config"
	"github.com/google/uuid"
	"github.com/pelletier/go-toml"
	"golang.org/x/crypto/ssh"
)

func NewWorkerBuilder(conn *ssh.Client, cfg config.ValidatedConfig) *WorkerBuilder {
	return &WorkerBuilder{
		Name:     cfg.Pipeline.Build.Name,
		BuildID:  uuid.New(),
		State:    StartingBuild,
		RunsOn:   cfg.Endpoints,
		Steps:    cfg.Pipeline.Build.BuildSteps,
		CloneURL: cfg.GetCloneURL(),
		Conn:     conn,
	}
}

func (wb *WorkerBuilder) SyncRepository(rb provider.RepositoryReader) {

}

func (wb *WorkerBuilder) CreateMetadataFile() error {
	path := filepath.Join(buildPath, wb.Name)

	s, err := wb.Conn.NewSession()
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

	metadataWriteSession, err := wb.Conn.NewSession()
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
