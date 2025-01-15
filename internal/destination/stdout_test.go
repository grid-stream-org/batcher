package destination

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/grid-stream-org/batcher/internal/buffer"
	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/internal/outcome"
	"github.com/grid-stream-org/batcher/internal/types"
	"github.com/stretchr/testify/suite"
)

type StdoutDestinationTestSuite struct {
	suite.Suite
	ctx       context.Context
	dest      *stdoutDestination
	mockVC    *MockValidatorClient
	logBuffer *bytes.Buffer
	log       *slog.Logger
}

func (s *StdoutDestinationTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.mockVC = new(MockValidatorClient)
	s.logBuffer = new(bytes.Buffer)
	s.log = slog.New(slog.NewJSONHandler(s.logBuffer, nil))

	cfg := &config.Destination{
		Type: "stdout",
		Buffer: &config.Buffer{
			StartTime: time.Now(),
			Interval:  time.Minute,
			Offset:    time.Second * 30,
		},
	}

	dest, err := newStdoutDestination(s.ctx, cfg, s.mockVC, s.log)
	s.NoError(err)
	s.dest = dest.(*stdoutDestination)
}

func (s *StdoutDestinationTestSuite) TestAdd() {
	testCases := []struct {
		name        string
		input       interface{}
		setupMock   func()
		expectError bool
	}{
		{
			name: "valid outcome",
			input: &outcome.Outcome{
				Success:   true,
				WorkerID:  1,
				TaskID:    "task1",
				ProjectID: "project1",
				Data:      []types.RealTimeDERData{},
				CreatedAt: time.Now(),
			},
			setupMock: func() {
				// Setup mock for NotifyProject since this is a new project
				s.mockVC.On("NotifyProject", s.ctx, "project1").Return(nil).Once()
			},
			expectError: false,
		},
		{
			name: "valid outcome - existing project",
			input: &outcome.Outcome{
				Success:   true,
				WorkerID:  1,
				TaskID:    "task1",
				ProjectID: "project1",
				Data:      []types.RealTimeDERData{},
				CreatedAt: time.Now(),
			},
			setupMock: func() {
				// No NotifyProject call needed for existing project
			},
			expectError: false,
		},
		{
			name:        "invalid type",
			input:       "not an outcome",
			setupMock:   func() {},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.setupMock()
			err := s.dest.Add(s.ctx, tc.input)
			if tc.expectError {
				s.Error(err)
				s.Contains(err.Error(), "expected *outcome.Outcome")
			} else {
				s.NoError(err)
			}
			s.mockVC.AssertExpectations(s.T())
		})
	}
}

func (s *StdoutDestinationTestSuite) TestClose() {
	err := s.dest.Close()
	s.NoError(err)
	s.Contains(s.logBuffer.String(), "stdout destination closed")
}

func (s *StdoutDestinationTestSuite) TestFlushFunc() {
	testData := &buffer.FlushOutcome{
		Outcomes: []outcome.Outcome{
			{
				Success:   true,
				WorkerID:  1,
				TaskID:    "task1",
				ProjectID: "project1",
				Data:      []types.RealTimeDERData{},
			},
		},
		AvgOutputs: []types.AverageOutput{
			{
				ProjectID:     "project1",
				AverageOutput: 100.0,
				StartTime:     time.Now(),
				EndTime:       time.Now().Add(time.Hour),
			},
		},
	}

	err := s.dest.flushFunc(s.ctx, testData)
	s.NoError(err)
}

func TestStdoutDestinationSuite(t *testing.T) {
	suite.Run(t, new(StdoutDestinationTestSuite))
}
