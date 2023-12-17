package pool

import (
	"fmt"
	"testing"
)

func TestCreateBlockMessage(t *testing.T) {
	blockMessage, err := CreateBlockMessage(SyncRequestMsg, "123", uint64(1))
	if err != nil {
		fmt.Println(err)
	}
	if _, ok := blockMessage.(BlockMessage); ok {
		fmt.Println(true)
	}
}
