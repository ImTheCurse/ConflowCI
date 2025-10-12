package main

import (
	"context"
	"flag"
	"log"

	githubPb "github.com/ImTheCurse/ConflowCI/internal/provider/github/pb"
	grpcUtil "github.com/ImTheCurse/ConflowCI/pkg/grpc"
)

var ()

func main() {
	grpcUtil.DefineFlags()
	flag.Parse()
	addr := "localhost:8918"

	conn, err := grpcUtil.CreateNewClientConnection(addr)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()
	req := githubPb.SyncRequest{
		Name:         "demo-repo",
		CloneUrl:     "https://github.com/ImTheCurse/demo-repo",
		BranchRef:    "pull/6/head:pr-6",
		BranchName:   "main",
		RemoteOrigin: "origin",
		Dir:          "/tmp/build-test",
	}
	client := githubPb.NewGithubProviderClient(conn)
	resp, err := client.Clone(ctx, &req)
	if err != nil {
		log.Fatalf("failed to clone repo: %v", err)
	}
	log.Printf("got response: %v", resp)

	// app := fiber.New()
	// githubRouter := app.Group("/github")
	// router.TaskRouter(githubRouter)

	// app.Listen(":7777")

}
