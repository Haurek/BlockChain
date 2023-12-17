package blockchain

import (
	"encoding/json"
	"fmt"
	"testing"
)

type Msg struct {
	T    int    `json:"t"`
	Data []byte `json:"data"`
}

func TestBlock_NewBlock(t *testing.T) {
	input := []TXinput{
		{
			TxID:           []byte("1"),
			Index:          100,
			FromAddress:    []byte("input_address_1"),
			Signature:      []byte("input_signature_1"),
			PublicKeyBytes: []byte("input_public_key_1"),
		},
		// Add more inputs as needed
	}
	output := []TXoutput{
		{
			ToAddress:     []byte("input_address_1"),
			Value:         50,
			PublicKeyHash: []byte("input_public_key_1"),
		},
		// Add more outputs as needed
	}
	transaction := Transaction{
		ID:      []byte("some_transaction_id"),
		Inputs:  input,
		Outputs: output,
	}

	Transactions := []*Transaction{
		&transaction,
	}
	block := NewBlock([]byte("previous_hash"), Transactions, 0)
	// 使用 json.Marshal 进行序列化
	jsonData, err := json.Marshal(block)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	msg := Msg{
		T:    1,
		Data: jsonData,
	}
	mdata, err := json.Marshal(msg)
	// 输出序列化后的 JSON 数据
	var newBlock Block
	var newMsg Msg
	err = json.Unmarshal(mdata, &newMsg)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = json.Unmarshal(newMsg.Data, &newBlock)
	newBlock.Show()
}
