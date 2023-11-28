package block

import (
	"fmt"
	"testing"
)

func TestBlock_NewBlock(t *testing.T) {
	prehash := "prehash"
	block := &Block{}
	block.NewBlock(prehash)
	fmt.Println("Timestamp: ", block.Header.Timestamp)
	fmt.Println("Hash: ", block.Header.Hash)
	fmt.Println("PreHash: ", block.Header.PrevHash)
	fmt.Println("Nonce: ", block.Header.Nonce)
	fmt.Println("Height: ", block.Header.Height)
	fmt.Println("MerkelRoot: ", block.Header.MerkelRoot)
	fmt.Println("Transactions: ", block.Transactions)
}
