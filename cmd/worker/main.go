package main

import (
	"log"
	"net"
	"os"
	"strconv"

	"github.com/ImTheCurse/ConflowCI/internal/provider/github"
	githubPb "github.com/ImTheCurse/ConflowCI/internal/provider/github/pb"
	"google.golang.org/grpc"
)

var logger = log.New(os.Stdout, "[Worker Main]: ", log.Lshortfile|log.LstdFlags)

func main() {
	port := 8918
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		logger.Fatalf("Failed to listen on port %d", port)
	}
	server := grpc.NewServer()

	logger.Printf("Registering services...")
	// register services here
	githubPb.RegisterGithubProviderServer(server, &github.GitRepoReader{})

	logger.Printf("gRPC server Listening on port %d", port)
	if err := server.Serve(lis); err != nil {
		logger.Fatalf("Failed to serve gRPC server: %v", err)
	}

}
