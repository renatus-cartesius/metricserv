// Package encryption consists of a bunch of asymmetric encryption utils
package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"os"
)

var (
	ErrEmptyPrivateKey        = errors.New("private key is empty, set if firstly")
	ErrEmptyPublicKey         = errors.New("public key is empty, set if firstly")
	ErrCastingToRSAPublicKey  = errors.New("error when casting public key parsed by x509")
	ErrCastingToRSAPrivateKey = errors.New("error when casting private key parsed by x509")
)

type Processor interface {
	Decrypt([]byte) ([]byte, error)
	Encrypt([]byte) ([]byte, error)
}

type RSAProcessor struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey

	random io.Reader
}

func NewRSAProcessor() (*RSAProcessor, error) {
	return &RSAProcessor{
		random: rand.Reader,
	}, nil
}

func NewRSAPublicKey(keyPath string) (*rsa.PublicKey, error) {
	pemRaw, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	publicKeyBlock, _ := pem.Decode(pemRaw)
	publicKey, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)

	if err != nil {
		return nil, err
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, ErrCastingToRSAPublicKey
	}

	return rsaPublicKey, nil
}

func NewRSAPrivateKey(keyPath string) (*rsa.PrivateKey, error) {
	pemRaw, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	privateKeyBlock, _ := pem.Decode(pemRaw)
	return x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
}

func (rp *RSAProcessor) SetPublicKey(key *rsa.PublicKey) {
	rp.publicKey = key
}

func (rp *RSAProcessor) SetPrivateKey(key *rsa.PrivateKey) {
	rp.privateKey = key
}

func (rp *RSAProcessor) Encrypt(data []byte) ([]byte, error) {
	if rp.publicKey == nil {
		return nil, ErrEmptyPublicKey
	}

	cipherText, err := rsa.EncryptPKCS1v15(rp.random, rp.publicKey, data)
	if err != nil {
		return nil, err
	}

	return cipherText, nil
}

func (rp *RSAProcessor) Decrypt(ciphertext []byte) ([]byte, error) {
	if rp.privateKey == nil {
		return nil, ErrEmptyPrivateKey
	}

	data, err := rsa.DecryptPKCS1v15(rp.random, rp.privateKey, ciphertext)
	if err != nil {
		return nil, err
	}

	return data, nil
}
