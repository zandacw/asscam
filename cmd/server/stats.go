package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/langlandsbrogram/asscam/pkg/terminal"
)

type Stats struct {
	totalBytes    int
	start         time.Time
	intervalBytes int
	intervalTime  time.Time
	interval      uint8
	mu            sync.Mutex
}

func NewStats(interval uint8) *Stats {
	return &Stats{
		start:        time.Now(),
		intervalTime: time.Now(),
		interval:     interval,
	}
}

func (s *Stats) ProcessBytes(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.totalBytes += n
	s.intervalBytes += n
	if since := time.Since(s.intervalTime).Seconds(); since > float64(s.interval) {
		terminal.ClearScreen()
		fmt.Printf(
			"%.2f KB/s\n",
			float64(s.intervalBytes)/float64(since)/1000.0,
		)
		s.intervalBytes = 0
		s.intervalTime = time.Now()
	}

}

func (s *Stats) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.intervalBytes = 0
	s.intervalTime = time.Now()
}

// make sure that is no data is coming through the
// stats info resets to 0 bytes
func (s *Stats) Check() {
	for {
		<-time.After(1 * time.Second)
		s.ProcessBytes(0)

	}
}
