package lc

import (
	gq "github.com/enriquebris/goconcurrentqueue"
	"log"
	"sync"
	"time"
)

type TimedQueue struct {
	//commitQueue    []*CommitData
	ticker *time.Ticker
	cur    int
	prev   int
	//counter        int
	mutex          sync.Mutex
	epoch          int
	receiver       chan []*CommitData
	processedEpoch int
	cap            int
	d              time.Duration
	processed      map[int]bool
	queue          *gq.FIFO
	batchStartTime time.Time
}

func NewTimedQueue(ms int, capacity int, p chan []*CommitData) *TimedQueue {

	t := &TimedQueue{
		//commitQueue:    make([]*CommitData, capacity*1000),
		queue:  gq.NewFIFO(),
		ticker: nil,
		cur:    0,
		prev:   0,
		//counter:        0,
		epoch:          0,
		receiver:       p,
		processedEpoch: 0,
		cap:            capacity,
		processed:      make(map[int]bool),
	}
	epochDuration := time.Duration(ms) * time.Millisecond
	t.d = epochDuration
	t.ticker = time.NewTicker(epochDuration)
	go func() {
		for {
			select {
			case <-t.ticker.C:
				//log.Printf("Ticker: %v", e)
				t.mutex.Lock()
				t.processEpoch(t.epoch, t.cap)
				t.mutex.Unlock()
				//log.Printf("commit data to process in this epoch: %v",)
			}
		}
	}()
	t.batchStartTime = time.Now()
	return t
}

func (q *TimedQueue) Insert(data *CommitData) {
	//q.commitQueue[q.cur] = data
	q.queue.Enqueue(data)
	//q.cur = (q.cur + 1) % cap(q.commitQueue)
	//q.counter += 1
	capacityReached := q.queue.GetLen() >= q.cap

	if capacityReached {
		//q.mutex.Unlock()
		q.mutex.Lock()
		e := q.epoch
		//q.counter = 0
		q.processEpoch(e, q.cap)
		q.mutex.Unlock()
	}
}

func (q *TimedQueue) processEpoch(n int, count int) {
	//q.mutex.Lock()
	q.ticker.Stop()
	if _, ok := q.processed[n]; !ok && count > 0 && q.queue.GetLen() > 0 {
		if q.queue.GetLen() < count {
			count = q.queue.GetLen()
		}
		log.Printf("Processing epoch %v", n)
		q.processed[n] = true
		epochQueue := make([]*CommitData, count)
		for i := 0; i < count; i++ {
			v, _ := q.queue.Dequeue()
			epochQueue[i] = v.(*CommitData)
		}
		q.prev = q.cur
		q.receiver <- epochQueue
		q.epoch += 1
		//q.counter = 0
		log.Printf("BATCH FORMATION TIME, SIZE: %s, %v", time.Since(q.batchStartTime), count)
		q.batchStartTime = time.Now()
	}
	q.ticker.Reset(q.d)

}
