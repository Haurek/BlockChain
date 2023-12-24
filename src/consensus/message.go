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

// PBFTMessage type
type PBFTMessage struct {
	Type PBFTMsgType `json:"type"`
	Data []byte      `json:"data"`
}

// PrepareMessage pBFT prepare message
type PrepareMessage struct {
	ID        string `json:"id"`        // sender ID
	Height    uint64 `json:"height"`    // pack block height
	BlockHash []byte `json:"blockHash"` // pack block hash
	Block     []byte `json:"block"`     // pack block data
	View      uint64 `json:"view"`      // current view
	Sign      []byte `json:"sign"`      // signature
	PubKey    []byte `json:"pubKey"`    // sender public key
}

// SignMessage pBFT sign message
type SignMessage struct {
	ID        string `json:"id"`        // sender ID
	Height    uint64 `json:"height"`    // signed block height
	BlockHash []byte `json:"blockHash"` // signed block hash
	View      uint64 `json:"view"`      // current view
	Sign      []byte `json:"sign"`      // signature
	PubKey    []byte `json:"pubKey"`    // sender public key
}

// CommitMessage pBFT commit message
type CommitMessage struct {
	ID        string `json:"id"`        // sender ID
	Height    uint64 `json:"height"`    // commit block height
	BlockHash []byte `json:"blockHash"` // commit block hash
	View      uint64 `json:"view"`      // current view
	Sign      []byte `json:"sign"`      // signature
	PubKey    []byte `json:"pubKey"`    //sender pbulic key
}

// ViewChangeMessage pBFT view change message
type ViewChangeMessage struct {
	ID        string `json:"id"`        // sender ID
	Height    uint64 `json:"height"`    // sender chain height
	BlockHash []byte `json:"blockHash"` // sender best block hash
	View      uint64 `json:"view"`      // current view
	ToView    uint64 `json:"toView"`    // change to view
	Sign      []byte `json:"sign"`      // signature
	PubKey    []byte `json:"pubKey"`    // sender public key
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
