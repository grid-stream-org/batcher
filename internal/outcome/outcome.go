package outcome

import (
	"time"

	"github.com/grid-stream-org/batcher/internal/types"
)

type Outcome struct {
	Success           bool                    `json:"success"`
	WorkerID          int                     `json:"worker_id"`
	TaskID            string                  `json:"task_id"`
	ProjectID         string                  `json:"project_id"`
	ContractThreshold float64                 `json:"contract_threshold"`
	NetOutput         float64                 `json:"net_output"`
	DurationMS        int64                   `json:"duration_ms"`
	CreatedAt         time.Time               `json:"created_at"`
	Data              []types.RealTimeDERData `json:"data"`
}

func New(workerID int, taskID string, projectID string, data []types.RealTimeDERData, netOutput float64, duration time.Duration) *Outcome {
	var contractThreshold float64
	if len(data) > 0 {
		contractThreshold = data[0].ContractThreshold
	}

	return &Outcome{
		Success:           len(data) > 0,
		WorkerID:          workerID,
		TaskID:            taskID,
		ContractThreshold: contractThreshold,
		ProjectID:         projectID,
		NetOutput:         netOutput,
		DurationMS:        duration.Milliseconds(),
		CreatedAt:         time.Now(),
		Data:              data,
	}
}

func (o *Outcome) LogFields() []any {
	fields := []any{
		"component", "outcome",
		"success", o.Success,
		"worker_id", o.WorkerID,
		"task_id", o.TaskID,
		"project_id", o.ProjectID,
		"net_output", o.NetOutput,
		"duration_ms", o.DurationMS,
		"created_at", o.CreatedAt.Format(time.RFC3339),
	}
	return fields
}
