package consensus

import "sync"

// MsgLog save receive message for each seqnum
// @param Cap: max size of MsgLog
// @param Logs: seqnum->entry
type MsgLog struct {
	Cap  int
	Logs []*logEntry
	lock sync.Mutex
}

// logEntry entry of MsgLog
type logEntry struct {
	used            bool
	prepareLog      []*PrepareMessage
	commitLog       []*CommitMessage
	checkPointLog   []*CheckPointMessage
	prepareCount    uint64
	commitCount     uint64
	checkPointCount uint64
}

func (entry *logEntry) clearEntry() {
	entry.used = false
	entry.prepareCount = 0
	entry.commitCount = 0
	entry.checkPointCount = 0
	entry.prepareLog = make([]*PrepareMessage, 10)
	entry.commitLog = make([]*CommitMessage, 10)
	entry.checkPointLog = make([]*CheckPointMessage, 10)
}

func NewMsgLog(cap int) *MsgLog {
	log := &MsgLog{
		Cap:  cap,
		Logs: make([]*logEntry, cap),
	}
	for _, entry := range log.Logs {
		entry.clearEntry()
	}
	return log
}

func (l *MsgLog) AddMessage(seqnum uint64, msg *PBFTMessage) {
	l.lock.Lock()
	defer l.lock.Unlock()

	// check entry in used
	if !l.Logs[seqnum].used {
		l.Logs[seqnum].used = true
	}
	data, msgType := msg.SplitMessage()
	switch msgType {
	case PrepareMsg:
		if prepare, ok := data.(PrepareMessage); ok {
			l.Logs[seqnum].prepareLog = append(l.Logs[seqnum].prepareLog, &prepare)
			l.Logs[seqnum].prepareCount++
		}
	case CommitMsg:
		if commit, ok := data.(CommitMessage); ok {
			l.Logs[seqnum].commitLog = append(l.Logs[seqnum].commitLog, &commit)
			l.Logs[seqnum].commitCount++
		}
	case CheckPointMsg:
		if checkpoint, ok := data.(CheckPointMessage); ok {
			l.Logs[seqnum].checkPointLog = append(l.Logs[seqnum].checkPointLog, &checkpoint)
			l.Logs[seqnum].checkPointCount++
		}
	}
}

func (l *MsgLog) CleanLogs() {
	l.lock.Lock()
	defer l.lock.Unlock()

	// clear all entry
	for i := 0; i < l.Cap; i++ {
		l.Logs[i].clearEntry()
	}
}
