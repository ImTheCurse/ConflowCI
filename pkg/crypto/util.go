package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

var logger = log.New(os.Stdout, "[Crypto]: ", log.Lshortfile|log.LstdFlags)

// GenerateKeys generates a new RSA key pair and saves them to files.
func GenerateKeys() ([]byte, []byte, error) {
	if err := os.MkdirAll("keys", 0700); err != nil {
		return nil, nil, err
	}

	logger.Println("Generating public and private keys...")

	// Generate key
	privateKey, err := rsa.GenerateKey(rand.Reader, 3072)
	if err != nil {
		return nil, nil, err
	}
	err = privateKey.Validate()
	if err != nil {
		return nil, nil, err
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}
	publicKeyBytes := ssh.MarshalAuthorizedKey(publicKey)

	logger.Println("Generated keys: id_rsa (private), id_rsa.pub (public)")
	os.WriteFile("keys/id_rsa.pub", publicKeyBytes, 0644)
	os.WriteFile("keys/id_rsa", privateKeyPEM, 0600)
	return publicKeyBytes, privateKeyPEM, nil
}
