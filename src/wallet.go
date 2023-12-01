package BlockChain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"os"
)

// Wallet type
type Wallet struct {
	address        []byte // base58-like from public key
	privateKey     *ecdsa.PrivateKey
	publicKey      *ecdsa.PublicKey
	publicKeyBytes []byte
}

// GetAddress return wallet address
func (wallet *Wallet) GetAddress() []byte {
	return wallet.address
}

// GetPublicKeyBytes return wallet public key bytes
func (wallet *Wallet) GetPublicKeyBytes() []byte {
	return wallet.publicKeyBytes
}

// GetPublicKey return wallet public key
func (wallet *Wallet) GetPublicKey() *ecdsa.PublicKey {
	return wallet.publicKey
}

// SetAddress set wallet address
func (wallet *Wallet) SetAddress(address []byte) {
	wallet.address = address
}

// SetPublicKey set public key
func (wallet *Wallet) SetPublicKey(key *ecdsa.PublicKey) {
	wallet.publicKeyBytes = elliptic.MarshalCompressed(key.Curve, key.X, key.Y)
	wallet.publicKey = key
}

// SetPrivateKey set private key
func (wallet *Wallet) SetPrivateKey(key *ecdsa.PrivateKey) {
	wallet.privateKey = key
}

// CreateWallet create a new wallet
func CreateWallet() *Wallet {
	// generate ECDSA key pair
	wallet := &Wallet{}
	publicKey, privateKey, err := generateKeyPair()
	if err != nil {
		fmt.Println("Error generating key pair:", err)
	}
	wallet.InitWallet(publicKey, privateKey)
	return wallet
}

func (wallet *Wallet) InitWallet(publicKey *ecdsa.PublicKey, privateKey *ecdsa.PrivateKey) *Wallet {
	wallet.SetPublicKey(publicKey)
	wallet.SetPrivateKey(privateKey)
	wallet.SetAddress(GenerateAddress(wallet.GetPublicKeyBytes()))
	return wallet
}

// LoadWallet Load key pair as wallet from local file
func LoadWallet(publicKeyPath, privateKeyPath string) (*Wallet, error) {
	publicKey, err := LoadPublicKey(publicKeyPath)
	if err != nil {
		return nil, err
	}

	privateKey, err := LoadPrivateKey(privateKeyPath)
	if err != nil {
		return nil, err
	}

	wallet := &Wallet{}
	return wallet.InitWallet(publicKey, privateKey), nil
}

// SaveWallet save key pair in file
func (wallet *Wallet) SaveWallet(publicKeyPath, privateKeyPath string) error {
	if err := SavePublicKey(wallet.publicKey, publicKeyPath); err != nil {
		return err
	}

	if err := SavePrivateKey(wallet.privateKey, privateKeyPath); err != nil {
		_ = os.Remove(publicKeyPath)
		return err
	}

	return nil
}

// GenerateAddress generate address from public key to base58-like array
// refer to BTC address generation
func GenerateAddress(publicKeyBytes []byte) []byte {
	// generate public key sha256 hash
	keyHash := Sha256Hash(publicKeyBytes)

	// calculate RIPEMD-160 hash
	ripemd160HashValue := Ripemd160Hash(keyHash)

	// add version number as prefix
	versionHash := append([]byte{0x00}, ripemd160HashValue...)

	// calculate checksum
	checksum := CalculateChecksum(versionHash)

	Hash := append(versionHash, checksum...)

	// base58 encode the hash value
	return Base58Encode(Hash)
}

// Address2PublicKeyHash address to public key hash
func Address2PublicKeyHash(address []byte) []byte {
	return Base58Decode(address)
}

// CheckAddress check checksum of an address
func CheckAddress(addr []byte) bool {
	publickHash := Base58Decode(addr)
	// last 4 byte is checksum
	addr_checksum := publickHash[len(publickHash)-4:]

	version := publickHash[0]
	publickHash = publickHash[1 : len(publickHash)-4]

	// calculate
	checksum := CalculateChecksum(append([]byte{version}, publickHash...))

	return bytes.Compare(addr_checksum, checksum) == 0
}
