package lc

import "log"

type Log struct {
	LogIndex    int32
	logEntry    map[int32]*LogEntry
	BatchedData chan []*CommitData
	Certificate chan *Certificate
}

func NewLog() *Log {
	l := &Log{
		LogIndex:    -1,
		logEntry:    make(map[int32]*LogEntry),
		BatchedData: make(chan []*CommitData),
		Certificate: make(chan *Certificate),
	}
	go l.ProcessBatches()
	go l.Certify()
	return l
}

func (l *Log) ProcessBatches() {
	for {
		batch := <-l.BatchedData
		l.LogIndex += 1
		currentLogIndex := l.LogIndex
		log.Printf("Processing new batch for log index: %v", currentLogIndex)
		log.Printf("Batch contains %v transactions", len(batch))

	}
}

func (l *Log) Certify() {
	for {
		certificate := <-l.Certificate
		log.Printf("Certified for log position %v", certificate.LogID)
		if _, ok := l.logEntry[certificate.LogID]; ok {
			l.logEntry[certificate.LogID].TeCertificate = certificate
		}
	}
}
