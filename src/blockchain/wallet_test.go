package blockchain

import (
	"bytes"
	"fmt"
	"testing"
)

func Test_CreateWallet(t *testing.T) {
	wallet := CreateWallet()
	fmt.Println("+v\n", wallet)
}

func TestWallet_SaveWallet(t *testing.T) {
	wallet := CreateWallet()
	wallet.SaveWallet(LocalPublicKeyFile, LocalPrivateKeyFile)
}

func TestLoadWallet(t *testing.T) {
	wallet := CreateWallet()
	wallet.SaveWallet(LocalPublicKeyFile, LocalPrivateKeyFile)
	newWallet, _ := LoadWallet(LocalPublicKeyFile, LocalPrivateKeyFile)
	fmt.Println(bytes.Equal(wallet.publicKeyBytes, newWallet.GetPublicKeyBytes()))
}
