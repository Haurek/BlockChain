package BlockChain

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestCreateChain(t *testing.T) {
	wallet, err := LoadWallet(LocalPublicKeyFile, LocalPrivateKeyFile)
	HandleError(err)
	chain := CreateChain(wallet.GetAddress())
	fmt.Println("Tip: ", hex.EncodeToString(chain.Tip))
}

func TestLoadChain(t *testing.T) {
	chain := LoadChain()
	fmt.Println("Tip: ", hex.EncodeToString(chain.Tip))
}

func TestAddBlock(t *testing.T) {
	chain := LoadChain()
	wallet, err := LoadWallet(LocalPublicKeyFile, LocalPrivateKeyFile)
	HandleError(err)
	toWallet, err := LoadWallet("./wallet/hh_publickey.pem", "./wallet/hh_privatekey.pem")
	HandleError(err)
	Tx, err := NewTransaction(wallet, chain, toWallet.address, 1)
	HandleError(err)
	Txs := []*Transaction{Tx}
	block := NewBlock(chain.Tip, Txs)
	chain.AddBlock(block)
	fmt.Println("new block hash: ", hex.EncodeToString(block.Header.Hash))
	fmt.Println("chain tip: ", hex.EncodeToString(chain.Tip))

}

func TestFindBlock(t *testing.T) {
	chain := LoadChain()
	fmt.Println("Tip: ", hex.EncodeToString(chain.Tip))
	block := chain.FindBlock(chain.Tip)
	fmt.Println("Block hash: ", hex.EncodeToString(block.Header.Hash))
}

func TestChain_FindUTXO(t *testing.T) {
	chain := LoadChain()
	block := chain.FindBlock(chain.Tip)
	fmt.Println("Block Tx ID: ", hex.EncodeToString(block.Transactions[0].ID))
	utxosMap := chain.FindUTXO()
	for key, value := range utxosMap {
		fmt.Println("ID: ", key)
		for _, out := range value {
			fmt.Println("index: ", out.Index)
			fmt.Println("value: ", out.Output.Value)
		}
	}
}
