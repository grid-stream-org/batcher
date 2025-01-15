package outcome

import (
	"time"

	"github.com/grid-stream-org/batcher/internal/types"
)

type Outcome struct {
	Success     bool                    `json:"success"`
	WorkerID    int                     `json:"worker_id"`
	TaskID      string                  `json:"task_id"`
	Data        []types.RealTimeDERData `json:"data"`
	ProjectID   string                  `json:"project_id"`
	TotalOutput float64                 `json:"total_der_output"`
	DurationMS  int64                   `json:"duration_ms"`
	CreatedAt   time.Time               `json:"created_at"`
}

func New(workerID int, taskID string, projectID string, data []types.RealTimeDERData, totalOutput float64, duration time.Duration) Outcome {
	return Outcome{
		Success:     data != nil,
		WorkerID:    workerID,
		TaskID:      taskID,
		Data:        data,
		ProjectID:   projectID,
		TotalOutput: totalOutput,
		DurationMS:  duration.Milliseconds(),
		CreatedAt:   time.Now(),
	}
}

func (o *Outcome) LogFields() []any {
	fields := []any{
		"component", "outcome",
		"success", o.Success,
		"worker_id", o.WorkerID,
		"task_id", o.TaskID,
		"project_id", o.ProjectID,
		"total_output", o.TotalOutput,
		"duration_ms", o.DurationMS,
		"created_at", o.CreatedAt.Format(time.RFC3339),
	}
	return fields
}
