package buffer

import (
	"sync"
	"time"

	pb "github.com/grid-stream-org/grid-stream-protos/gen/validator/v1"

	"github.com/grid-stream-org/batcher/internal/types"
)

type AvgCache struct {
	mu        sync.Mutex
	items     map[string]*RunningAvg
	startTime time.Time
	endTime   time.Time
}

func NewAvgCache(startTime time.Time, endTime time.Time) *AvgCache {
	return &AvgCache{
		items:     make(map[string]*RunningAvg),
		startTime: startTime,
		endTime:   endTime,
	}
}

func (ac *AvgCache) Add(k string, v float64) bool {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ra, exists := ac.items[k]
	if !exists {
		ra = NewRunningAvg(k, ac.startTime, ac.endTime)
		ac.items[k] = ra
	}
	ra.Add(v)
	return exists
}

func (ac *AvgCache) GetOutputs() []types.AverageOutput {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	outputs := make([]types.AverageOutput, 0, len(ac.items))
	for _, ra := range ac.items {
		outputs = append(outputs, *ra.average)
	}
	return outputs
}

func (ac *AvgCache) GetProtoOutputs() []*pb.AverageOutput {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	outputs := make([]*pb.AverageOutput, 0, len(ac.items))
	for _, ra := range ac.items {
		avg := &pb.AverageOutput{
			ProjectId:     ra.average.ProjectID,
			AverageOutput: ra.average.AverageOutput,
			StartTime:     ra.average.StartTime.Format(time.RFC3339),
			EndTime:       ra.average.EndTime.Format(time.RFC3339),
		}
		outputs = append(outputs, avg)
	}
	return outputs
}

func (ac *AvgCache) Reset() {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	for k := range ac.items {
		delete(ac.items, k)
	}
}
