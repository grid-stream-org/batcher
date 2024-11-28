package outcome

import "time"

type AvgOutcome struct {
	Success    bool             `json:"success"`
	DataVal    map[string][]any `json:"data"`
	Project    string           `json:"project_id"`
	AvgOutput  float64          `json:"avg_output"`
	DurationMS int64            `json:"duration_ms"`
	CreatedAt  time.Time        `json:"timestamp"`
}

func NewAvgOutcome(projectID string, data map[string][]any, avgOutput float64, duration time.Duration) Outcome {
	return &AvgOutcome{
		Success:    data != nil,
		DataVal:    data,
		Project:    projectID,
		AvgOutput:  avgOutput,
		DurationMS: duration.Milliseconds(),
		CreatedAt:  time.Now(),
	}
}

func (o *AvgOutcome) ProjectID() string {
	return o.Project
}

func (o *AvgOutcome) CurrentOutput() float64 {
	return o.AvgOutput
}

func (o *AvgOutcome) Data() map[string][]any {
	return o.DataVal
}

func (o *AvgOutcome) LogFields() []any {
	fields := []any{
		"component", "outcome",
		"success", o.Success,
		"project_id", o.Project,
		"avg_output", o.AvgOutput,
		"duration_ms", o.DurationMS,
		"created_at", o.CreatedAt,
	}
	return fields
}
