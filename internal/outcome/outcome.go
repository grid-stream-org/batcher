package outcome

import "time"

type Outcome struct {
	Success   bool          `json:"success"`
	WorkerID  int           `json:"workerId"`
	TableName string        `json:"table"`
	DataValue any           `json:"data"`
	Duration  time.Duration `json:"duration"`
	CreatedAt time.Time     `json:"timestamp"`
}

func New(workerId int, data any, table string, duration time.Duration) *Outcome {
	return &Outcome{
		Success:   data != nil,
		WorkerID:  workerId,
		TableName: table,
		DataValue: data,
		Duration:  time.Duration(duration.Milliseconds()),
		CreatedAt: time.Now(),
	}
}

func (o *Outcome) Table() string {
	return o.TableName
}

func (o *Outcome) Data() any {
	return o.DataValue
}

func (o *Outcome) LogFields() []any {
	fields := []any{
		"component", "outcome",
		"worker_id", o.WorkerID,
		"table", o.Table,
		"success", o.Success,
		"duration_ms", o.Duration.Milliseconds(),
		"created_at", o.CreatedAt,
	}
	return fields
}
