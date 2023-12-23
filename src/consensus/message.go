package consensus

import "encoding/json"

type PBFTMsgType int32

const (
	DefaultMsg PBFTMsgType = iota
	SignMsg
	PrepareMsg
	CommitMsg
	ViewChangeMsg
)

type PBFTMessage struct {
	Type PBFTMsgType `json:"type"`
	Data []byte      `json:"data"`
}

// PrepareMessage pBFT prepare message
// @param ReplicaID: ID of replica peer who send this message
type PrepareMessage struct {
	ID        string `json:"id"`
	Height    uint64 `json:"height"`
	BlockHash []byte `json:"blockHash"`
	Block     []byte `json:"block"`
	View      uint64 `json:"view"`
	Sign      []byte `json:"sign"`
	PubKey    []byte `json:"pubKey"`
}

type SignMessage struct {
	ID        string `json:"id"`
	Height    uint64 `json:"height"`
	BlockHash []byte `json:"blockHash"`
	View      uint64 `json:"view"`
	Sign      []byte `json:"sign"`
	PubKey    []byte `json:"pubKey"`
}

// CommitMessage pBFT prepare message
// @param ReplicaID: ID of replica peer who send this message
type CommitMessage struct {
	ID        string `json:"id"`
	Height    uint64 `json:"height"`
	BlockHash []byte `json:"blockHash"`
	View      uint64 `json:"view"`
	Sign      []byte `json:"sign"`
	PubKey    []byte `json:"pubKey"`
}

// ViewChangeMessage sent when an error occurs on the primary node
// @param NewView: View + 1
// @param StableCheckPoint: stable point of replica
// @param CheckPointSet: 2f + 1 valid node checkpoint set
// @param PrepareMsgSet: set of request messages numbered greater than n in the previous view in node i that have reached prepared status
type ViewChangeMessage struct {
	ID string `json:"id"`
}

// SplitMessage splits the PBFTMessage into the corresponding message struct based on its type.
func (m *PBFTMessage) SplitMessage() (interface{}, PBFTMsgType) {
	switch m.Type {
	case SignMsg:
		var sMsg SignMessage
		err := json.Unmarshal(m.Data, &sMsg)
		if err != nil {
			return nil, DefaultMsg // Return default message type on unmarshal error
		}
		return sMsg, SignMsg // Return the SignMessage and its corresponding message type

	case PrepareMsg:
		var pMsg PrepareMessage
		err := json.Unmarshal(m.Data, &pMsg)
		if err != nil {
			return nil, DefaultMsg // Return default message type on unmarshal error
		}
		return pMsg, PrepareMsg // Return the PrepareMessage and its corresponding message type

	case CommitMsg:
		var cMsg CommitMessage
		err := json.Unmarshal(m.Data, &cMsg)
		if err != nil {
			return nil, DefaultMsg // Return default message type on unmarshal error
		}
		return cMsg, CommitMsg // Return the CommitMessage and its corresponding message type

	case ViewChangeMsg:
		var vcMsg ViewChangeMessage
		err := json.Unmarshal(m.Data, &vcMsg)
		if err != nil {
			return nil, DefaultMsg // Return default message type on unmarshal error
		}
		return vcMsg, ViewChangeMsg // Return the ViewChangeMessage and its corresponding message type

	default:
		return nil, DefaultMsg // Return default message type for unknown PBFTMsgType
	}
}
