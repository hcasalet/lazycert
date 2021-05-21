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
	d              time.Duration
	processed      map[int]bool
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
		processed:      make(map[int]bool),
	}
	epochDuration := time.Duration(ms) * time.Millisecond
	t.d = epochDuration
	t.ticker = time.NewTicker(epochDuration)
	go func() {
		for {
			select {
			case e := <-t.ticker.C:
				log.Printf("Ticker: %v", e)
				t.processEpoch(t.epoch)
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
	e := q.epoch
	q.mutex.Unlock()
	if capacityReached {
		go q.processEpoch(e)
	}
}

func (q *TimedQueue) processEpoch(n int) {
	q.mutex.Lock()
	if _, ok := q.processed[n]; !ok {
		log.Printf("Processing epoch %v", n)
		q.processed[n] = true
		q.ticker.Reset(q.d)
		epochQueue := make([]*CommitData, q.counter)
		for i := 0; i < q.counter; i++ {
			epochQueue[i] = q.commitQueue[(q.prev+i)%cap(q.commitQueue)]
			q.commitQueue[(q.prev+i)%cap(q.commitQueue)] = nil
		}
		q.prev = q.cur
		q.receiver <- epochQueue
		q.epoch += 1
		q.counter = 0
	} else {
		log.Printf("Epoch already processed. %v", n)
	}
	q.mutex.Unlock()
}
