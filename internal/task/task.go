package task

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/google/uuid"
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

func (t *Task) Execute(workerId int) (*outcome.Outcome, error) {
	start := time.Now()
	var ders []outcome.DER

	if err := json.Unmarshal(t.payload, &ders); err != nil {
		return nil, errors.Wrap(err, "failed to parse message payload")
	}

	if len(ders) == 0 {
		return nil, ErrNoDERs
	}

	data := []outcome.RealTimeDERData{}
	var totalOutput float64 = 0
	var projID = ders[0].ProjectID
	for _, der := range ders {
		// Validation, we shouldnt ever get a payload of DERS from various project IDs
		if der.ProjectID != projID {
			return nil, ErrVariousProjectIDs
		}

		totalOutput += der.CurrentOutput
		derData := outcome.RealTimeDERData{
			ID:                uuid.New().String(),
			DerID:             der.DerID,
			DeviceID:          der.DeviceID,
			Timestamp:         der.Timestamp,
			CurrentOutput:     der.CurrentOutput,
			Units:             der.Units,
			ProjectID:         der.ProjectID,
			IsOnline:          der.IsOnline,
			IsStandalone:      der.IsStandalone,
			ConnectionStartAt: der.ConnectionStartAt,
			CurrentSoc:        der.CurrentSoc,
		}
		data = append(data, derData)
	}

	o := outcome.New(workerId, t.id, projID, data, totalOutput, time.Since(start))
	return &o, nil
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
