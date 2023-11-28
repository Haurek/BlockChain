// Package Wallet
package Wallet

import (
	"BlockChain/src/base58"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	//"goland.org/x/crypto/ripemd160"
	"goland/x/crypto/ripemd160"
	"math/big"
)

// Wallet type
type Wallet struct {
	address    []byte // base58-like from public key
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
}

// NewWallet create a new wallet
func (wallet *Wallet) NewWallet() *Wallet {
	// generate ECDSA key pair
	err := wallet.generateKeyPair()
	if err != nil {
		fmt.Println("Error generating key pair:", err)
	}

	// generate address
	wallet.generateAddress()

	return wallet
}

// generateKeyPair generate ecdsa key pair
func (wallet *Wallet) generateKeyPair() error {
	curve := elliptic.P256()

	// generate private key
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return err
	}

	// generate public key
	publicKey := &privateKey.PublicKey

	wallet.PrivateKey = privateKey
	wallet.PublicKey = publicKey
	return nil
}

// generateAddress generate address from public key to base58-like array
// refer to BTC address generation
func (wallet *Wallet) generateAddress() []byte {
	// serialize public key
	publicKeyBytes := elliptic.Marshal(wallet.PublicKey.Curve, wallet.PublicKey.X, wallet.PublicKey.Y)

	// generate public key sha256 hash
	hash := sha256.New()
	hash.Write(publicKeyBytes)
	sha256Hash := hash.Sum(nil)

	// calculate RIPEMD-160 hash
	ripemd160Hash := ripemd160.New()
	ripemd160Hash.Write(sha256Hash)
	ripemd160HashValue := ripemd160Hash.Sum(nil)

	// add version number as prefix
	versionHash := append([]byte{0x00}, ripemd160HashValue...)

	// calculate checksum
	checksum := calculateChecksum(versionHash)

	Hash := append(versionHash, checksum...)

	// base58 encode the hash value
	wallet.address = base58.Base58Encode(Hash)
	return wallet.address
}

// CheckAddress check checksum of an address
func (wallet *Wallet) CheckAddress(addr string) bool {
	publickHash := base58.Base58Decode([]byte(addr))
	// last 4 byte is checksum
	addr_checksum := publickHash[len(publickHash)-4:]

	version := publickHash[0]
	publickHash = publickHash[1 : len(publickHash)-4]

	// calculate
	checksum := calculateChecksum(append([]byte{version}, publickHash...))

	return bytes.Compare(addr_checksum, checksum) == 0
}

// Sign a message
func (wallet *Wallet) Sign(message []byte) ([]byte, error) {
	hash := sha256.Sum256(message)

	r, s, err := ecdsa.Sign(rand.Reader, wallet.PrivateKey, hash[:])
	if err != nil {
		return nil, err
	}

	// serialize
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	signature := append(rBytes, sBytes...)

	return signature, nil
}

// Verify a message
func (wallet *Wallet) Verify(message []byte, signature []byte) bool {
	hash := sha256.Sum256(message)

	rBytes := signature[:len(signature)/2]
	sBytes := signature[len(signature)/2:]
	var r, s big.Int
	r.SetBytes(rBytes)
	s.SetBytes(sBytes)

	return ecdsa.Verify(wallet.PublicKey, hash[:], &r, &s)
}

// calculateChecksum calculate checksum for address generate
func calculateChecksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])
	// the first 4 byte are checksum
	return secondHash[:4]
}
