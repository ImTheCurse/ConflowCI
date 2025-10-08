package github

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/ImTheCurse/ConflowCI/internal/sync"
	"github.com/ImTheCurse/ConflowCI/pkg/crypto"
	"github.com/ImTheCurse/ConflowCI/pkg/ssh"
)

func TestClonePullRequest(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "clone_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name            string
		payload         *PullRequestPayload
		token           string
		dir             string
		CheckFilesExist []string
		expectError     bool
	}{
		{
			name: "clone with a valid public repo URL",
			payload: &PullRequestPayload{
				Repository: Repository{
					Name: "Hello-World",
				},
				PullRequest: PullRequest{
					OriginBranch: Branch{
						Ref: "test",
						SHA: "b3cbd5bbd7e81436d2eee04537ea2b4c0cad4cdf",
						Repo: Repo{
							Name:     "Hello-World",
							CloneURL: "https://github.com/octocat/Hello-World.git",
							Owner: User{
								Login: "octat",
							},
						},
					},
				},
			},
			token:           "",
			dir:             filepath.Join(tempDir, "test1"),
			expectError:     false,
			CheckFilesExist: []string{"CONTRIBUTING.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// make sure that dir dosen't exist, otherwise
			// clone will fail.
			os.RemoveAll(tt.dir)

			repo, err := tt.payload.ClonePullRequest(tt.token, tt.dir)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if repo == nil {
				t.Errorf("expected repository to exist, got nil.")
			}
			for _, fname := range tt.CheckFilesExist {
				fpath := filepath.Join(tt.dir, fname)
				if _, err := os.Stat(fpath); errors.Is(err, os.ErrNotExist) {
					t.Errorf("expected file %s to exist, got %v", fname, err)
				}
			}
		})
	}

}

func TestClone(t *testing.T) {
	reader := GitRepoReader{
		Name:         "Hello-World",
		CloneURL:     "https://github.com/octocat/Hello-World.git",
		BranchName:   "test",
		RemoteOrigin: "",
	}
	pub, _, err := crypto.GenerateKeys()
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}
	defer os.RemoveAll("keys")
	ctx := context.Background()
	container, err := ssh.CreateSSHServerContainer(string(pub))
	if err != nil {
		t.Fatalf("Failed to start SSH server container: %v", err)
	}
	_, _, _ = container.Exec(ctx, []string{"apk", "add", "git"})

	Ep := ssh.Ep
	fmt.Println("SSH server running at", Ep.Host, Ep.Port)
	defer container.Terminate(ctx)

	port := strconv.Itoa(int(Ep.Port))
	err = ssh.AddHostKeyToKnownHosts(Ep.Host, port)
	if err != nil {
		t.Fatalf("Failed to add host key to known hosts: %v", err)
	}

	cfg, err := ssh.CreateTestConfig()
	if err != nil {
		t.Fatalf("Failed to create SSH config: %v", err)
	}

	conn, err := ssh.NewSSHConn(Ep, cfg)
	if err != nil {
		t.Fatalf("Failed to create SSH connection: %v", err)
	}
	defer conn.Close()

	out, err := reader.Clone(conn, sync.BuildPath)
	fmt.Printf("Clone output: %s", out)
	if err != nil {
		t.Fatalf("Failed to clone repository: %v", err)
	}

	s, err := conn.NewSession()
	if err != nil {
		t.Fatalf("Failed to create SSH session: %v", err)
	}
	defer s.Close()

	path := filepath.Join(sync.BuildPath, reader.Name, "README")
	cmd := `cat ` + path
	bt, err := s.CombinedOutput(cmd)
	if err != nil {
		t.Errorf("Failed to execute command: %v. got: %s", err, string(bt))
	}
	if string(bt) != "Hello World!\n" {
		t.Errorf("Expected %s got: %s", "Hello World!\n", string(bt))
	}
}

func TestFetch(t *testing.T) {
	reader := GitRepoReader{
		Name:     "demo-repo",
		CloneURL: "https://github.com/ImTheCurse/demo-repo",
		// we use a open forever pull request to check we are able to fetch a remote origin
		RemoteOrigin: "pull/5/head:pr-5",
	}
	pub, _, err := crypto.GenerateKeys()
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}
	defer os.RemoveAll("keys")
	ctx := context.Background()
	container, err := ssh.CreateSSHServerContainer(string(pub))
	if err != nil {
		t.Fatalf("Failed to start SSH server container: %v", err)
	}
	defer container.Terminate(ctx)

	_, _, _ = container.Exec(ctx, []string{"apk", "add", "git"})

	Ep := ssh.Ep
	fmt.Println("SSH server running at", Ep.Host, Ep.Port)

	port := strconv.Itoa(int(Ep.Port))
	err = ssh.AddHostKeyToKnownHosts(Ep.Host, port)
	if err != nil {
		t.Fatalf("Failed to add host key to known hosts: %v", err)
	}

	cfg, err := ssh.CreateTestConfig()
	if err != nil {
		t.Fatalf("Failed to create SSH config: %v", err)
	}

	conn, err := ssh.NewSSHConn(Ep, cfg)
	if err != nil {
		t.Fatalf("Failed to create SSH connection: %v", err)
	}
	defer conn.Close()

	_, err = reader.Clone(conn, sync.BuildPath)
	if err != nil {
		t.Fatalf("Failed to clone repository: %v", err)
	}
	path := filepath.Join(sync.BuildPath, reader.Name)
	fetchOut, err := reader.Fetch(conn, path)
	if err != nil {
		fmt.Printf("Fetch Out: %s", fetchOut)
		t.Fatalf("Failed to fetch repository: %v", err)
	}

	s, err := conn.NewSession()
	if err != nil {
		t.Fatalf("Failed to create SSH session: %v", err)
	}
	defer s.Close()

	cmd := fmt.Sprintf("cd %s && git checkout %s && cat %s", path, "pr-5", "README.md")
	bt, err := s.CombinedOutput(cmd)
	if err != nil {
		t.Errorf("Failed to execute command: %v. got: %s", err, string(bt))
	}

	actual := strings.TrimSpace(string(bt))
	expected := strings.TrimSpace(`Switched to branch 'pr-5'
	# demo-repo

	this is a demo small change`)

	actual = strings.ReplaceAll(actual, "\t", "")
	expected = strings.ReplaceAll(expected, "\t", "")

	if actual != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, actual)
	}
}

func TestCreateWorkTree(t *testing.T) {
	reader := GitRepoReader{
		Name:     "demo-repo",
		CloneURL: "https://github.com/ImTheCurse/demo-repo",
		// we use a open forever pull request to check we are able to fetch a remote origin
		RemoteOrigin: "pull/5/head:pr-5",
		BranchName:   "pr-5",
	}
	pub, _, err := crypto.GenerateKeys()
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}
	defer os.RemoveAll("keys")
	ctx := context.Background()
	container, err := ssh.CreateSSHServerContainer(string(pub))
	if err != nil {
		t.Fatalf("Failed to start SSH server container: %v", err)
	}
	defer container.Terminate(ctx)

	_, _, _ = container.Exec(ctx, []string{"apk", "add", "git"})

	Ep := ssh.Ep
	fmt.Println("SSH server running at", Ep.Host, Ep.Port)

	port := strconv.Itoa(int(Ep.Port))
	err = ssh.AddHostKeyToKnownHosts(Ep.Host, port)
	if err != nil {
		t.Fatalf("Failed to add host key to known hosts: %v", err)
	}

	cfg, err := ssh.CreateTestConfig()
	if err != nil {
		t.Fatalf("Failed to create SSH config: %v", err)
	}

	conn, err := ssh.NewSSHConn(Ep, cfg)
	if err != nil {
		t.Fatalf("Failed to create SSH connection: %v", err)
	}
	defer conn.Close()

	_, err = reader.Clone(conn, sync.BuildPath)
	if err != nil {
		t.Fatalf("Failed to clone repository: %v", err)
	}
	path := filepath.Join(sync.BuildPath, reader.Name)
	fetchOut, err := reader.Fetch(conn, path)
	if err != nil {
		fmt.Printf("Fetch Out: %s", fetchOut)
		t.Fatalf("Failed to fetch repository: %v", err)
	}
	err = reader.CreateWorkTree(conn, path, "../temp")
	if err != nil {
		t.Fatalf("Failed to create work tree: %v", err)
	}

	wrkTreePath := filepath.Join(path, "../temp")
	cmd := fmt.Sprintf("cd %s && cat %s", wrkTreePath, "README.md")

	s, err := conn.NewSession()
	if err != nil {
		t.Fatalf("Failed to create SSH session: %v", err)
	}
	defer s.Close()
	bt, err := s.CombinedOutput(cmd)
	if err != nil {
		t.Errorf("Failed to execute command: %v. got: %s", err, string(bt))
	}

	actual := strings.TrimSpace(string(bt))
	expected := strings.TrimSpace(`# demo-repo

	this is a demo small change`)

	actual = strings.ReplaceAll(actual, "\t", "")
	expected = strings.ReplaceAll(expected, "\t", "")

	if actual != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, actual)
	}
}

func TestRemoveWorkTree(t *testing.T) {
	reader := GitRepoReader{
		Name:     "demo-repo",
		CloneURL: "https://github.com/ImTheCurse/demo-repo",
		// we use a open forever pull request to check we are able to fetch a remote origin
		RemoteOrigin: "pull/5/head:pr-5",
		BranchName:   "pr-5",
	}
	pub, _, err := crypto.GenerateKeys()
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}
	defer os.RemoveAll("keys")
	ctx := context.Background()
	container, err := ssh.CreateSSHServerContainer(string(pub))
	if err != nil {
		t.Fatalf("Failed to start SSH server container: %v", err)
	}
	defer container.Terminate(ctx)

	_, _, _ = container.Exec(ctx, []string{"apk", "add", "git"})

	Ep := ssh.Ep
	fmt.Println("SSH server running at", Ep.Host, Ep.Port)

	port := strconv.Itoa(int(Ep.Port))
	err = ssh.AddHostKeyToKnownHosts(Ep.Host, port)
	if err != nil {
		t.Fatalf("Failed to add host key to known hosts: %v", err)
	}

	cfg, err := ssh.CreateTestConfig()
	if err != nil {
		t.Fatalf("Failed to create SSH config: %v", err)
	}

	conn, err := ssh.NewSSHConn(Ep, cfg)
	if err != nil {
		t.Fatalf("Failed to create SSH connection: %v", err)
	}
	defer conn.Close()

	_, err = reader.Clone(conn, sync.BuildPath)
	if err != nil {
		t.Fatalf("Failed to clone repository: %v", err)
	}
	path := filepath.Join(sync.BuildPath, reader.Name)
	fetchOut, err := reader.Fetch(conn, path)
	if err != nil {
		fmt.Printf("Fetch Out: %s", fetchOut)
		t.Fatalf("Failed to fetch repository: %v", err)
	}
	err = reader.CreateWorkTree(conn, path, "../temp")
	if err != nil {
		t.Fatalf("Failed to create work tree: %v", err)
	}
	err = reader.RemoveWorkTree(conn, path, "../temp")
	if err != nil {
		t.Fatalf("Failed to remove work tree: %v", err)
	}

	cmd := fmt.Sprintf("ls %s", sync.BuildPath)
	s, err := conn.NewSession()
	if err != nil {
		t.Fatalf("Failed to create SSH session: %v", err)
	}
	defer s.Close()
	bt, err := s.CombinedOutput(cmd)
	if err != nil {
		t.Errorf("Failed to execute command: %v. got: %s", err, string(bt))
	}

	if strings.Contains(string(bt), "temp") {
		t.Errorf("Unexpected temp folder, RemoveWorkTree should have removed the folder.")
	}

}
