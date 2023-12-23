package consensus

import "sync"

// MsgLog represents a cache for consensus messages.
type MsgLog struct {
	prepares map[string]*PrepareMessage
	signs    map[string]*SignMessage
	commits  map[string]*CommitMessage
	views    map[string]*ViewChangeMessage
	lock     sync.Mutex
}

// NewMsgLog creates and initializes a new MsgLog instance.
func NewMsgLog() *MsgLog {
	log := &MsgLog{
		prepares: make(map[string]*PrepareMessage),
		commits:  make(map[string]*CommitMessage),
		signs:    make(map[string]*SignMessage),
		views:    make(map[string]*ViewChangeMessage),
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
			l.views[id] = &view
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
	case ViewChangeMsg:
		_, ok := l.views[id]
		return ok
	}
	return false
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
func (l *MsgLog) GetViewChangeLog(id string) *ViewChangeMessage {
	l.lock.Lock()
	defer l.lock.Unlock()

	return l.views[id]
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
	case ViewChangeMsg:
		return uint64(len(l.views))
	case SignMsg:
		return uint64(len(l.signs))
	}
	return 0
}
