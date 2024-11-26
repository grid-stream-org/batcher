package task

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/grid-stream-org/batcher/internal/outcome"
	"github.com/grid-stream-org/batcher/pkg/models"
	"github.com/pkg/errors"
)

var (
	ErrNoDERs = errors.New("received empty DER array")
)

type Task struct {
	ID        string
	payload   []byte
	CreatedAt time.Time
}

func NewTask(payload []byte) Task {
	return Task{
		ID:        makeID(payload),
		payload:   payload,
		CreatedAt: time.Now(),
	}
}

func (t *Task) Execute(workerId int) (*outcome.Outcome, error) {
	start := time.Now()
	var ders []models.DER

	if err := json.Unmarshal(t.payload, &ders); err != nil {
		return nil, errors.Wrap(err, "failed to parse message payload")
	}

	if len(ders) == 0 {
		return nil, ErrNoDERs
	}

	finish := time.Now()
	duration := finish.Sub(start)
	o := outcome.New(workerId, ders, "test-table", duration)
	return o, nil
}

func makeID(payload []byte) string {
	hash := sha256.Sum256(payload)
	return hex.EncodeToString(hash[:])
}

func (t *Task) LogFields() []any {
	return []any{
		"component", "task",
		"id", t.ID,
		"created_at", t.CreatedAt,
	}
}
