package util

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"log"
)

func KeyToFile(key *ecdsa.PrivateKey, filename string) error {
	bytes, err := x509.MarshalECPrivateKey(key)
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: bytes,
	})
	err = ioutil.WriteFile(filename, pemBytes, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	return err
}

func FileToKey(filename string) (*ecdsa.PrivateKey, error) {
	pemBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalln(err)
	}

	block, _ := pem.Decode(pemBytes)

	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		log.Fatalln(err)
	}
	return privateKey, err
}
