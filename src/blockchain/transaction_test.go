package blockchain

//
//import (
//	"fmt"
//	"testing"
//)
//
//func TestTransaction_Serialize(t *testing.T) {
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
//	// 序列化为字节流
//	serializedData, err := transaction.Serialize()
//	if err != nil {
//		fmt.Println("Error during serialization:", err)
//		return
//	}
//
//	// 打印序列化后的数据
//	fmt.Printf("Serialized Data: %x\n", serializedData)
//
//	// 反序列化
//	newTx := &Transaction{}
//	deserializedTransaction, err := newTx.Deserialize(serializedData)
//	if err != nil {
//		fmt.Println("Error during deserialization:", err)
//		return
//	}
//
//	// 打印反序列化后的 Transaction
//	fmt.Printf("Deserialized Transaction: %+v\n", deserializedTransaction)
//	fmt.Println(transaction.Hash())
//}
