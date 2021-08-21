package ecdsa

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func LoadPrivateKey(filepath string) (*ecdsa.PrivateKey, error) {
	raw, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(raw)
	data := block.Bytes

	return x509.ParseECPrivateKey(data)
}

func LoadPublicKey(filepath string) (*ecdsa.PublicKey, error) {
	raw, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(raw)
	data := block.Bytes

	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKIXPublicKey(data); err != nil {
		if cert, err := x509.ParseCertificate(data); err == nil {
			parsedKey = cert.PublicKey
		} else {
			return nil, err
		}
	}

	var pubKey *ecdsa.PublicKey
	var ok bool
	if pubKey, ok = parsedKey.(*ecdsa.PublicKey); !ok {
		return nil, fmt.Errorf("data doesn't contain valid ECDSA Public Key")
	}

	return pubKey, nil
}
