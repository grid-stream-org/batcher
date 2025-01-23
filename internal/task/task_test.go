package task

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/grid-stream-org/batcher/internal/outcome"
	"github.com/grid-stream-org/batcher/internal/types"
	"github.com/stretchr/testify/suite"
)

type TaskTestSuite struct {
	suite.Suite
	validDERs []types.DER
}

func (s *TaskTestSuite) SetupTest() {
	// Setup test data
	s.validDERs = []types.DER{
		{
			ProjectID:             "project1",
			DerID:                 "der1",
			CurrentOutput:         100.5,
			Units:                 "kW",
			IsOnline:              true,
			IsStandalone:          false,
			Timestamp:             types.NillableTime{Time: time.Now()},
			ConnectionStartAt:     types.NillableTime{Time: time.Now().Add(-1 * time.Hour)},
			CurrentSoc:            85.5,
			ContractThreshold:     120.0,
			Baseline:              120.0,
			PowerMeterMeasurement: 200,
		},
		{
			ProjectID:             "project1",
			DerID:                 "der2",
			CurrentOutput:         50.5,
			Units:                 "kW",
			IsOnline:              true,
			IsStandalone:          false,
			Timestamp:             types.NillableTime{Time: time.Now()},
			ConnectionStartAt:     types.NillableTime{Time: time.Now().Add(-2 * time.Hour)},
			CurrentSoc:            75.0,
			ContractThreshold:     80.0,
			Baseline:              120.0,
			PowerMeterMeasurement: 200,
		},
	}
}

func (s *TaskTestSuite) TestNewTask() {
	// Test creating new task with valid payload
	payload, err := json.Marshal(s.validDERs)
	s.NoError(err)

	task := NewTask(payload)

	s.NotEmpty(task.id)
	s.Equal(payload, task.payload)
	s.NotZero(task.createdAt)

	// Test idempotency of task ID generation
	task2 := NewTask(payload)
	s.Equal(task.id, task2.id)
}

func (s *TaskTestSuite) TestTaskExecute() {
	testCases := []struct {
		name        string
		payload     interface{}
		expectError error
		validate    func(*outcome.Outcome)
	}{
		{
			name:        "valid payload",
			payload:     s.validDERs,
			expectError: nil,
			validate: func(o *outcome.Outcome) {
				s.Equal("project1", o.ProjectID)
				s.Equal(49.0, o.NetOutput) // 100.5 + 50.5
				s.Len(o.Data, 2)
				s.NotEmpty(o.TaskID)
				s.Equal(1, o.WorkerID) // we'll use worker ID 1 for testing
			},
		},
		{
			name:        "empty DER array",
			payload:     []types.DER{},
			expectError: ErrNoDERs,
			validate:    nil,
		},
		{
			name:        "invalid JSON payload",
			payload:     make(chan int), // channels can't be marshaled to JSON
			expectError: nil,            // we expect a JSON marshaling error, but not a specific one
			validate:    nil,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			payload, err := json.Marshal(tc.payload)
			if err != nil && tc.expectError == nil {
				s.T().Skip("skipping test case due to expected JSON marshaling error")
				return
			}
			s.NoError(err)

			task := NewTask(payload)
			outcome, err := task.Execute(1) // using worker ID 1 for testing

			if tc.expectError != nil {
				s.ErrorIs(err, tc.expectError)
				s.Nil(outcome)
			} else if tc.validate != nil {
				s.NoError(err)
				s.NotNil(outcome)
				tc.validate(outcome)
			}
		})
	}
}

func (s *TaskTestSuite) TestLogFields() {
	payload, err := json.Marshal(s.validDERs)
	s.NoError(err)

	task := NewTask(payload)
	logFields := task.LogFields()

	s.Len(logFields, 6) // 3 key-value pairs
	s.Equal("component", logFields[0])
	s.Equal("task", logFields[1])
	s.Equal("id", logFields[2])
	s.Equal(task.id, logFields[3])
	s.Equal("created_at", logFields[4])
	s.NotEmpty(logFields[5]) // RFC3339 formatted time string
}

func (s *TaskTestSuite) TestMakeID() {
	// Test that makeID generates consistent IDs for the same payload
	payload := []byte("test payload")
	id1 := makeID(payload)
	id2 := makeID(payload)
	s.Equal(id1, id2)

	// Test that different payloads generate different IDs
	differentPayload := []byte("different payload")
	id3 := makeID(differentPayload)
	s.NotEqual(id1, id3)
}

func TestTaskSuite(t *testing.T) {
	suite.Run(t, new(TaskTestSuite))
}
