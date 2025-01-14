package buffer

import (
	"time"

	"github.com/grid-stream-org/batcher/internal/types"
)

type RunningAvg struct {
	sum     float64
	count   int64
	average *types.AverageOutput
}

func NewRunningAvg(projectID string, startTime time.Time, endTime time.Time) *RunningAvg {
	return &RunningAvg{
		sum:   0,
		count: 0,
		average: &types.AverageOutput{
			ProjectID: projectID,
			StartTime: startTime,
			EndTime:   endTime,
		},
	}
}

func (ra *RunningAvg) Add(v float64) {
	ra.sum += v
	ra.count++
	ra.average.AverageOutput = ra.sum / float64(ra.count)
}
