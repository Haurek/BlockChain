package Wallet

import (
	"fmt"
	"testing"
)

func TestWallet_NewWallet(t *testing.T) {
	wallet := &Wallet{}
	wallet.NewWallet()
	fmt.Println("Public Key: ", wallet.PublicKey)
	fmt.Println("Private Key: ", wallet.PrivateKey)
	fmt.Println("Address: ", string(wallet.address))
}

func TestSign_and_Verify(t *testing.T) {
	wallet := &Wallet{}
	wallet.NewWallet()

	message := []byte("abcd")

	signature, err := wallet.Sign(message)
	if err != nil {
		fmt.Println("Error signing message:", err)
		return
	}

	fmt.Println("Signature:", string(signature))
	valid := wallet.Verify(message, signature)
	fmt.Println("Signature valid:", valid)
}

func TestWallet_CheckAddress(t *testing.T) {
	wallet := &Wallet{}
	wallet.NewWallet()
	fmt.Println(wallet.CheckAddress(string(wallet.address)))
}
