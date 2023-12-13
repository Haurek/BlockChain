package mycrypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
)

// GenerateKeyPair generate ecdsa key pair
func GenerateKeyPair() (*ecdsa.PublicKey, *ecdsa.PrivateKey, error) {
	curve := elliptic.P256()

	// generate private key
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	// generate public key
	publicKey := &privateKey.PublicKey
	return publicKey, privateKey, nil
}

// Sign message use PrivateKey
func Sign(key *ecdsa.PrivateKey, message []byte) ([]byte, error) {
	hash := sha256.Sum256(message)

	// sign
	r, s, err := ecdsa.Sign(rand.Reader, key, hash[:])
	if err != nil {
		return nil, err
	}

	// serialize
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	signature := append(rBytes, sBytes...)

	return signature, nil
}

// Verify signature
func Verify(key *ecdsa.PublicKey, message []byte, signature []byte) bool {
	hash := sha256.Sum256(message)

	// get signature
	rBytes := signature[:len(signature)/2]
	sBytes := signature[len(signature)/2:]
	var r, s big.Int
	r.SetBytes(rBytes)
	s.SetBytes(sBytes)

	// verify
	return ecdsa.Verify(key, hash[:], &r, &s)
}

func Bytes2PublicKey(publicKeyBytes []byte) *ecdsa.PublicKey {
	x, y := elliptic.UnmarshalCompressed(elliptic.P256(), publicKeyBytes)
	return &ecdsa.PublicKey{elliptic.P256(), x, y}
}

func PublicKey2Bytes(key *ecdsa.PublicKey) []byte {
	return elliptic.MarshalCompressed(key.Curve, key.X, key.Y)
}

func LoadPublicKey(publicKeyPath string) (*ecdsa.PublicKey, error) {
	publicKeyBytes, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		return nil, err
	}

	pemBlock, _ := pem.Decode(publicKeyBytes)
	if pemBlock == nil {
		return nil, fmt.Errorf("failed to decode PEM block for public key")
	}

	publicKey, err := x509.ParsePKIXPublicKey(pemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	ecdsaPublicKey, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an ECDSA public key")
	}

	return ecdsaPublicKey, nil
}

func LoadPrivateKey(privateKeyPath string) (*ecdsa.PrivateKey, error) {
	privateKeyBytes, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}

	pemBlock, _ := pem.Decode(privateKeyBytes)
	if pemBlock == nil {
		return nil, fmt.Errorf("failed to decode PEM block for private key")
	}

	privateKey, err := x509.ParseECPrivateKey(pemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func SavePublicKey(publicKey *ecdsa.PublicKey, publicKeyPath string) error {
	file, err := os.Create(publicKeyPath)
	if err != nil {
		return err
	}
	defer file.Close()

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return err
	}

	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	err = pem.Encode(file, pemBlock)
	if err != nil {
		return err
	}

	return nil
}

func SavePrivateKey(privateKey *ecdsa.PrivateKey, privateKeyPath string) error {
	file, err := os.Create(privateKeyPath)
	if err != nil {
		return err
	}
	defer file.Close()

	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return err
	}

	pemBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	err = pem.Encode(file, pemBlock)
	if err != nil {
		return err
	}

	return nil
}
