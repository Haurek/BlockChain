package blockchain

import (
	"fmt"
	"testing"
)

func TestVerifyTransactions(t *testing.T) {
	tx1 := &Transaction{
		ID: []byte("tx1"),
		Inputs: []TXinput{
			{
				TxID:           []byte("prevTx1"),
				Index:          0, // double spent
				FromAddress:    []byte("address1"),
				Signature:      []byte("signature1"),
				PublicKeyBytes: []byte("pubkey1"),
			},
		},
		Outputs: []TXoutput{
			{
				Value:         10,
				ToAddress:     []byte("address2"),
				PublicKeyHash: []byte("pubkeyhash2"),
			},
		},
	}

	// 创建一个交易2，包含一个双花输入
	tx2 := &Transaction{
		ID: []byte("tx2"),
		Inputs: []TXinput{
			{
				TxID:           []byte("prevTx2"),
				Index:          1,
				FromAddress:    []byte("address3"),
				Signature:      []byte("signature3"),
				PublicKeyBytes: []byte("pubkey3"),
			},
			{
				TxID:           []byte("prevTx1"),
				Index:          0, // double spent
				FromAddress:    []byte("address1"),
				Signature:      []byte("signature1"),
				PublicKeyBytes: []byte("pubkey1"),
			},
		},
		Outputs: []TXoutput{
			{
				Value:         5,
				ToAddress:     []byte("address4"),
				PublicKeyHash: []byte("pubkeyhash4"),
			},
			{
				Value:         5,
				ToAddress:     []byte("address5"),
				PublicKeyHash: []byte("pubkeyhash5"),
			},
		},
	}

	Txs := []*Transaction{tx1, tx2}

	usedOutputs := make(map[string]struct{})
	for _, Tx := range Txs {
		for _, input := range Tx.Inputs {
			// check double spent
			outputIdentifier := fmt.Sprintf("%s:%d", input.TxID, input.Index)
			if _, exists := usedOutputs[outputIdentifier]; exists {
				fmt.Println("double spent")
				return
			}
			usedOutputs[outputIdentifier] = struct{}{}
		}
	}
	fmt.Println("success")
}
