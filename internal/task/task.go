package task

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/grid-stream-org/batcher/internal/outcome"
	"github.com/grid-stream-org/batcher/internal/types"
	"github.com/pkg/errors"
)

var (
	ErrNoDERs = errors.New("received empty DER array")
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

func (t *Task) Execute(workerId int) (*outcome.Outcome, error) {
	start := time.Now()
	var ders []types.DER

	if err := json.Unmarshal(t.payload, &ders); err != nil {
		return nil, errors.Wrap(err, "failed to parse message payload")
	}

	if len(ders) == 0 {
		return nil, ErrNoDERs
	}

	data := []types.RealTimeDERData{}
	var netOutput float64 = ders[0].PowerMeterMeasurement
	for _, der := range ders {
		netOutput -= der.CurrentOutput
		derData := types.RealTimeDERData{
			ID:  uuid.New().String(),
			DER: der,
		}
		data = append(data, derData)
	}

	o := outcome.New(workerId, t.id, ders[0].ProjectID, data, netOutput, time.Since(start))
	return o, nil
}

func makeID(payload []byte) string {
	hash := sha256.Sum256(payload)
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func (t *Task) LogFields() []any {
	return []any{
		"component", "task",
		"id", t.id,
		"created_at", t.createdAt.Format(time.RFC3339),
	}
}
