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
		name                      string
		workerID                  int
		taskID                    string
		projectID                 string
		data                      []types.RealTimeDERData
		netOutput                 float64
		duration                  time.Duration
		expectedSuccess           bool
		expectedContractThreshold float64
	}{
		{
			name:      "successful outcome",
			workerID:  1,
			taskID:    "task1",
			projectID: "project1",
			data: []types.RealTimeDERData{
				{ID: "1", DER: types.DER{DerID: "der1", ContractThreshold: 0.75}},
			},
			netOutput:                 100.5,
			duration:                  500 * time.Millisecond,
			expectedSuccess:           true,
			expectedContractThreshold: 0.75,
		},
		{
			name:                      "nil data outcome",
			workerID:                  2,
			taskID:                    "task2",
			projectID:                 "project2",
			data:                      nil,
			netOutput:                 0,
			duration:                  100 * time.Millisecond,
			expectedSuccess:           false,
			expectedContractThreshold: 0,
		},
		{
			name:                      "empty data outcome",
			workerID:                  3,
			taskID:                    "task3",
			projectID:                 "project3",
			data:                      []types.RealTimeDERData{},
			netOutput:                 0,
			duration:                  100 * time.Millisecond,
			expectedSuccess:           false,
			expectedContractThreshold: 0,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			before := time.Now()
			outcome := New(tc.workerID, tc.taskID, tc.projectID, tc.data, tc.netOutput, tc.duration)
			after := time.Now()

			s.Equal(tc.expectedSuccess, outcome.Success)
			s.Equal(tc.workerID, outcome.WorkerID)
			s.Equal(tc.taskID, outcome.TaskID)
			s.Equal(tc.projectID, outcome.ProjectID)
			s.Equal(tc.data, outcome.Data)
			s.Equal(tc.netOutput, outcome.NetOutput)
			s.Equal(tc.duration.Milliseconds(), outcome.DurationMS)
			s.Equal(tc.expectedContractThreshold, outcome.ContractThreshold)
			s.True(outcome.CreatedAt.After(before) || outcome.CreatedAt.Equal(before))
			s.True(outcome.CreatedAt.Before(after) || outcome.CreatedAt.Equal(after))
		})
	}
}

func (s *OutcomeTestSuite) TestLogFields() {
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	outcome := &Outcome{
		Success:    true,
		WorkerID:   1,
		TaskID:     "task1",
		ProjectID:  "project1",
		NetOutput:  100.5,
		DurationMS: 500,
		CreatedAt:  testTime,
	}

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
	s.Equal("net_output", fields[10])
	s.Equal(100.5, fields[11])
	s.Equal("duration_ms", fields[12])
	s.Equal(int64(500), fields[13])
	s.Equal("created_at", fields[14])
	s.Equal("2024-01-01T12:00:00Z", fields[15])
}

func TestOutcomeSuite(t *testing.T) {
	suite.Run(t, new(OutcomeTestSuite))
}
