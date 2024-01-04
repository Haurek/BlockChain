package consensus

import (
	"fmt"
	"testing"
)

func TestMsgLog(t *testing.T) {
	log := NewMsgLog(4)
	testPrepareMessage := PrepareMessage{
		ID:        "test1",
		Height:    0,
		BlockHash: []byte("test"),
		Block:     []byte("block"),
		View:      0,
		Sign:      []byte("sign"),
		PubKey:    []byte("key"),
	}
	testCommitMessage := CommitMessage{
		ID:        "test2",
		Height:    0,
		BlockHash: []byte("test"),
		View:      0,
		Sign:      []byte("sign"),
		PubKey:    []byte("key"),
	}
	testCommitMessage2 := CommitMessage{
		ID:        "test3",
		Height:    0,
		BlockHash: []byte("test"),
		View:      0,
		Sign:      []byte("sign"),
		PubKey:    []byte("key"),
	}
	testCommitMessage3 := CommitMessage{
		ID:        "test4",
		Height:    0,
		BlockHash: []byte("test"),
		View:      0,
		Sign:      []byte("sign"),
		PubKey:    []byte("key"),
	}
	testSignMessage := SignMessage{
		ID:        "test3",
		Height:    0,
		BlockHash: []byte("test"),
		View:      0,
		Sign:      []byte("sign"),
		PubKey:    []byte("key"),
	}
	anotherPrepareMessage := PrepareMessage{
		ID:        "test1",
		Height:    1,
		BlockHash: []byte("test"),
		Block:     []byte("block"),
		View:      0,
		Sign:      []byte("sign"),
		PubKey:    []byte("key"),
	}
	log.AddMessage(PrepareMsg, testPrepareMessage)
	log.AddMessage(CommitMsg, testCommitMessage)
	log.AddMessage(CommitMsg, testCommitMessage2)
	log.AddMessage(CommitMsg, testCommitMessage3)
	log.AddMessage(SignMsg, testSignMessage)
	fmt.Printf("have message: %v\n", log.HaveLog(PrepareMsg, testPrepareMessage.ID, testPrepareMessage.Height))
	fmt.Printf("Commit count: %d\n", log.Count(CommitMsg, testCommitMessage.Height))
	fmt.Printf("Sign log: %v\n", log.GetSignLog(testSignMessage.ID, testSignMessage.Height))

	fmt.Printf("have checkpoint %d prepare message: %v\n", anotherPrepareMessage.Height, log.HaveLog(PrepareMsg, anotherPrepareMessage.ID, anotherPrepareMessage.Height))
	log.AddMessage(PrepareMsg, anotherPrepareMessage)
	fmt.Printf("prepare count: %d\n", log.Count(PrepareMsg, anotherPrepareMessage.Height))
	log.ClearLog()
	fmt.Printf("Commit count: %d\n", log.Count(CommitMsg, testCommitMessage.Height))

}
