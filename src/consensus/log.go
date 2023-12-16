package consensus

import "sync"

// MsgLog save receive message for each seqnum
// @param Cap: max size of MsgLog
// @param Logs: seqnum->entry
type MsgLog struct {
	Cap  uint64
	Logs []*logEntry
	lock sync.Mutex
}

// logEntry entry of MsgLog
type logEntry struct {
	used            bool
	requestLog      *RequestMessage
	prePrepareLog   *PrePrepareMessage
	prepareLog      map[string]*PrepareMessage
	commitLog       map[string]*CommitMessage
	checkPointLog   map[string]*CheckPointMessage
	prepareCount    uint64
	commitCount     uint64
	checkPointCount uint64
	replyCount      uint64
}

func (entry *logEntry) initEntry() {
	entry.used = false
	entry.prepareCount = 0
	entry.commitCount = 0
	entry.checkPointCount = 0
	entry.replyCount = 0
	entry.prePrepareLog = nil
	entry.requestLog = nil
	entry.prepareLog = make(map[string]*PrepareMessage)
	entry.commitLog = make(map[string]*CommitMessage)
	entry.checkPointLog = make(map[string]*CheckPointMessage)
}

func (entry *logEntry) clearEntry() {
	entry.used = false
	entry.prepareCount = 0
	entry.commitCount = 0
	entry.checkPointCount = 0
	entry.replyCount = 0
	entry.requestLog = nil
	entry.prePrepareLog = nil
	clear(entry.prepareLog)
	clear(entry.commitLog)
	clear(entry.checkPointLog)
}

func NewMsgLog(cap uint64) *MsgLog {
	log := &MsgLog{
		Cap:  cap,
		Logs: make([]*logEntry, cap),
	}
	return log
}

func (l *MsgLog) AddMessage(seqnum uint64, msgType PBFTMsgType, data interface{}) {
	l.lock.Lock()
	defer l.lock.Unlock()

	// check entry in used
	if !l.Logs[seqnum].used {
		l.Logs[seqnum].used = true
	}
	switch msgType {
	case RequestMsg:
		if request, ok := data.(RequestMessage); ok {
			l.Logs[seqnum].requestLog = &request
		}
	case PrePrepareMsg:
		if prePrepare, ok := data.(PrePrepareMessage); ok {
			l.Logs[seqnum].prePrepareLog = &prePrepare
		}
	case PrepareMsg:
		if prepare, ok := data.(PrepareMessage); ok {
			id := prepare.ReplicaID
			l.Logs[seqnum].prepareLog[id] = &prepare
			l.Logs[seqnum].prepareCount++
		}
	case CommitMsg:
		if commit, ok := data.(CommitMessage); ok {
			id := commit.ReplicaID
			l.Logs[seqnum].commitLog[id] = &commit
			l.Logs[seqnum].commitCount++
		}
	case CheckPointMsg:
		if checkpoint, ok := data.(CheckPointMessage); ok {
			id := checkpoint.ReplicaID
			l.Logs[seqnum].checkPointLog[id] = &checkpoint
			l.Logs[seqnum].checkPointCount++
		}
	case ReplyMsg:
		if _, ok := data.(ReplyMessage); ok {
			l.Logs[seqnum].replyCount++
		}
	}
}

func (l *MsgLog) HaveLog(seqnum uint64, msgType PBFTMsgType, id string) bool {
	l.lock.Lock()
	defer l.lock.Unlock()

	// check entry in used
	if !l.Logs[seqnum].used {
		return false
	}
	switch msgType {
	case PrepareMsg:
		_, ok := l.Logs[seqnum].prepareLog[id]
		return ok
	case CommitMsg:
		_, ok := l.Logs[seqnum].commitLog[id]
		return ok
	case CheckPointMsg:
		_, ok := l.Logs[seqnum].checkPointLog[id]
		return ok
	}
	return false
}

func (l *MsgLog) GetPrePrepareLog(seqnum uint64) *PrePrepareMessage {
	l.lock.Lock()
	defer l.lock.Unlock()

	return l.Logs[seqnum].prePrepareLog
}

func (l *MsgLog) GetRequestLog(seqnum uint64) *RequestMessage {
	l.lock.Lock()
	defer l.lock.Unlock()

	return l.Logs[seqnum].requestLog
}

func (l *MsgLog) GetPrepareLog(seqnum uint64, id string) *PrepareMessage {
	l.lock.Lock()
	defer l.lock.Unlock()

	return l.Logs[seqnum].prepareLog[id]
}

func (l *MsgLog) GetLogCount(seqnum uint64, msgType PBFTMsgType) uint64 {
	l.lock.Unlock()
	defer l.lock.Unlock()
	if !l.Logs[seqnum].used {
		return 0
	}
	switch msgType {
	case PrepareMsg:
		return l.Logs[seqnum].prepareCount
	case CommitMsg:
		return l.Logs[seqnum].commitCount
	case CheckPointMsg:
		return l.Logs[seqnum].checkPointCount
	case ReplyMsg:
		return l.Logs[seqnum].replyCount
	}
	return 0
}

func (l *MsgLog) ClearLogs() {
	l.lock.Lock()
	defer l.lock.Unlock()

	// clear all entry
	for _, log := range l.Logs {
		log.clearEntry()
	}
}
