package Wallet

import (
	"fmt"
	"testing"
)

func TestWallet_NewWallet(t *testing.T) {
	wallet := &Wallet{}
	wallet.NewWallet()
	fmt.Println("Public Key: ", string(wallet.GetPublicKey()))
	fmt.Println("Private Key: ", wallet.PrivateKey)
	fmt.Println("Address: ", string(wallet.GetAddress()))
}

func TestWallet_CheckAddress(t *testing.T) {
	wallet := &Wallet{}
	wallet.NewWallet()
	fmt.Println(CheckAddress(wallet.Address))
}
