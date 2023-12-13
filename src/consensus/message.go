package consensus

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
	Data interface{} `json:"data"`
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
		data, ok := m.Data.(map[string]interface{})
		if !ok {
			return nil, DefaultMsg
		}
		reMsg := RequestMessage{
			Timestamp: int64(data["view"].(float64)),
			ClientID:  data["clientID"].(string),
			TxsBytes:  []byte(data["txsBytes"].(string)),
		}
		return reMsg, RequestMsg
	case PrePrepareMsg:
		data, ok := m.Data.(map[string]interface{})
		if !ok {
			return nil, DefaultMsg
		}
		ppMsg := PrePrepareMessage{
			View:   uint64(data["view"].(float64)),
			SeqNum: uint64(data["seqNum"].(float64)),
			Digest: []byte(data["digest"].(string)),
			ReqMsg: []byte(data["reqMsg"].(string)),
			Sign:   []byte(data["sign"].(string)),
			PubKey: []byte(data["pubKey"].(string)),
		}
		return ppMsg, PrePrepareMsg

	case PrepareMsg:
		data, ok := m.Data.(map[string]interface{})
		if !ok {
			return nil, DefaultMsg
		}
		pMsg := PrepareMessage{
			View:      uint64(data["view"].(float64)),
			SeqNum:    uint64(data["seqNum"].(float64)),
			Digest:    []byte(data["digest"].(string)),
			ReplicaID: data["replicaID"].(string),
			Sign:      []byte(data["sign"].(string)),
			PubKey:    []byte(data["pubKey"].(string)),
		}
		return pMsg, PrepareMsg

	case CommitMsg:
		data, ok := m.Data.(map[string]interface{})
		if !ok {
			return nil, DefaultMsg
		}
		cMsg := CommitMessage{
			View:      uint64(data["view"].(float64)),
			SeqNum:    uint64(data["seqNum"].(float64)),
			Digest:    []byte(data["digest"].(string)),
			ReplicaID: data["replicaID"].(string),
			Sign:      []byte(data["sign"].(string)),
			PubKey:    []byte(data["pubKey"].(string)),
		}
		return cMsg, CommitMsg
	case ReplyMsg:
		data, ok := m.Data.(map[string]interface{})
		if !ok {
			return nil, DefaultMsg
		}
		rMsg := ReplyMessage{
			View:      uint64(data["view"].(float64)),
			Timestamp: int64(data["view"].(float64)),
			ReplicaID: data["replicaID"].(string),
		}
		return rMsg, ReplyMsg
	case CheckPointMsg:
		data, ok := m.Data.(map[string]interface{})
		if !ok {
			return nil, DefaultMsg
		}
		cpMsg := CheckPointMessage{
			SeqNum:    uint64(data["seqNum"].(float64)),
			Digest:    []byte(data["digest"].(string)),
			ReplicaID: data["replicaID"].(string),
		}
		return cpMsg, CheckPointMsg

	case ViewChangeMsg:
		data, ok := m.Data.(map[string]interface{})
		if !ok {
			return nil, DefaultMsg
		}
		vcMsg := ViewChangeMessage{
			NewView:          uint64(data["newView"].(float64)),
			StableCheckPoint: uint64(data["stableCheckPoint"].(float64)),
			ReplicaID:        data["replicaID"].(string),
		}

		checkpointSet, ok := data["checkPointSet"].(map[string]interface{})
		if !ok {
			return nil, DefaultMsg
		}
		vcMsg.CheckPointSet = make(map[string]uint64)
		for k, v := range checkpointSet {
			seqNum, ok := v.(float64)
			if !ok {
				return nil, DefaultMsg
			}
			vcMsg.CheckPointSet[k] = uint64(seqNum)
		}

		prepareMsgSet, ok := data["prepareMsgSet"].(map[string]interface{})
		if !ok {
			return nil, DefaultMsg
		}
		vcMsg.PrepareMsgSet = make(map[uint64]PrepareMessage)
		for _, v := range prepareMsgSet {
			seqNum, ok := v.(map[string]interface{})
			if !ok {
				return nil, DefaultMsg
			}
			pMsg := PrepareMessage{
				View:      uint64(seqNum["view"].(float64)),
				SeqNum:    uint64(seqNum["seqNum"].(float64)),
				Digest:    []byte(seqNum["digest"].(string)),
				ReplicaID: seqNum["replicaID"].(string),
			}
			vcMsg.PrepareMsgSet[pMsg.SeqNum] = pMsg
		}

		return vcMsg, ViewChangeMsg

	case ViewChangeAckMsg:
		// ViewChangeAckMessage doesn't have additional fields
		return ViewChangeAckMessage{}, ViewChangeAckMsg

	case NewViewMsg:
		data, ok := m.Data.(map[string]interface{})
		if !ok {
			return nil, DefaultMsg
		}
		nvMsg := NewViewMessage{
			NewView: uint64(data["newView"].(float64)),
		}

		viewChangeSet, ok := data["viewChangeSet"].(map[string]interface{})
		if !ok {
			return nil, DefaultMsg
		}
		nvMsg.ViewChangeSet = make(map[string]ViewChangeMessage)
		for k, v := range viewChangeSet {
			viewChangeData, ok := v.(map[string]interface{})
			if !ok {
				return nil, DefaultMsg
			}

			viewChangeMsg := ViewChangeMessage{
				NewView:          uint64(viewChangeData["newView"].(float64)),
				StableCheckPoint: uint64(viewChangeData["stableCheckPoint"].(float64)),
				ReplicaID:        viewChangeData["replicaID"].(string),
			}

			checkpointSet, ok := viewChangeData["checkPointSet"].(map[string]interface{})
			if !ok {
				return nil, DefaultMsg
			}
			viewChangeMsg.CheckPointSet = make(map[string]uint64)
			for ck, cv := range checkpointSet {
				seqNum, ok := cv.(float64)
				if !ok {
					return nil, DefaultMsg
				}
				viewChangeMsg.CheckPointSet[ck] = uint64(seqNum)
			}

			prepareMsgSet, ok := viewChangeData["prepareMsgSet"].(map[string]interface{})
			if !ok {
				return nil, DefaultMsg
			}
			viewChangeMsg.PrepareMsgSet = make(map[uint64]PrepareMessage)
			for _, pv := range prepareMsgSet {
				seqNum, ok := pv.(map[string]interface{})
				if !ok {
					return nil, DefaultMsg
				}
				pMsg := PrepareMessage{
					View:      uint64(seqNum["view"].(float64)),
					SeqNum:    uint64(seqNum["seqNum"].(float64)),
					Digest:    []byte(seqNum["digest"].(string)),
					ReplicaID: seqNum["replicaID"].(string),
				}
				viewChangeMsg.PrepareMsgSet[pMsg.SeqNum] = pMsg
			}

			nvMsg.ViewChangeSet[k] = viewChangeMsg
		}

		oldSet, ok := data["oldSet"].(map[string]interface{})
		if !ok {
			return nil, DefaultMsg
		}
		nvMsg.OldSet = make(map[uint64]PrepareMessage)
		for _, v := range oldSet {
			seqNum, ok := v.(map[string]interface{})
			if !ok {
				return nil, DefaultMsg
			}
			pMsg := PrepareMessage{
				View:      uint64(seqNum["view"].(float64)),
				SeqNum:    uint64(seqNum["seqNum"].(float64)),
				Digest:    []byte(seqNum["digest"].(string)),
				ReplicaID: seqNum["replicaID"].(string),
			}
			nvMsg.OldSet[pMsg.SeqNum] = pMsg
		}

		return nvMsg, NewViewMsg

	default:
		return nil, DefaultMsg
	}
}
