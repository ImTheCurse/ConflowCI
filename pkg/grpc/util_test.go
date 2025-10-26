package grpc

import (
	"flag"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TestCreateNewClientConnection(t *testing.T) {
	DefineFlags()
	flag.Parse()
	addr := "localhost:4093"

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatal("Failed to listen on tcp")
	}

	ServerCertificatePath = "testdata/fake_conflow_server_cert"
	ServerKeyPath = "testdata/fake_conflow_server_key"
	CAPath = "testdata/fake_root_ca"
	ClientCertificatePath = "testdata/fake_conflow_client_cert"
	ClientKeyPath = "testdata/fake_conflow_client_key"

	tlsCfg, err := GetWorkerTLSConfig(
		CAPath,
		ServerCertificatePath,
		ServerKeyPath,
	)
	if err != nil {
		t.Fatalf("Failed to get TLS config: %v", err)
	}
	creds := credentials.NewTLS(tlsCfg)
	server := grpc.NewServer(grpc.Creds(creds))
	t.Cleanup(func() {
		server.Stop()
	})

	go func() {
		logger.Println("gRPC server started")
		if err := server.Serve(lis); err != nil {
			t.Errorf("Failed to serve gRPC server: %v", err)
		}
	}()
	_, err = CreateNewClientConnection(addr)
	if err != nil {
		t.Errorf("Failed to create client connection: %v", err)
	}
}
