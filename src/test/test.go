package main

import (
	"encoding/json"
	"fmt"
)

// TxMsgType type of transaction message
type TxMsgType int32
type BlockMsgType int32

const (
	SendTxMsg TxMsgType = iota
)

// TxMessage is type of transaction message
type TxMessage struct {
	Type    TxMsgType `json:"type"`
	TxBytes []byte    `json:"tx"`
}

const (
	DefaultMsg BlockMsgType = iota
	SyncRequestMsg
	SyncResponseMsg
	BlockRequestMsg
	BlockResponseMsg
	NewBlockBroadcastMsg
)

type BlockMessage struct {
	Type BlockMsgType `json:"type"`
	Data []byte       `json:"data"`
}

type SyncRequestMessage struct {
	//Type        BlockMsgType `json:"type"`
	NodeID      string `json:"nodeID"`
	BlockHeight uint64 `json:"blockHeight"`
}

type SyncResponseMessage struct {
	//Type       BlockMsgType `json:"type"`
	FromID     string `json:"fromID"`
	ToID       string `json:"toID"`
	BestHeight uint64 `json:"bestHeight"`
}

type BlockRequestMessage struct {
	//Type   BlockMsgType `json:"type"`
	NodeID string `json:"nodeID"`
	Min    uint64 `json:"min"`
	Max    uint64 `json:"max"`
}

type BlockResponseMessage struct {
	//Type   BlockMsgType `json:"type"`
	FromID string `json:"fromID"`
	ToID   string `json:"toID"`
	Height uint64 `json:"height"`
	Hash   []byte `json:"hash"`
	Block  []byte `json:"block"`
}

type NewBlockMessage struct {
	//Type   BlockMsgType `json:"type"`
	Height uint64 `json:"height"`
	Hash   []byte `json:"hash"`
	Block  []byte `json:"block"`
}

// CreateBlockMessage function
func CreateBlockMessage(t BlockMsgType, data ...interface{}) (interface{}, error) {
	blockMessage := &BlockMessage{
		Type: t,
	}
	var msg interface{}
	var err error
	switch t {
	case SyncRequestMsg:
		msg, err = createSyncRequestMessage(data...)
		if err != nil {
			return nil, err
		}
	case SyncResponseMsg:
		msg, err = createSyncResponseMessage(data...)
		if err != nil {
			return nil, err
		}
	case BlockRequestMsg:
		msg, err = createBlockRequestMessage(data...)
		if err != nil {
			return nil, err
		}
	case BlockResponseMsg:
		msg, err = createBlockResponseMessage(data...)
		if err != nil {
			return nil, err
		}
	case NewBlockBroadcastMsg:
		msg, err = createNewBlockMessage(data...)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown message type: %v", t)
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	blockMessage.Data = payload
	return blockMessage, nil
}

func createSyncRequestMessage(data ...interface{}) (*SyncRequestMessage, error) {
	if len(data) != 2 {
		return nil, fmt.Errorf("invalid number of arguments for SyncRequestMessage")
	}

	nodeID, ok1 := data[0].(string)
	blockHeight, ok2 := data[1].(uint64)

	if !ok1 || !ok2 {
		return nil, fmt.Errorf("invalid argument types for SyncRequestMessage")
	}

	return &SyncRequestMessage{
		NodeID:      nodeID,
		BlockHeight: blockHeight,
	}, nil
}

func createSyncResponseMessage(data ...interface{}) (*SyncResponseMessage, error) {
	if len(data) != 3 {
		return nil, fmt.Errorf("invalid number of arguments for SyncResponseMessage")
	}

	fromID, ok1 := data[0].(string)
	toID, ok2 := data[1].(string)
	bestHeight, ok3 := data[2].(uint64)

	if !ok1 || !ok2 || !ok3 {
		return nil, fmt.Errorf("invalid argument types for SyncResponseMessage")
	}

	return &SyncResponseMessage{
		FromID:     fromID,
		ToID:       toID,
		BestHeight: bestHeight,
	}, nil
}

func createBlockRequestMessage(data ...interface{}) (*BlockRequestMessage, error) {
	if len(data) != 3 {
		return nil, fmt.Errorf("invalid number of arguments for BlockRequestMessage")
	}

	nodeID, ok1 := data[0].(string)
	minimum, ok2 := data[1].(uint64)
	maximum, ok3 := data[2].(uint64)

	if !ok1 || !ok2 || !ok3 {
		return nil, fmt.Errorf("invalid argument types for BlockRequestMessage")
	}

	return &BlockRequestMessage{
		NodeID: nodeID,
		Min:    minimum,
		Max:    maximum,
	}, nil
}

func createBlockResponseMessage(data ...interface{}) (*BlockResponseMessage, error) {
	if len(data) != 5 {
		return nil, fmt.Errorf("invalid number of arguments for BlockResponseMessage")
	}

	fromID, ok1 := data[0].(string)
	toID, ok2 := data[1].(string)
	height, ok3 := data[2].(uint64)
	hash, ok4 := data[3].([]byte)
	block, ok5 := data[4].([]byte)

	if !ok1 || !ok2 || !ok3 || !ok4 || !ok5 {
		return nil, fmt.Errorf("invalid argument types for BlockResponseMessage")
	}

	return &BlockResponseMessage{
		FromID: fromID,
		ToID:   toID,
		Height: height,
		Hash:   hash,
		Block:  block,
	}, nil
}

func createNewBlockMessage(data ...interface{}) (*NewBlockMessage, error) {
	if len(data) != 3 {
		return nil, fmt.Errorf("invalid number of arguments for NewBlockMessage")
	}

	height, ok1 := data[0].(uint64)
	hash, ok2 := data[1].([]byte)
	block, ok3 := data[2].([]byte)

	if !ok1 || !ok2 || !ok3 {
		return nil, fmt.Errorf("invalid argument types for NewBlockMessage")
	}

	return &NewBlockMessage{
		Height: height,
		Hash:   hash,
		Block:  block,
	}, nil
}

// SplitMessage spilt PBFTMessage into the message struct corresponding to its type
func (m *BlockMessage) SplitMessage() (interface{}, BlockMsgType) {
	switch m.Type {
	case SyncRequestMsg:
		var syncReqMsg SyncRequestMessage
		err := json.Unmarshal(m.Data, &syncReqMsg)
		if err != nil {
			return nil, DefaultMsg
		}
		return syncReqMsg, SyncRequestMsg
	case SyncResponseMsg:
		var syncResMsg SyncResponseMessage
		err := json.Unmarshal(m.Data, &syncResMsg)
		if err != nil {
			return nil, DefaultMsg
		}
		return syncResMsg, SyncResponseMsg
	case BlockRequestMsg:
		var blockReqMsg BlockRequestMessage
		err := json.Unmarshal(m.Data, &blockReqMsg)
		if err != nil {
			return nil, DefaultMsg
		}
		return blockReqMsg, BlockRequestMsg
	case BlockResponseMsg:
		var blockResMsg BlockResponseMessage
		err := json.Unmarshal(m.Data, &blockResMsg)
		if err != nil {
			return nil, DefaultMsg
		}
		return blockResMsg, BlockResponseMsg
	case NewBlockBroadcastMsg:
		var newBlockMsg NewBlockMessage
		err := json.Unmarshal(m.Data, &newBlockMsg)
		if err != nil {
			return nil, DefaultMsg
		}
		return newBlockMsg, NewBlockBroadcastMsg
	default:
		return nil, DefaultMsg
	}
}

func main() {
	blockMessage, err := CreateBlockMessage(SyncRequestMsg, "123", uint64(1))
	if err != nil {
		fmt.Println(err)
	}
	data, err := json.Marshal(blockMessage)
	if err != nil {
		fmt.Println(err)
	}
	var newBlockMessage BlockMessage
	err = json.Unmarshal(data, &newBlockMessage)
	msg, _ := newBlockMessage.SplitMessage()
	if m, ok := msg.(SyncRequestMessage); ok {
		fmt.Println(m.NodeID)
		fmt.Println(m.BlockHeight)
	}
}
