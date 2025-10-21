package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"flag"

	"github.com/ImTheCurse/ConflowCI/internal/mq"
	mqpb "github.com/ImTheCurse/ConflowCI/internal/mq/pb"
	"github.com/ImTheCurse/ConflowCI/internal/provider/github"
	providerPB "github.com/ImTheCurse/ConflowCI/internal/provider/pb"
	"github.com/ImTheCurse/ConflowCI/internal/sync"
	syncPB "github.com/ImTheCurse/ConflowCI/internal/sync/pb"
	grpcUtil "github.com/ImTheCurse/ConflowCI/pkg/grpc"
	"google.golang.org/grpc"
)

var logger = log.New(os.Stdout, "[Worker Main]: ", log.Lshortfile|log.LstdFlags)

var (
	port = flag.Int("port", 8918, "port to listen on")
	host = flag.String("addr", "localhost", "address to connect to")
)

func main() {
	grpcUtil.DefineFlags()
	flag.Parse()
	addr := fmt.Sprintf("%s:%d", *host, *port)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatalf("Failed to listen on port %d", *port)
	}
	server := grpc.NewServer()

	logger.Printf("Registering services...")
	providerPB.RegisterRepositoryProviderServer(server, &github.GitRepoReader{})

	// Connect to the local machine, since the worker execute a gRPC method locally
	// and was defined to connect to client.
	conn, err := grpcUtil.CreateNewClientConnection(addr)
	if err != nil {
		logger.Fatalf("Failed to create client connection: %v", err)
	}
	client := providerPB.NewRepositoryProviderClient(conn)
	wbs := sync.NewWorkerBuilderServer(client)
	syncPB.RegisterWorkerBuilderServer(server, wbs)
	mqpb.RegisterConsumerServicerServer(server, &mq.ConsumerServer{})
	// TODO:
	// syncPB.RegisterFileExtractorServer(server,sync)

	logger.Printf("gRPC server Listening on port %d", *port)
	if err := server.Serve(lis); err != nil {
		logger.Fatalf("Failed to serve gRPC server: %v", err)
	}

}
