package blockchain

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestCreateChain(t *testing.T) {
	wallet, err := LoadWallet(LocalPublicKeyFile, LocalPrivateKeyFile)
	if err != nil {
		t.Error(err)
	}
	chain, err := CreateChain(wallet.GetAddress(), "./database/", "./test.log")
	if err != nil {
		t.Error(err)
	}
	fmt.Println("Tip: ", hex.EncodeToString(chain.Tip))
}
