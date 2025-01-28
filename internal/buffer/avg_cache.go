package buffer

import (
	"sync"
	"time"

	pb "github.com/grid-stream-org/grid-stream-protos/gen/validator/v1"

	"github.com/grid-stream-org/batcher/internal/outcome"
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

func (ac *AvgCache) Add(o *outcome.Outcome) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ra, ok := ac.items[o.ProjectID]
	if !ok {
		ra = NewRunningAvg(o, ac.startTime, ac.endTime, o.Baseline)
		ac.items[o.ProjectID] = ra
	}
	ra.Add(o.NetOutput)
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
			ProjectId:         ra.average.ProjectID,
			AverageOutput:     ra.average.AverageOutput,
			ContractThreshold: ra.average.ContractThreshold,
			StartTime:         ra.average.StartTime.Format(time.RFC3339),
			EndTime:           ra.average.EndTime.Format(time.RFC3339),
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
