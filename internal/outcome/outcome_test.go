package outcome

import (
	"testing"
	"time"

	"github.com/grid-stream-org/batcher/internal/types"
	"github.com/stretchr/testify/suite"
)

type OutcomeTestSuite struct {
	suite.Suite
}

func (s *OutcomeTestSuite) TestNew() {
	testCases := []struct {
		name        string
		workerID    int
		taskID      string
		projectID   string
		data        []types.RealTimeDERData
		totalOutput float64
		duration    time.Duration
	}{
		{
			name:      "successful outcome",
			workerID:  1,
			taskID:    "task1",
			projectID: "project1",
			data: []types.RealTimeDERData{
				{ID: "1", DER: types.DER{DerID: "der1"}},
			},
			totalOutput: 100.5,
			duration:    500 * time.Millisecond,
		},
		{
			name:        "nil data outcome",
			workerID:    2,
			taskID:      "task2",
			projectID:   "project2",
			data:        nil,
			totalOutput: 0,
			duration:    100 * time.Millisecond,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			before := time.Now()
			outcome := New(tc.workerID, tc.taskID, tc.projectID, tc.data, tc.totalOutput, tc.duration)
			after := time.Now()

			s.Equal(tc.data != nil, outcome.Success)
			s.Equal(tc.workerID, outcome.WorkerID)
			s.Equal(tc.taskID, outcome.TaskID)
			s.Equal(tc.projectID, outcome.ProjectID)
			s.Equal(tc.data, outcome.Data)
			s.Equal(tc.totalOutput, outcome.NetOutput)
			s.Equal(tc.duration.Milliseconds(), outcome.DurationMS)

			// CreatedAt should be between before and after
			s.True(outcome.CreatedAt.After(before) || outcome.CreatedAt.Equal(before))
			s.True(outcome.CreatedAt.Before(after) || outcome.CreatedAt.Equal(after))
		})
	}
}

func (s *OutcomeTestSuite) TestLogFields() {
	outcome := New(
		1,
		"task1",
		"project1",
		[]types.RealTimeDERData{{ID: "1"}},
		100.5,
		500*time.Millisecond,
	)

	fields := outcome.LogFields()

	s.Len(fields, 16) // 8 key-value pairs
	s.Equal("component", fields[0])
	s.Equal("outcome", fields[1])
	s.Equal("success", fields[2])
	s.Equal(true, fields[3])
	s.Equal("worker_id", fields[4])
	s.Equal(1, fields[5])
	s.Equal("task_id", fields[6])
	s.Equal("task1", fields[7])
	s.Equal("project_id", fields[8])
	s.Equal("project1", fields[9])
	s.Equal("total_output", fields[10])
	s.Equal(100.5, fields[11])
	s.Equal("duration_ms", fields[12])
	s.Equal(int64(500), fields[13])
	s.Equal("created_at", fields[14])
}

func TestOutcomeSuite(t *testing.T) {
	suite.Run(t, new(OutcomeTestSuite))
}
