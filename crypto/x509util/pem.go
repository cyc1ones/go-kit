package x509util

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func EncodePKCS8PrivateKey(pk any) ([]byte, error) {
	pkcs8, err := x509.MarshalPKCS8PrivateKey(pk)
	if err != nil {
		return nil, err
	}
	pem := pem.EncodeToMemory(&pem.Block{
		Bytes: pkcs8,
		Type:  "PRIVATE KEY",
	})
	return pem, nil
}

func EncodePKIXPublicKey(pubKey any) ([]byte, error) {
	pkix, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return nil, err
	}
	pem := pem.EncodeToMemory(&pem.Block{
		Bytes: pkix,
		Type:  "PUBLIC KEY",
	})
	return pem, nil
}

func DecodePKCS8PrivateKey(pemData []byte) (any, error) {
	block, _ := pem.Decode(pemData)
	if block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("unexpected block type: %s", block.Type)
	}
	return x509.ParsePKCS8PrivateKey(block.Bytes)
}

func DecodePKIXPublicKey(pemData []byte) (any, error) {
	block, _ := pem.Decode(pemData)
	if block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("unexpected block type: %s", block.Type)
	}
	return x509.ParsePKIXPublicKey(block.Bytes)
}

func LoadPKIXPublicKey(path string) (any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return DecodePKIXPublicKey(data)
}

func LoadPKCS8PrivateKey(path string) (any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return DecodePKCS8PrivateKey(data)
}
