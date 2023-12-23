package consensus

import "sync"

// MsgLog consensus message cache
type MsgLog struct {
	prepares map[string]*PrepareMessage
	signs    map[string]*SignMessage
	commits  map[string]*CommitMessage
	views    map[string]*ViewChangeMessage
	lock     sync.Mutex
}

func NewMsgLog() *MsgLog {
	log := &MsgLog{
		prepares: make(map[string]*PrepareMessage),
		commits:  make(map[string]*CommitMessage),
		signs:    make(map[string]*SignMessage),
		views:    make(map[string]*ViewChangeMessage),
	}
	return log
}

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

func (l *MsgLog) GetPrepareLog(id string) *PrepareMessage {
	l.lock.Lock()
	defer l.lock.Unlock()

	return l.prepares[id]
}

func (l *MsgLog) GetCommitLog(id string) *CommitMessage {
	l.lock.Lock()
	defer l.lock.Unlock()

	return l.commits[id]
}

func (l *MsgLog) GetSignLog(id string) *SignMessage {
	l.lock.Lock()
	defer l.lock.Unlock()

	return l.signs[id]
}

func (l *MsgLog) GetViewChangeLog(id string) *ViewChangeMessage {
	l.lock.Lock()
	defer l.lock.Unlock()

	return l.views[id]
}

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
