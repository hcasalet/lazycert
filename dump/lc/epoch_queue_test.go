package lc

import (
	"log"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestNewTimedQueue(t *testing.T) {
	f := make(chan []*CommitData)

	cap := 20
	n := 1000
	sum := 0
	sum = setup(cap, f, n, sum)
	if sum != n {
		t.Fatalf("Expected %v, Got %v", n, sum)
	}
	t.Logf("Final Sum: %v", sum)
}

func TestTimedQueueWithLessThanCapacity(t *testing.T) {
	f := make(chan []*CommitData)

	cap := 20
	n := 1
	sum := 0
	sum = setup(cap, f, n, sum)
	if sum != n {
		t.Fatalf("Expected %v, Got %v", n, sum)
	}
	t.Logf("Final Sum: %v", sum)
}

func setup(cap int, f chan []*CommitData, n int, sum int) int {
	q := NewTimedQueue(2, cap, f)
	go insertData(q, f, n)
	/*	data := <-f
		log.Printf("Data size := %v", len(data))
		if len(data) != 5 {
			t.Fatalf("Expected 5, Got %v", len(data))
		}*/

	for data := range f {
		log.Printf("Data size := %v", len(data))
		sum += len(data)
		//t.Logf("Sum: %v", sum)
	}
	return sum
}

func insertData(q *TimedQueue, f chan []*CommitData, n int) {
	for i := 0; i < n; i++ {
		q.Insert(generateRandomCommitData())
	}
	time.Sleep(100 * time.Millisecond)
	close(f)
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

/*func temp() {
	fifo := gq.NewFIFO()
	fifo.Enqueue(10)

	log.Println(fifo)
	log.Printf("queue capacity: %v, Queue Len: %v\n", fifo.GetCap(), fifo.GetLen())
	v, err := fifo.Dequeue()
	if err != nil {
		log.Printf("error when trying to get the value from queue: %v\n", err)
	} else {
		val := v.(int)

		log.Printf("Value: %v \n",val)
	}
}*/
