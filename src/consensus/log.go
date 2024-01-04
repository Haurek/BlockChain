package consensus

import (
	"BlockChain/src/blockchain"
	"sync"
)

// MsgLog represents a cache for consensus messages
type MsgLog struct {
	nodeNum uint64      // pBFT node number
	logs    []*LogEntry // cache of each checkpoint receive message
	lock    sync.Mutex
}

// LogEntry represents a log entry for a checkpoint
// map: node ID -> message
type LogEntry struct {
	prepares map[string]*PrepareMessage    // prepare message cache
	signs    map[string]*SignMessage       // sign message cache
	commits  map[string]*CommitMessage     // commit message cache
	views    map[string]*ViewChangeMessage // view change message cache
	block    *blockchain.Block             // block cache
}

func initEntry() *LogEntry {
	e := &LogEntry{
		prepares: make(map[string]*PrepareMessage),
		signs:    make(map[string]*SignMessage),
		commits:  make(map[string]*CommitMessage),
		views:    make(map[string]*ViewChangeMessage),
		block:    nil,
	}
	return e
}

// NewMsgLog creates and initializes a new MsgLog instance
func NewMsgLog(num uint64) *MsgLog {
	log := &MsgLog{
		nodeNum: num,
		logs:    make([]*LogEntry, num),
	}
	for i := 0; i < int(num); i++ {
		log.logs[i] = initEntry()
	}
	return log
}

// AddMessage adds a message of a specific type to the MsgLog cache
func (l *MsgLog) AddMessage(msgType PBFTMsgType, data interface{}) {
	l.lock.Lock()
	defer l.lock.Unlock()

	switch msgType {
	case PrepareMsg:
		if prepare, ok := data.(PrepareMessage); ok {
			l.logs[prepare.Height%l.nodeNum].prepares[prepare.ID] = &prepare
		}
	case CommitMsg:
		if commit, ok := data.(CommitMessage); ok {
			l.logs[commit.Height%l.nodeNum].commits[commit.ID] = &commit
		}
	case SignMsg:
		if sign, ok := data.(SignMessage); ok {
			l.logs[sign.Height%l.nodeNum].signs[sign.ID] = &sign
		}
	case ViewChangeMsg:
		if view, ok := data.(ViewChangeMessage); ok {
			l.logs[view.Height%l.nodeNum].views[view.ID] = &view
		}
	}
}

// HaveLog checks if a specific type of message with a given ID exists in the MsgLog cache
func (l *MsgLog) HaveLog(msgType PBFTMsgType, id string, height uint64) bool {
	l.lock.Lock()
	defer l.lock.Unlock()

	switch msgType {
	case PrepareMsg:
		_, ok := l.logs[height%l.nodeNum].prepares[id]
		return ok
	case CommitMsg:
		_, ok := l.logs[height%l.nodeNum].commits[id]
		return ok
	case SignMsg:
		_, ok := l.logs[height%l.nodeNum].signs[id]
		return ok
	case ViewChangeMsg:
		_, ok := l.logs[height%l.nodeNum].views[id]
		return ok
	}
	return false
}

// CacheBlock add block into log cache
func (l *MsgLog) CacheBlock(b *blockchain.Block) {
	l.lock.Lock()
	defer l.lock.Unlock()
	height := b.Header.Height
	l.logs[height%l.nodeNum].block = b
}

// GetBlock return log cache block
func (l *MsgLog) GetBlock(height uint64) *blockchain.Block {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.logs[height%l.nodeNum].block
}

// HaveBlock check if a block exists in the MsgLog cache
func (l *MsgLog) HaveBlock(height uint64) bool {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.logs[height%l.nodeNum].block != nil
}

// GetSignLog retrieves a SignMessage from the MsgLog cache using its ID.
func (l *MsgLog) GetSignLog(id string, height uint64) *SignMessage {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.logs[height%l.nodeNum].signs[id]
}

// Count returns the count of messages of a specific type in the MsgLog cache
func (l *MsgLog) Count(msgType PBFTMsgType, height uint64) uint64 {
	l.lock.Lock()
	defer l.lock.Unlock()

	switch msgType {
	case PrepareMsg:
		return uint64(len(l.logs[height%l.nodeNum].prepares))
	case CommitMsg:
		return uint64(len(l.logs[height%l.nodeNum].commits))
	case SignMsg:
		return uint64(len(l.logs[height%l.nodeNum].signs))
	case ViewChangeMsg:
		return uint64(len(l.logs[height%l.nodeNum].views))
	}
	return 0
}

func (l *MsgLog) ClearLog() {
	l.lock.Lock()
	defer l.lock.Unlock()
	for i := 0; i < int(l.nodeNum); i++ {
		l.logs[i] = initEntry()
	}
}
