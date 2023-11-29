package BlockChain

import (
	"fmt"
	"testing"
)

func TestBlockHeader_Serializea_and_DeSerializea(t *testing.T) {
	header := &BlockHeader{1234, []byte{1, 2}, []byte{1, 2}, 1234, 12, []byte{1, 2}}
	raw := header.Serialize()
	fmt.Println(raw)
	newheader := &BlockHeader{}
	newheader.Deserialize(raw)
	fmt.Println(newheader.Timestamp)
	fmt.Println(newheader.Hash)
}
