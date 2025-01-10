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

func NewAvgCache() *AvgCache {
	ac := &AvgCache{
		items: make(map[string]*RunningAvg),
		stop:  make(chan struct{}),
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

func (ac *AvgCache) Reset() {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	for _, avg := range ac.items {
		avg.Reset()
	}
}

func (ac *AvgCache) Close() {
	for k := range ac.items {
		ac.Delete(k)
	}

	if ac.interval > 0 {
		close(ac.stop)
	}
}
