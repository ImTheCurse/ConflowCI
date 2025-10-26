package grpc

import (
	"crypto/tls"
	"crypto/x509"
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
		c, err := GetClientTLSConfig(
			CAPath,
			ClientCertificatePath,
			ClientKeyPath,
		)
		if err != nil {
			return nil, err
		}
		creds = credentials.NewTLS(c)
	} else {
		creds = insecure.NewCredentials()
	}

	conn, err = grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
	return
}
func loadCA(path string) (*x509.CertPool, error) {
	ca, err := os.ReadFile(os.ExpandEnv(path))
	if err != nil {
		logger.Fatalf("Failed to load CA certificate: %v", err)
		return nil, err
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(ca)
	return pool, nil
}

func GetWorkerTLSConfig(rootCAPath, srvCertPath, srvKeyPath string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(srvCertPath, srvKeyPath)
	if err != nil {
		logger.Printf("Failed to load certificate: %v", err)
		return nil, err
	}

	CA, err := loadCA(rootCAPath)
	if err != nil {
		return nil, err
	}
	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		// Require and verify client cert for mTLS:
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  CA,
		MinVersion: tls.VersionTLS13,
	}
	return tlsCfg, nil
}
func GetClientTLSConfig(rootCAPath, clientCertPath, clientKeyPath string) (*tls.Config, error) {
	CA, err := loadCA(rootCAPath)
	if err != nil {
		return nil, err
	}
	cert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		logger.Printf("Failed to load certificate: %v", err)
		return nil, err
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      CA,
		MinVersion:   tls.VersionTLS13,
	}
	return tlsCfg, nil
}
