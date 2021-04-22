package lc

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
)

type Key struct {
	privateKey *rsa.PrivateKey
}

func NewKey() *Key {
	config := NewConfig()
	_, err := os.Stat(config.privateKeyFileName)
	key := &Key{privateKey: nil}
	if os.IsNotExist(err) {
		privateKeyFile, err := os.Create(config.privateKeyFileName)
		if err != nil {
			log.Fatalf("Error creating privateKeyFile: %v", err)
		}
		key.createPrivateKey()
		key.writePrivateKeyToFile(privateKeyFile)
		defer privateKeyFile.Close()
	}
	return key
}

/**
Parts of this code are taken from: https://www.systutorials.com/how-to-generate-rsa-private-and-public-key-pair-in-go-lang/
*/

func (k *Key) createPrivateKey() {
	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("Cannot generate RSA key\n")
		os.Exit(1)
	}
	k.privateKey = privatekey
}
func (k *Key) writePrivateKeyToFile(filePath *os.File) {
	var privateKeyBytes []byte = x509.MarshalPKCS1PrivateKey(k.privateKey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	err := pem.Encode(filePath, privateKeyBlock)
	if err != nil {
		fmt.Printf("error when encode private pem: %s \n", err)
		os.Exit(1)
	}
}
