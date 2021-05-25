package lc

import (
	"log"
	"math/rand"
	"strconv"
	"testing"
)

func TestNewTimedQueue(t *testing.T) {
	f := make(chan []*CommitData)
	q := NewTimedQueue(2, 5, f)
	go insertData(q)
	data := <-f
	log.Printf("Data size := %v", len(data))
	if len(data) != 5 {
		t.Fatalf("Expected 5, Got %v", len(data))
	}
}

func insertData(q *TimedQueue) {
	for i := 0; i < 10000; i++ {
		q.Insert(generateRandomCommitData())
	}
}

func generateRandomCommitData() *CommitData {
	d := make([]*KeyVal, 1)
	d[0] = &KeyVal{
		Key:   []byte(strconv.Itoa(rand.Int())),
		Value: []byte(strconv.Itoa(rand.Int())),
	}
	dat := &CommitData{
		Data: d,
	}
	return dat
}
