package lc

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type Key struct {
	privateKey *rsa.PrivateKey
}

func NewKey(privateKeyFileName string) *Key {
	_, err := os.Stat(privateKeyFileName)
	key := &Key{privateKey: nil}
	if os.IsNotExist(err) {
		privateKeyFile, err := os.Create(privateKeyFileName)
		if err != nil {
			log.Fatalf("Error creating privateKeyFile: %v", err)
		}
		key.createPrivateKey()
		key.writePrivateKeyToFile(privateKeyFile)
		defer privateKeyFile.Close()
	} else {
		rawPrivKey, err := ioutil.ReadFile(privateKeyFileName)
		if err != nil {
			log.Fatalf("Could not open PEM file: %v", privateKeyFileName)
		}
		privPem, _ := pem.Decode(rawPrivKey)
		privPemBytes := privPem.Bytes
		parsedkey, err := x509.ParsePKCS1PrivateKey(privPemBytes)
		key.privateKey = parsedkey
		log.Println("Parsed Key: ", parsedkey)
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

func (k *Key) GetPublicKey() []byte {
	publicKey := &k.privateKey.PublicKey
	publicKeyInBytes := x509.MarshalPKCS1PublicKey(publicKey)
	return publicKeyInBytes
}

func (k *Key) SignMessage(message []byte) (signature []byte) {
	rnd := rand.Reader
	signature = nil
	hashedMessage := sha256.Sum256(message)
	signature, err := rsa.SignPKCS1v15(rnd, k.privateKey, crypto.SHA256, hashedMessage[:])
	if err != nil {
		log.Printf("Error occurred while signing message: %v ", err)
	} else {
		log.Printf("Signature generated.")
	}
	return signature
}

func VerifyMessage(hashedMessage []byte, signature []byte, publicKey *rsa.PublicKey) (valid bool) {
	err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashedMessage[:], signature)
	if err != nil {
		log.Printf("Error during signature verification: %v", err)
		valid = false
	} else {
		log.Printf("Signature validated")
		valid = true
	}
	return valid
}

func GetPublicKeyFromBytes(rawPublicKey []byte) (publicKey *rsa.PublicKey) {
	pk, err := x509.ParsePKCS1PublicKey(rawPublicKey)
	if err != nil {
		log.Printf("Could not parse rsa.PublicKey from the raw public key: %v", rawPublicKey)
	} else {
		publicKey = pk
	}

	return publicKey
}
