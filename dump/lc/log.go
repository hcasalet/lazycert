package lc

import (
	hm "github.com/cornelk/hashmap"
	"github.com/golang/protobuf/proto"
	"log"
	"time"
)

type Log struct {
	LogIndex int32
	logEntry hm.HashMap
	//logEntry              map[int32]*LogEntry
	BatchedData           chan []*CommitData
	Certificate           chan *Certificate
	logEntryUpdateChannel chan *LogEntry
	dbDict                hm.HashMap
	//dbDict                map[string][]byte
	config             *Config
	logProcessingTimes hm.HashMap
}

func NewLog(cfg *Config) *Log {
	l := &Log{
		LogIndex: -1,
		logEntry: hm.HashMap{},
		//logEntry:              make(map[int32]*LogEntry),
		BatchedData: make(chan []*CommitData),
		Certificate: make(chan *Certificate),
		dbDict:      hm.HashMap{},
		//dbDict:                make(map[string][]byte),
		config:                cfg,
		logEntryUpdateChannel: nil,
		logProcessingTimes:    hm.HashMap{},
	}
	go l.processBatches()
	go l.certify()
	return l
}

func (l *Log) processBatches() {
	for batch := range l.BatchedData {
		//log.Printf("received txn batch: %v\n", batch)
		batchSize := len(batch)
		batchProcessingStartTime := time.Now()
		l.LogIndex += 1
		log.Printf("received txn batch for epoch: %v, batchsize: %v\n", l.LogIndex, len(batch))
		currentLogIndex := l.LogIndex
		l.Propose(currentLogIndex, batch)
		if l.logEntryUpdateChannel != nil {
			//l.logEntryUpdateChannel <- l.logEntry[currentLogIndex]
			if v, ok := l.logEntry.Get(currentLogIndex); ok {
				tv := v.(*LogEntry) // Type Assertion
				l.logEntryUpdateChannel <- tv

			}
		}
		/**
		This is a provisional update to the current data before the current log entry has been certified.
		*/
		l.updateDBDict(currentLogIndex)
		log.Printf("LOG COMMIT TIME, LOGID, batch size: %s, %v, %v", time.Since(batchProcessingStartTime), currentLogIndex, batchSize)
	}
}

func (l *Log) Propose(currentLogIndex int32, batch []*CommitData) (status bool) {
	//log.Printf("Processing new batch for log index: %v", currentLogIndex)
	//log.Printf("Batch contains %v transactions", len(batch))
	//if _, ok := l.logEntry[currentLogIndex]; !ok {
	if _, ok := l.logEntry.Get(currentLogIndex); !ok {
		entry := &LogEntry{
			LogID: currentLogIndex,
			Data: &BlockInfo{
				LogID: currentLogIndex,
				Data:  batch,
			},
			TeCertificate: nil,
		}
		l.logEntry.Insert(currentLogIndex, entry)
		//l.logEntry[currentLogIndex] = entry
		//log.Printf("Log Entry at index %v: %v", currentLogIndex, entry)
		status = true
	} else {
		v, ok := l.logEntry.Get(currentLogIndex)
		log.Printf("A log entry at %v already exists: %v, %v", currentLogIndex, v, ok)
		//log.Printf("A log entry at %v already exists: %v", currentLogIndex, l.logEntry[currentLogIndex])
		status = false
	}
	return status
}

func (l *Log) certify() {
	for certificate := range l.Certificate {
		logIndex := certificate.LogID
		log.Printf("Received certificate for log position %v", logIndex)
		//if _, ok := l.logEntry[logIndex]; ok {
		if e, ok := l.logEntry.Get(logIndex); ok {
			entry := e.(*LogEntry)

			if entry.TeCertificate == nil {
				entry.TeCertificate = certificate
				/**
				Update DB dictionary after certification of the data from TE.
				*/
				// TODO Check if this should be done. Should we overwrite previously held data?
				//l.updateDBDict(logIndex)
			} else {
				log.Printf("Possible duplicate certificate received for index %v, %v", logIndex, certificate)
			}
		}
	}
}

func (l *Log) updateDBDict(logIndex int32) {
	//entry := l.logEntry[logIndex].Data.Data
	e, _ := l.logEntry.Get(logIndex)
	entry := e.(*LogEntry).Data.Data
	for _, data := range entry {
		for _, kv := range data.Data {
			l.dbDict.Set(string(kv.Key), kv.Value)
		}
	}
}

func (l *Log) Read(key string) ([]byte, bool) {
	d, ok := l.dbDict.Get(key)
	var data []byte
	if ok {
		data = d.([]byte)
	} else {
		data = nil
	}
	return data, ok
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
	//log.Printf("Marshalling blockinfo: %v\n", l.Data)
	messageBytes, err := proto.Marshal(l.Data)
	a = nil
	if err == nil {
		hashedMessage, signature := k.SignMessage(messageBytes)
		a = &AcceptMsg{
			Header: &Header{
				Node: n,
			},
			AcceptHash: hashedMessage[:],
			Signature:  signature,
			Block: &BlockInfo{
				LogID: l.LogID,
				Data:  nil,
			},
			TermID: term,
		}
	}
	//log.Printf("lodID:HASH = %v, %v", l.LogID, a. )
	return a
}
