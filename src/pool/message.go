package pool

// TxMsgType type of transaction message
type TxMsgType int32
type BlockMsgType int32

const (
	SendTxMsg TxMsgType = iota
)

const (
	DefaultMsg BlockMsgType = iota
	SyncRequestMsg
	SyncResponseMsg
	BlockRequestMsg
	BlockResponseMsg
	NewBlockBroadcastMsg
)

// TxMessage is type of transaction message
type TxMessage struct {
	Type    TxMsgType `json:"type"`
	TxBytes []byte    `json:"tx"`
}

type BlockMessage struct {
	Type BlockMsgType `json:"type"`
	Data interface{}  `json:"data"`
}

type SyncRequestMessage struct {
	NodeID      string `json:"nodeID"`
	BlockHeight uint64 `json:"blockHeight"`
}

type SyncResponseMessage struct {
	FromID     string `json:"fromID"`
	ToID       string `json:"toID"`
	BestHeight uint64 `json:"bestHeight"`
}

type BlockRequestMessage struct {
	NodeID string `json:"nodeID"`
	Min    uint64 `json:"min"`
	Max    uint64 `json:"max"`
}

type BlockResponseMessage struct {
	FromID string `json:"fromID"`
	ToID   string `json:"toID"`
	Height uint64 `json:"height"`
	Hash   []byte `json:"hash"`
	Block  []byte `json:"block"`
}

type NewBlockMessage struct {
	Height uint64 `json:"height"`
	Hash   []byte `json:"hash"`
	Block  []byte `json:"block"`
}

// SplitMessage spilt PBFTMessage into the message struct corresponding to its type
func (m *BlockMessage) SplitMessage() (interface{}, BlockMsgType) {
	switch m.Type {
	case SyncRequestMsg:
		data, ok := m.Data.(map[string]interface{})
		if !ok {
			return nil, DefaultMsg
		}
		syncReqMsg := SyncRequestMessage{
			NodeID:      data["nodeID"].(string),
			BlockHeight: uint64(data["blockHeight"].(float64)),
		}
		return syncReqMsg, SyncRequestMsg
	case SyncResponseMsg:
		data, ok := m.Data.(map[string]interface{})
		if !ok {
			return nil, DefaultMsg
		}
		syncResMsg := SyncResponseMessage{
			FromID:     data["fromID"].(string),
			ToID:       data["toID"].(string),
			BestHeight: uint64(data["bestHeight"].(float64)),
		}
		return syncResMsg, SyncResponseMsg
	case BlockRequestMsg:
		data, ok := m.Data.(map[string]interface{})
		if !ok {
			return nil, DefaultMsg
		}
		blockReqMsg := BlockRequestMessage{
			NodeID: data["nodeID"].(string),
			Min:    uint64(data["min"].(float64)),
			Max:    uint64(data["max"].(float64)),
		}
		return blockReqMsg, BlockRequestMsg
	case BlockResponseMsg:
		data, ok := m.Data.(map[string]interface{})
		if !ok {
			return nil, DefaultMsg
		}
		blockResMsg := BlockResponseMessage{
			FromID: data["fromID"].(string),
			ToID:   data["toID"].(string),
			Height: uint64(data["height"].(float64)),
			Hash:   []byte(data["hash"].(string)),
			Block:  []byte(data["block"].(string)),
		}
		return blockResMsg, BlockResponseMsg
	case NewBlockBroadcastMsg:
		data, ok := m.Data.(map[string]interface{})
		if !ok {
			return nil, DefaultMsg
		}
		newBlockMsg := NewBlockMessage{
			Height: uint64(data["height"].(float64)),
			Hash:   []byte(data["hash"].(string)),
			Block:  []byte(data["block"].(string)),
		}
		return newBlockMsg, NewBlockBroadcastMsg
	default:
		return nil, DefaultMsg
	}
}
