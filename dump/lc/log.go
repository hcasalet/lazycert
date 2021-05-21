package lc

import (
	"crypto/sha256"
	"github.com/golang/protobuf/proto"
	"log"
)

type Log struct {
	LogIndex              int32
	logEntry              map[int32]*LogEntry
	BatchedData           chan []*CommitData
	Certificate           chan *Certificate
	logEntryUpdateChannel chan *LogEntry
	dbDict                map[string][]byte
	config                *Config
}

func NewLog(cfg *Config) *Log {
	l := &Log{
		LogIndex:    -1,
		logEntry:    make(map[int32]*LogEntry),
		BatchedData: make(chan []*CommitData),
		Certificate: make(chan *Certificate),
		config:      cfg,
	}
	go l.processBatches()
	go l.certify()
	return l
}

func (l *Log) processBatches() {
	for batch := range l.BatchedData {

		l.LogIndex += 1
		currentLogIndex := l.LogIndex
		l.Propose(currentLogIndex, batch)
		/**
		This is a provisional update to the current data before the current log entry has been certified.
		*/
		go l.updateDBDict(currentLogIndex)
	}
}

func (l *Log) Propose(currentLogIndex int32, batch []*CommitData) (status bool) {
	log.Printf("Processing new batch for log index: %v", currentLogIndex)
	log.Printf("Batch contains %v transactions", len(batch))
	if _, ok := l.logEntry[currentLogIndex]; !ok {
		entry := &LogEntry{
			LogID: currentLogIndex,
			Data: &BlockInfo{
				LogID: currentLogIndex,
				Data:  batch,
			},
			TeCertificate: nil,
		}
		l.logEntry[currentLogIndex] = entry
		log.Printf("Log Entry at index %v: %v", currentLogIndex, entry)
		status = true
	} else {
		log.Printf("A log entry at %v already exists: %v", currentLogIndex, l.logEntry[currentLogIndex])
		status = false
	}
	return status
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

func (l *Log) SetLogEntryUpdateChannel(c chan *LogEntry) {
	l.logEntryUpdateChannel = c
}

func ConvertToProposeData(l LogEntry, n *NodeInfo) *ProposeData {
	p := &ProposeData{
		Header: &Header{
			Node: n,
		},
		LogBlock: l.Data,
	}
	return p
}

func ConvertToAcceptMsg(l *LogEntry, n *NodeInfo, term int32, k *Key) (a *AcceptMsg) {
	bytes, err := proto.Marshal(l.Data)
	acceptHash := sha256.Sum256(bytes)
	a = nil
	if err != nil {
		signature := k.SignMessage(bytes)
		a = &AcceptMsg{
			Header: &Header{
				Node: n,
			},
			AcceptHash: acceptHash[0:32],
			Signature:  signature,
			Block:      nil,
			TermID:     term,
		}
	}
	return a
}
