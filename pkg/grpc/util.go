package grpc

import (
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var logger = log.New(os.Stdout, "[gRPC]: ", log.Lshortfile|log.LstdFlags)

func CreateNewClientConnection(addr string) (conn *grpc.ClientConn, err error) {
	var creds credentials.TransportCredentials
	if *TlsFlag {
		creds = credentials.NewClientTLSFromCert(nil, "")
	} else {
		creds = insecure.NewCredentials()
	}

	conn, err = grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
	return
}
