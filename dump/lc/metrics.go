package lc

import (
	hm "github.com/cornelk/hashmap"
	"log"
	"time"
)

type Metrics struct {
	proposeDuration                  hm.HashMap
	batchFormationDuration           hm.HashMap
	batchCertificationDuration       hm.HashMap
	ProposeAddedChannel              chan int32
	ProposeCommittedChannel          chan int32
	BeginBatchCreateChannel          chan int32
	InitiateLogCertificationChannel  chan int32
	CompletedLogCertificationChannel chan int32
}

func (m *Metrics) LogProposeAdded() {
	for logIndex := range m.ProposeAddedChannel {
		//log.Printf("Propose creaed for log ID %v\n", logIndex)
		startTime := time.Now()
		m.proposeDuration.Insert(logIndex, startTime)
	}
}
func (m *Metrics) LogProposeCommitted() {

}
func (m *Metrics) LogBeginBatchCreation() {

}
func (m *Metrics) LogBeginBatchCompleted() {

}
func (m *Metrics) LogInitiateCertification() {

}

func (m *Metrics) LogCompletedCertification() {

}

func (m *Metrics) LogAllMetrics() {
	log.Printf("ReplicationDuration, Certification Duration, BatchFormationDuration")
	var minLen int
	minLen = funcName(m.batchCertificationDuration.Len(), m.proposeDuration.Len())
	minLen = funcName(minLen, m.batchFormationDuration.Len())

}

func funcName(m int, n int) (minLen int) {
	if m < n {
		minLen = m
	} else {
		minLen = n
	}
	return
}
