package outcome

import "time"

type TaskOutcome struct {
	Success        bool             `json:"success"`
	WorkerID       int              `json:"worker_id"`
	TaskID         string           `json:"task_id"`
	DataVal        map[string][]any `json:"data"`
	Project        string           `json:"project_id"`
	TotalDEROutput float64          `json:"total_der_output"`
	DurationMS     int64            `json:"duration_ms"`
	CreatedAt      time.Time        `json:"timestamp"`
}

/*
EVERYTHING IS BROKEN

- I basically over engineering the buffer, it should come back in internal/ and hold the running averages
calculated and updated in a goroutine. the destination should always buffer the data, so simplify.

- There should be a destination that also sends to validator.

- PutAll is useless, need to have a more custom query, so dont worry about the map[string]any

- I do like the custom flush func on the buffer, so just change the genericness of the buffer

1. Buffer is no longer generic or pkg/
2. Outcome gets buffered, prolly a two structs one for realtime data and one for averages.
3. On flush do the thing it needs to do for destination
4. Still not sure the best way to implement the destination, its honestly over engineered,
but I just like the factory

*/

func NewTaskOutcome(workerID int, taskID string, projectID string, data map[string][]any, totalDEROutput float64, duration time.Duration) Outcome {
	return &TaskOutcome{
		Success:        data != nil,
		WorkerID:       workerID,
		TaskID:         taskID,
		DataVal:        data,
		Project:        projectID,
		TotalDEROutput: totalDEROutput,
		DurationMS:     duration.Milliseconds(),
		CreatedAt:      time.Now(),
	}
}

func (o *TaskOutcome) ProjectID() string {
	return o.Project
}

func (o *TaskOutcome) CurrentOutput() float64 {
	return o.TotalDEROutput
}

func (o *TaskOutcome) Data() map[string][]any {
	return o.DataVal
}

func (o *TaskOutcome) LogFields() []any {
	fields := []any{
		"component", "outcome",
		"success", o.Success,
		"worker_id", o.WorkerID,
		"task_id", o.TaskID,
		"project_id", o.Project,
		"total_der_output", o.TotalDEROutput,
		"duration_ms", o.DurationMS,
		"created_at", o.CreatedAt,
	}
	return fields
}
