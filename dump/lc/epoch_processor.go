package lc

import (
	"log"
	"sync"
	"time"
)

type TimedQueue struct {
	commitQueue    []*CommitData
	ticker         *time.Ticker
	cur            int
	prev           int
	counter        int
	mutex          sync.Mutex
	epoch          int
	receiver       chan []*CommitData
	processedEpoch int
	cap            int
}

func NewTimedQueue(ms int, capacity int, p chan []*CommitData) *TimedQueue {

	t := &TimedQueue{
		commitQueue:    make([]*CommitData, capacity*3),
		ticker:         nil,
		cur:            0,
		prev:           0,
		counter:        0,
		epoch:          0,
		receiver:       p,
		processedEpoch: 0,
		cap:            capacity,
	}
	t.ticker = time.NewTicker(time.Duration(ms) * time.Millisecond)
	go func() {
		for {
			select {
			case e := <-t.ticker.C:
				log.Printf("Epoch #%v, time: %v", t.epoch, e)
				t.processEpoch()
				//log.Printf("commit data to process in this epoch: %v",)
			}
		}
	}()
	return t
}

func (q *TimedQueue) Insert(data *CommitData) {
	q.mutex.Lock()
	q.commitQueue[q.cur] = data
	q.cur = (q.cur + 1) % cap(q.commitQueue)
	q.counter += 1
	capacityReached := q.counter >= q.cap
	q.mutex.Unlock()
	if capacityReached {
		go q.processEpoch()
	}
}

func (q *TimedQueue) processEpoch() {
	q.mutex.Lock()
	epochQueue := make([]*CommitData, q.counter)
	for i := 0; i < q.counter; i++ {
		epochQueue[i] = q.commitQueue[(q.prev+i)%cap(q.commitQueue)]
		q.commitQueue[(q.prev+i)%cap(q.commitQueue)] = nil
	}
	q.prev = q.cur
	q.mutex.Unlock()
	q.receiver <- epochQueue
}
