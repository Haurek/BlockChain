package consensus

import (
	"BlockChain/src/blockchain"
	"encoding/hex"
	"sync"
)

// MsgLog represents a cache for consensus messages
// consensus msg map: node id -> receive message
// block map: block hash -> block
// view change: [to view](node id -> receive message)
type MsgLog struct {
	prepares map[string]*PrepareMessage      // prepare message cache
	signs    map[string]*SignMessage         // sign message cache
	commits  map[string]*CommitMessage       // commit message cache
	views    []map[string]*ViewChangeMessage // view change message cache
	block    map[string]*blockchain.Block    // block cache
	lock     sync.Mutex
}

// NewMsgLog creates and initializes a new MsgLog instance.
func NewMsgLog() *MsgLog {
	log := &MsgLog{
		prepares: make(map[string]*PrepareMessage),
		commits:  make(map[string]*CommitMessage),
		signs:    make(map[string]*SignMessage),
		views:    make([]map[string]*ViewChangeMessage, 256),
		block:    make(map[string]*blockchain.Block),
	}
	return log
}

// AddMessage adds a message of a specific type to the MsgLog cache.
func (l *MsgLog) AddMessage(msgType PBFTMsgType, data interface{}) {
	l.lock.Lock()
	defer l.lock.Unlock()

	switch msgType {
	case PrepareMsg:
		if prepare, ok := data.(PrepareMessage); ok {
			id := prepare.ID
			l.prepares[id] = &prepare
		}
	case CommitMsg:
		if commit, ok := data.(CommitMessage); ok {
			id := commit.ID
			l.commits[id] = &commit
		}
	case SignMsg:
		if sign, ok := data.(SignMessage); ok {
			id := sign.ID
			l.signs[id] = &sign
		}
	case ViewChangeMsg:
		if view, ok := data.(ViewChangeMessage); ok {
			id := view.ID
			to := view.ToView
			l.views[to][id] = &view
		}
	}
}

// HaveLog checks if a specific type of message with a given ID exists in the MsgLog cache.
func (l *MsgLog) HaveLog(msgType PBFTMsgType, id string) bool {
	l.lock.Lock()
	defer l.lock.Unlock()

	switch msgType {
	case PrepareMsg:
		_, ok := l.prepares[id]
		return ok
	case CommitMsg:
		_, ok := l.commits[id]
		return ok
	case SignMsg:
		_, ok := l.signs[id]
		return ok
	}
	return false
}

// CacheBlock add block into log cache
func (l *MsgLog) CacheBlock(b *blockchain.Block) {
	l.lock.Lock()
	defer l.lock.Unlock()
	hash := hex.EncodeToString(b.Header.Hash)
	l.block[hash] = b
}

// GetBlock return log cache block
func (l *MsgLog) GetBlock(h []byte) *blockchain.Block {
	l.lock.Lock()
	defer l.lock.Unlock()
	hash := hex.EncodeToString(h)
	return l.block[hash]
}

// HaveBlock check if a block exists in the MsgLog cache
func (l *MsgLog) HaveBlock(h []byte) bool {
	l.lock.Lock()
	defer l.lock.Unlock()
	hash := hex.EncodeToString(h)
	_, exists := l.block[hash]
	return exists
}

// GetPrepareLog retrieves a PrepareMessage from the MsgLog cache using its ID.
func (l *MsgLog) GetPrepareLog(id string) *PrepareMessage {
	l.lock.Lock()
	defer l.lock.Unlock()

	return l.prepares[id]
}

// GetCommitLog retrieves a CommitMessage from the MsgLog cache using its ID.
func (l *MsgLog) GetCommitLog(id string) *CommitMessage {
	l.lock.Lock()
	defer l.lock.Unlock()

	return l.commits[id]
}

// GetSignLog retrieves a SignMessage from the MsgLog cache using its ID.
func (l *MsgLog) GetSignLog(id string) *SignMessage {
	l.lock.Lock()
	defer l.lock.Unlock()

	return l.signs[id]
}

// GetViewChangeLog retrieves a ViewChangeMessage from the MsgLog cache using its ID.
func (l *MsgLog) GetViewChangeLog(to uint64, id string) *ViewChangeMessage {
	l.lock.Lock()
	defer l.lock.Unlock()

	return l.views[to][id]
}

// Count returns the count of messages of a specific type in the MsgLog cache.
func (l *MsgLog) Count(msgType PBFTMsgType) uint64 {
	l.lock.Lock()
	defer l.lock.Unlock()

	switch msgType {
	case PrepareMsg:
		return uint64(len(l.prepares))
	case CommitMsg:
		return uint64(len(l.commits))
	case SignMsg:
		return uint64(len(l.signs))
	}
	return 0
}

// ViewChangeCount returns the count of ViewChangeMessage in the MsgLog cache.
func (l *MsgLog) ViewChangeCount(to uint64) uint64 {
	l.lock.Lock()
	defer l.lock.Unlock()

	return uint64(len(l.views[to]))
}

func (l *MsgLog) ClearLog() {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.prepares = make(map[string]*PrepareMessage)
	l.signs = make(map[string]*SignMessage)
	l.commits = make(map[string]*CommitMessage)
	l.views = make([]map[string]*ViewChangeMessage, 256)
	l.block = make(map[string]*blockchain.Block)
}
