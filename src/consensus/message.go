package consensus

import "encoding/json"

type PBFTMsgType int32

const (
	DefaultMsg PBFTMsgType = iota
	RequestMsg
	PrePrepareMsg
	PrepareMsg
	CommitMsg
	ReplyMsg
	CheckPointMsg
	ViewChangeMsg
	ViewChangeAckMsg
	NewViewMsg
)

type PBFTMessage struct {
	Type PBFTMsgType `json:"type"`
	Data []byte      `json:"data"`
}

// RequestMessage primary node proposal new block
type RequestMessage struct {
	Timestamp int64  `json:"timestamp"`
	ClientID  string `json:"clientID"`
	TxsBytes  []byte `json:"txsBytes"`
}

// PrePrepareMessage pBFT pre-prepare message
// @param Msg: is origin data from client request
type PrePrepareMessage struct {
	View   uint64 `json:"view"`
	SeqNum uint64 `json:"seqNum"`
	Digest []byte `json:"digest"`
	ReqMsg []byte `json:"reqMsg"`
	Sign   []byte `json:"sign"`
	PubKey []byte `json:"pubKey"`
}

// PrepareMessage pBFT prepare message
// @param ReplicaID: ID of replica peer who send this message
type PrepareMessage struct {
	View      uint64 `json:"view"`
	SeqNum    uint64 `json:"seqNum"`
	Digest    []byte `json:"digest"`
	ReplicaID string `json:"replicaID"`
	Sign      []byte `json:"sign"`
	PubKey    []byte `json:"pubKey"`
}

// CommitMessage pBFT prepare message
// @param ReplicaID: ID of replica peer who send this message
type CommitMessage struct {
	View      uint64 `json:"view"`
	SeqNum    uint64 `json:"seqNum"`
	Digest    []byte `json:"digest"`
	ReplicaID string `json:"replicaID"`
	Sign      []byte `json:"sign"`
	PubKey    []byte `json:"pubKey"`
}

type ReplyMessage struct {
	View      uint64 `json:"view"`
	Timestamp int64  `json:"timestamp"`
	ClientID  string `json:"clientID"`
	ReplicaID string `json:"replicaID"`
}

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

// SplitMessage spilt PBFTMessage into the message struct corresponding to its type
func (m *PBFTMessage) SplitMessage() (interface{}, PBFTMsgType) {
	switch m.Type {
	case RequestMsg:
		var reMsg RequestMessage
		err := json.Unmarshal(m.Data, &reMsg)
		if err != nil {
			return nil, DefaultMsg
		}
		return reMsg, RequestMsg
	case PrePrepareMsg:
		var ppMsg PrePrepareMessage
		err := json.Unmarshal(m.Data, &ppMsg)
		if err != nil {
			return nil, DefaultMsg
		}
		return ppMsg, PrePrepareMsg

	case PrepareMsg:
		var pMsg PrepareMessage
		err := json.Unmarshal(m.Data, &pMsg)
		if err != nil {
			return nil, DefaultMsg
		}
		return pMsg, PrepareMsg

	case CommitMsg:
		var cMsg CommitMessage
		err := json.Unmarshal(m.Data, &cMsg)
		if err != nil {
			return nil, DefaultMsg
		}
		return cMsg, CommitMsg
	case ReplyMsg:
		var rMsg RequestMessage
		err := json.Unmarshal(m.Data, &rMsg)
		if err != nil {
			return nil, DefaultMsg
		}
		return rMsg, ReplyMsg
	case CheckPointMsg:
		var cpMsg CheckPointMessage
		err := json.Unmarshal(m.Data, &cpMsg)
		if err != nil {
			return nil, DefaultMsg
		}
		return cpMsg, CheckPointMsg

	case ViewChangeMsg:
		var vcMsg ViewChangeMessage
		err := json.Unmarshal(m.Data, &vcMsg)
		if err != nil {
			return nil, DefaultMsg
		}
		return vcMsg, ViewChangeMsg

	case ViewChangeAckMsg:
		// ViewChangeAckMessage doesn't have additional fields
		return ViewChangeAckMessage{}, ViewChangeAckMsg

	case NewViewMsg:
		var nvMsg NewViewMessage
		err := json.Unmarshal(m.Data, &nvMsg)
		if err != nil {
			return nil, DefaultMsg
		}
		return nvMsg, NewViewMsg

	default:
		return nil, DefaultMsg
	}
}
