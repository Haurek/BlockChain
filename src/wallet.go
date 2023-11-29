package BlockChain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"

	//"goland.org/x/crypto/ripemd160"
	"goland/x/crypto/ripemd160"
)

// Wallet type
type Wallet struct {
	Address    []byte // base58-like from public key
	PrivateKey *ecdsa.PrivateKey
	PublicKey  []byte
}

//// GetAddress return wallet address
//func (wallet *Wallet) GetAddress() []byte {
//	return wallet.Address
//}
//
//// GetPublicKey return wallet public key
//func (wallet *Wallet) GetPublicKey() []byte {
//	return wallet.PublicKey
//}

// NewWallet create a new wallet
func (wallet *Wallet) NewWallet() *Wallet {
	// generate ECDSA key pair
	err := wallet.generateKeyPair()
	if err != nil {
		fmt.Println("Error generating key pair:", err)
	}

	wallet.Address = GenerateAddress(wallet.PublicKey)

	return wallet
}

// LoadWallet Load key pair as wallet from local file
func (wallet *Wallet) LoadWallet(path string) *Wallet {
	// TODO
	return nil
}

// saveKeyPair save key pair in file
func (wallet *Wallet) saveKeyPair() {
	// TODO
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
	wallet.PublicKey = elliptic.MarshalCompressed(publicKey.Curve, publicKey.X, publicKey.Y)
	return nil
}

// GenerateAddress generate address from public key to base58-like array
// refer to BTC address generation
func GenerateAddress(publicKeyBytes []byte) []byte {
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
	return Base58Encode(Hash)
}

// Sign message use Wallet.PrivateKey
func (wallet *Wallet) Sign(message []byte) ([]byte, error) {
	hash := sha256.Sum256(message)

	// sign
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

// Verify signature
func Verify(publicKey []byte, message []byte, signature []byte) bool {
	hash := sha256.Sum256(message)

	// get public key
	x, y := elliptic.UnmarshalCompressed(elliptic.P256(), publicKey)
	key := ecdsa.PublicKey{elliptic.P256(), x, y}

	// get signature
	rBytes := signature[:len(signature)/2]
	sBytes := signature[len(signature)/2:]
	var r, s big.Int
	r.SetBytes(rBytes)
	s.SetBytes(sBytes)

	// verify
	return ecdsa.Verify(&key, hash[:], &r, &s)
}

// CheckAddress check checksum of an address
func CheckAddress(addr []byte) bool {
	publickHash := Base58Decode(addr)
	// last 4 byte is checksum
	addr_checksum := publickHash[len(publickHash)-4:]

	version := publickHash[0]
	publickHash = publickHash[1 : len(publickHash)-4]

	// calculate
	checksum := calculateChecksum(append([]byte{version}, publickHash...))

	return bytes.Compare(addr_checksum, checksum) == 0
}

// calculateChecksum calculate checksum for address generate
func calculateChecksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])
	// the first 4 byte are checksum
	return secondHash[:4]
}
