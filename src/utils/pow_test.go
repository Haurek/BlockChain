package utils

//
//import (
//	"fmt"
//	"testing"
//)
//
//func TestMining_and_Proof(t *testing.T) {
//	input := []TXinput{
//		{
//			Index:     1,
//			Value:     100,
//			Address:   []byte("input_address_1"),
//			Signature: []byte("input_signature_1"),
//			PublicKey: []byte("input_public_key_1"),
//		},
//		// Add more inputs as needed
//	}
//	output := []TXoutput{
//		{
//			Index:   1,
//			Value:   50,
//			Address: []byte("output_address_1"),
//		},
//		// Add more outputs as needed
//	}
//	transaction := Transaction{
//		ID:     []byte("some_transaction_id"),
//		Inputs: input,
//		Output: output,
//	}
//
//	block := &Block{}
//	Transactions := []*Transaction{
//		&transaction,
//	}
//	block.NewBlock([]byte("previous_hash"), Transactions)
//	fmt.Println(block.Header.Hash)
//	fmt.Println(block.Header.Nonce)
//	fmt.Println(Proof(block.Header))
//}
