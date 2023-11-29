package BlockChain

import (
	"encoding/hex"
	"fmt"
	"log"
	"testing"
)

func TestBase58(t *testing.T) {
	rawHash := "00010966776006953D5567439E5E39F86A0D273BEED61967F6"
	hash, err := hex.DecodeString(rawHash)
	if err != nil {
		log.Fatal(err)
	}

	encoded := Base58Encode(hash)
	fmt.Println("Encode: ", string(encoded))
	decoded := Base58Decode([]byte("16UwLL9Risc3QfPqBUvKofHmBQ7wMtjvM"))
	fmt.Println("Decode: ", decoded)
}
