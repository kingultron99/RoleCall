package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
)

func GenerateKeyPair() (err error) {
	// Generate a new ECDSA private key using the P-521 curve
	privKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return err
	}

	// Marshal the private key
	privBytes, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		return err
	}

	// Marshal the public key
	pubBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		return err
	}

	// Encode the private and public keys to PEM format
	privblock := pem.Block{
		Type:    "EC PRIVATE KEY",
		Headers: nil,
		Bytes:   privBytes,
	}
	pubblock := pem.Block{
		Type:    "PUBLIC KEY",
		Headers: nil,
		Bytes:   pubBytes,
	}
	privPEM := pem.EncodeToMemory(&privblock)
	pubPEM := pem.EncodeToMemory(&pubblock)

	err = os.WriteFile(os.Getenv("KEY_DIR")+"/private_key.pem", privPEM, 0600)
	if err != nil {
		return err
	}

	err = os.WriteFile(os.Getenv("KEY_DIR")+"/public_key.pem", pubPEM, 0644)
	if err != nil {
		return err
	}

	return nil
}

func LoadPrivateKey() (*ecdsa.PrivateKey, error) {
	keybytes, err := os.ReadFile(os.Getenv("KEY_DIR") + "/private_key.pem")
	if err != nil {
		log.Printf("failed to read private key: %v", err)
		return nil, err
	}

	block, _ := pem.Decode(keybytes)
	if block == nil || block.Type != "EC PRIVATE KEY" {
		log.Printf("failed to decode PEM block containing private key")
		return nil, err
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		log.Printf("failed to parse private key: %v", err)
		return nil, err
	}
	return key, nil
}

func LoadPublicKey() (*ecdsa.PublicKey, error) {
	keybytes, err := os.ReadFile(os.Getenv("KEY_DIR") + "/public_key.pem")
	if err != nil {
		log.Printf("failed to read public key: %v", err)
		return nil, err
	}

	block, _ := pem.Decode(keybytes)
	if block == nil || block.Type != "PUBLIC KEY" {
		log.Printf("failed to decode PEM block containing public key")
		return nil, err
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Printf("failed to parse public key: %v", err)
		return nil, err
	}

	pubKey, ok := pubInterface.(*ecdsa.PublicKey)
	if !ok {
		log.Printf("not ECDSA public key")
		return nil, err
	}
	return pubKey, nil
}
