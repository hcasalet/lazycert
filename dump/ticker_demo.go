package main

import (
	"log"
	"time"
)

type MyQueue struct {
	ticker        *time.Ticker
	cap           int
	count         int
	duration_unit time.Duration
	duration      int
}

func (m *MyQueue) runner() {
	go func() {
		for {
			m.count += 1
			time.Sleep(1 * m.duration_unit)
			if m.count == m.cap {
				log.Println("Capacity reached: Calling processQueue")
				m.processQueue()
			}
		}
	}()
}

func (m *MyQueue) processQueue() {
	m.ticker.Stop()
	log.Printf("Processing the queue, count=%v", m.count)
	m.count = 0
	m.ticker.Reset(time.Duration(m.duration) * m.duration_unit)
}
func main() {
	m := MyQueue{
		cap:           15,
		count:         0,
		duration_unit: time.Millisecond,
		duration:      25,
	}
	m.ticker = time.NewTicker(m.duration_unit * time.Duration(m.duration))

	go m.runner()
	go func() {
		for {
			select {
			case <-m.ticker.C:
				log.Println("Ticker ticked!")
				m.processQueue()
			}
		}
	}()
	done := make(chan bool, 1)
	<-done
}
