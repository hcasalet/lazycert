package lc

import "log"

type Log struct {
	LogIndex    int32
	logEntry    map[int32]*LogEntry
	BatchedData chan []*CommitData
	Certificate chan *Certificate
	dbDict      map[string][]byte
}

func NewLog() *Log {
	l := &Log{
		LogIndex:    -1,
		logEntry:    make(map[int32]*LogEntry),
		BatchedData: make(chan []*CommitData),
		Certificate: make(chan *Certificate),
	}
	go l.processBatches()
	go l.certify()
	return l
}

func (l *Log) processBatches() {
	for batch := range l.BatchedData {

		l.LogIndex += 1
		currentLogIndex := l.LogIndex
		log.Printf("Processing new batch for log index: %v", currentLogIndex)
		log.Printf("Batch contains %v transactions", len(batch))
		entry := &LogEntry{
			LogID: currentLogIndex,
			Data: &BlockInfo{
				LogID: currentLogIndex,
				Data:  batch,
			},
			TeCertificate: nil,
		}
		log.Printf("Log Entry at index %v: %v", currentLogIndex, entry)
		/**
		This is a provisional update to the current data before the current log entry has been certified.
		*/
		go l.updateDBDict(currentLogIndex)
	}
}

func (l *Log) certify() {
	for certificate := range l.Certificate {
		logIndex := certificate.LogID
		log.Printf("Certified for log position %v", logIndex)
		if _, ok := l.logEntry[logIndex]; ok {
			if l.logEntry[logIndex].TeCertificate == nil {
				l.logEntry[logIndex].TeCertificate = certificate
				/**
				Update DB dictionary after certification of the data from TE.
				*/
				go l.updateDBDict(logIndex)
			} else {
				log.Printf("Possible duplicate certificate received for index %v, %v", logIndex, certificate)
			}
		}
	}
}

func (l *Log) updateDBDict(logIndex int32) {
	entry := l.logEntry[logIndex].Data.Data
	for _, data := range entry {
		for _, kv := range data.Data {
			l.dbDict[string(kv.Key)] = kv.Value
		}
	}
}

func (l *Log) Read(key string) ([]byte, bool) {
	status := false
	if _, ok := l.dbDict[key]; ok {
		status = ok
	}
	return l.dbDict[key], status
}
