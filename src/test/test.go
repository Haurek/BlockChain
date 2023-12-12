package test

import (
	"encoding/json"
	"fmt"
)

type PBFTMsgType int32

const (
	DefaultMsg PBFTMsgType = iota
	PrePrepareMsg
	PrepareMsg
	CommitMsg
	ReplyMsg
	CheckPointMsg
	ViewChangeMsg
	ViewChangeAckMsg
	NewView
)

type PBFTMessage struct {
	Type PBFTMsgType `json:"type"`
	Data interface{} `json:"data"`
}

// PrePrepareMessage pBFT pre-prepare message
// @param Data: is origin data from client request
type PrePrepareMessage struct {
	View   uint64 `json:"view"`
	SeqNum uint64 `json:"seqNum"`
	Digest []byte `json:"digest"`
	Msg    []byte `json:"msg"`
}

// PrepareMessage pBFT prepare message
// @param ReplicaID: ID of replica peer who send this message
type PrepareMessage struct {
	View      uint64 `json:"view"`
	SeqNum    uint64 `json:"seqNum"`
	Digest    []byte `json:"digest"`
	ReplicaID string `json:"replicaID"`
}

// CommitMessage pBFT prepare message
// @param ReplicaID: ID of replica peer who send this message
type CommitMessage struct {
	View      uint64 `json:"view"`
	SeqNum    uint64 `json:"seqNum"`
	Digest    []byte `json:"digest"`
	ReplicaID string `json:"replicaID"`
}

//type ReplyMessage struct {
//	Generic PBFTMessage
//}

// CheckPointMessage send this message when checkpoint condition is met
// @param ReplicaID: ID of replica peer who send this message
type CheckPointMessage struct {
	SeqNum    uint64 `json:"seqNum"`
	Digest    []byte `json:"digest"`
	ReplicaID string `json:"replicaID"`
}

// ViewChangeMessage sent when an error occurs on the primary node
// @param NewView: View + 1
// @param StableCheckPoint: stable point of replica
// @param CheckPointSet: 2f + 1 valid node checkpoint set
// @param PrepareMsgSet: set of request messages numbered greater than n in the previous view in node i that have reached prepared status
type ViewChangeMessage struct {
	NewView          uint64                    `json:"newView"`
	StableCheckPoint uint64                    `json:"stableCheckPoint"`
	CheckPointSet    map[string]uint64         `json:"checkPointSet"`
	PrepareMsgSet    map[uint64]PrepareMessage `json:"prepareMsgSet"`
	ReplicaID        string                    `json:"replicaID"`
}

// ViewChangeAckMessage ack of view-change message
type ViewChangeAckMessage struct {
}

// NewViewMessage new primary node send this message
// @param NewView: View + 1
// @param ViewChangeSet: set of new primary node receive valid view-change message
// @param OldSet:
type NewViewMessage struct {
	NewView       uint64                       `json:"newView"`
	ViewChangeSet map[string]ViewChangeMessage `json:"viewChangeSet"`
	OldSet        map[uint64]PrepareMessage    `json:"oldSet"`
}

func main() {
	// 创建一个 PrePrepareData 结构
	msg := CheckPointMessage{
		SeqNum:    0,
		Digest:    []byte("abca"),
		ReplicaID: "asdad",
	}

	//// 创建一个 PrePrepareMessage
	//prePrepareMessage := PrePrepareMessage{
	//	PrePrepareData: prePrepareData,
	//	// 添加其他 PrePrepareMessage 的字段...
	//}

	// 将 PrePrepareMessage 转换为 PBFTMessage
	pbftMessage := PBFTMessage{
		Type: CheckPointMsg,
		Data: msg,
	}

	// 序列化 PBFTMessage
	serialized, err := json.Marshal(pbftMessage)
	if err != nil {
		return
	}

	fmt.Printf("Serialized PBFTMessage: %s\n", serialized)

	// 反序列化 PBFTMessage
	var deserialized PBFTMessage
	err = json.Unmarshal(serialized, &deserialized)
	if err != nil {
		return
	}
	for key, value := range deserialized.Data.(map[string]interface{}) {
		fmt.Printf("key: %s, value: %v\n", key, value)
	}
}
