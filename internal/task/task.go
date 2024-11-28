package task

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/grid-stream-org/batcher/internal/outcome"
	"github.com/pkg/errors"
)

var (
	ErrNoDERs            = errors.New("received empty DER array")
	ErrVariousProjectIDs = errors.New("invalid payload; various project ids")
)

type Task struct {
	id        string
	payload   []byte
	createdAt time.Time
}

func NewTask(payload []byte) Task {
	return Task{
		id:        makeID(payload),
		payload:   payload,
		createdAt: time.Now(),
	}
}

func (t *Task) ID() string {
	return t.id
}

func (t *Task) Execute(workerId int) (outcome.Outcome, error) {
	start := time.Now()
	var ders []outcome.DER

	if err := json.Unmarshal(t.payload, &ders); err != nil {
		return nil, errors.Wrap(err, "failed to parse message payload")
	}

	if len(ders) == 0 {
		return nil, ErrNoDERs
	}

	projID, err := projectID(ders)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var totalDEROutput float64 = 0

	data := make(map[string][]any)

	o := outcome.NewTaskOutcome(workerId, t.id, projID, data, totalDEROutput, time.Since(start))
	return o, nil
}

func projectID(ders []outcome.DER) (string, error) {
	var projID = ders[0].ProjectID
	for _, der := range ders {
		// Validation, we shouldn ever get a payload of DERS from various project IDs
		if der.ProjectID != projID {
			return "", ErrVariousProjectIDs
		}
	}
	return projID, nil
}

func makeID(payload []byte) string {
	hash := sha256.Sum256(payload)
	return hex.EncodeToString(hash[:])
}

func (t *Task) LogFields() []any {
	return []any{
		"component", "task",
		"id", t.id,
		"created_at", t.createdAt,
	}
}
