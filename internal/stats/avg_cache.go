package stats

import (
	"sync"
	"time"
)

type AvgCache struct {
	items    map[string]*RunningAvg
	mu       sync.RWMutex
	interval time.Duration
	stop     chan struct{}
}

func NewAvgCache(cleanupInterval time.Duration) *AvgCache {
	ac := &AvgCache{
		items:    make(map[string]*RunningAvg),
		interval: cleanupInterval,
		stop:     make(chan struct{}),
	}

	if cleanupInterval > 0 {
		go ac.cleanupLoop()
	}

	return ac
}

func (ac *AvgCache) Add(k string, v float64) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	avg, ok := ac.items[k]
	if !ok {
		avg = NewRunningAvg()
		ac.items[k] = avg
	}
	avg.Add(v)
}

func (ac *AvgCache) Get(k string) (*RunningAvg, bool) {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	if avg, ok := ac.items[k]; ok {
		return avg, true
	}
	return nil, false
}

func (ac *AvgCache) GetAll() map[string]*RunningAvg {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.items
}

func (ac *AvgCache) Size() int {
	return len(ac.items)
}

func (ac *AvgCache) Delete(k string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	delete(ac.items, k)
}

func (ac *AvgCache) Reset(k string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if avg, ok := ac.items[k]; ok {
		avg.Reset()
	}
}

func (ac *AvgCache) cleanup() {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if ac.interval <= 0 {
		return
	}

	for k := range ac.items {
		delete(ac.items, k)
	}
}

func (ac *AvgCache) cleanupLoop() {
	ticker := time.NewTicker(ac.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ac.cleanup()
		case <-ac.stop:
			return
		}
	}
}

func (ac *AvgCache) Close() {
	if ac.interval > 0 {
		close(ac.stop)
	}
}
