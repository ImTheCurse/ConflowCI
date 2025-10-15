package sync

import (
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"testing"
	"time"

	git "github.com/ImTheCurse/ConflowCI/internal/provider/github"
	providerPB "github.com/ImTheCurse/ConflowCI/internal/provider/pb"
	syncPB "github.com/ImTheCurse/ConflowCI/internal/sync/pb"
	"github.com/ImTheCurse/ConflowCI/pkg/config"
	cgrpc "github.com/ImTheCurse/ConflowCI/pkg/grpc"
	"github.com/google/uuid"
	"github.com/pelletier/go-toml"
	"google.golang.org/grpc"
)

func RunGRPCBuilderServer(t *testing.T, portCh chan<- int) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		logger.Fatal("Failed to listen on RunGRPCBuilder server")
	}
	port := lis.Addr().(*net.TCPAddr).Port
	portCh <- port
	server := grpc.NewServer()

	logger.Printf("Registering services...")
	providerPB.RegisterRepositoryProviderServer(server, &git.GitRepoReader{})

	logger.Printf("gRPC server Listening on port %d", port)
	if err := server.Serve(lis); err != nil {
		logger.Fatalf("Failed to serve gRPC server: %v", err)
	}
}

func TestCreateMetadataFile(t *testing.T) {
	BuildPath = t.TempDir()
	defer os.RemoveAll(filepath.Join(BuildPath, "../"))
	cgrpc.DefineFlags()
	*cgrpc.TlsFlag = false
	flag.Parse()
	ch := make(chan int)
	go RunGRPCBuilderServer(t, ch)

	port := <-ch

	conn, err := cgrpc.CreateNewClientConnection(fmt.Sprintf("localhost:%d", port))
	if err != nil {
		t.Fatalf("Failed to create client connection: %v", err)
	}
	p := providerPB.NewRepositoryProviderClient(conn)
	s := WorkerBuilderServer{provider: p}

	cfg := syncPB.WorkerConfig{
		WorkerName: "test-worker",
		Req: &providerPB.SyncRequest{
			Name:     "test",
			CloneUrl: "https://github.com/user/testproject.git",
		},
	}

	if err := s.createMetadataFile(&cfg); err != nil {
		t.Fatalf("CreateMetadataFile returned error: %v", err)
	}

	metadataPath := filepath.Join(os.ExpandEnv(BuildPath), cfg.Req.Name, ".conflowci.toml")

	cmd := exec.Command("cat", metadataPath)
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to output metadata file: %v, output: %v", err, string(b))
	}
	reader := bytes.NewReader(b)

	var metadata BuildMetadata
	if err := toml.NewDecoder(reader).Decode(&metadata); err != nil {
		t.Fatalf("Failed to decode TOML: %v", err)
	}

	if metadata.Repository.Name != cfg.Req.Name {
		t.Errorf("Expected Repository.Name %s, got %s", cfg.Req.Name, metadata.Repository.Name)
	}
	if metadata.Repository.Source != cfg.Req.CloneUrl {
		t.Errorf("Expected Repository.Source %s, got %s", cfg.Req.CloneUrl, metadata.Repository.Source)
	}
	if metadata.State.Checksum == "" {
		t.Errorf("Expected non-empty checksum")
	}

	// Validate timestamps roughly (within 10 seconds)
	now := time.Now()
	clonedAt, _ := time.Parse(time.RFC3339, metadata.State.ClonedAt)
	if now.Sub(clonedAt) > 10*time.Second {
		t.Errorf("ClonedAt timestamp is too old")
	}
}

func TestBuildRepository(t *testing.T) {
	BuildPath = t.TempDir()
	defer os.RemoveAll(filepath.Join(BuildPath, "../"))
	cgrpc.DefineFlags()
	*cgrpc.TlsFlag = false
	flag.Parse()
	ch := make(chan int)
	go RunGRPCBuilderServer(t, ch)

	port := <-ch

	u, err := user.Current()
	if err != nil {
		t.Fatalf("Unable to get os user. got error: %s", err)
	}

	ep := config.EndpointInfo{
		Name: "test",
		Host: "localhost",
		Port: uint16(port),
	}

	wb := &WorkersBuilder{
		Name:       "demo-repo",
		BuildID:    uuid.New(),
		State:      StartingBuild,
		CloneURL:   "https://github.com/ImTheCurse/demo-repo.git",
		RunsOn:     []config.EndpointInfo{ep},
		Steps:      []string{"chmod +x whoami.sh", "./whoami.sh", `echo "third command"`},
		BranchRef:  "pull/6/head:pr-6",
		Remote:     "origin",
		BranchName: "another-change",
	}
	outputs := wb.BuildAllEndpoints()
	fmt.Printf("outputs: %v", outputs)
	if len(outputs) != 1 {
		t.Errorf("Expected 1 output, got %d", len(outputs))
	}
	expectedOut := strings.TrimSpace(fmt.Sprintf(`I am: %s
        small change
        third command`, u.Username))
	actual := strings.TrimSpace(outputs[0].Output)

	actual = strings.ReplaceAll(actual, " ", "")
	expected := strings.ReplaceAll(expectedOut, " ", "")

	if actual != expected {
		t.Errorf("Error: expected: %s, got: %s", expectedOut, outputs[0].Output)
	}
}
