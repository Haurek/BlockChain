package BlockChain

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestSerialize(t *testing.T) {
	block := NewGenesisBlock([]byte("asdasd"))
	data, err := Serialize(block)
	HandleError(err)
	var newblock Block
	err = Deserialize(data, &newblock)
	HandleError(err)
	fmt.Println("block preHash: ", hex.EncodeToString(block.Header.PrevHash))
	fmt.Println("new block preHash: ", hex.EncodeToString(newblock.Header.PrevHash))

}
