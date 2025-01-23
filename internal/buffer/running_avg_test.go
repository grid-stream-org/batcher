package buffer

import (
	"testing"
	"time"

	"github.com/grid-stream-org/batcher/internal/outcome"
	"github.com/grid-stream-org/batcher/internal/types"
	"github.com/stretchr/testify/suite"
)

type RunningAvgTestSuite struct {
	suite.Suite
	startTime time.Time
	endTime   time.Time
}

func (s *RunningAvgTestSuite) SetupTest() {
	s.startTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	s.endTime = time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC)
}

func (s *RunningAvgTestSuite) TestNewRunningAvg() {
	data := []types.RealTimeDERData{{
		ID: "der1",
		DER: types.DER{
			ProjectID:         "test-project",
			DerID:             "der1",
			ContractThreshold: 0.5,
			IsOnline:          true,
			CurrentOutput:     0.0,
		},
	}}
	o := outcome.New(1, "task1", "test-project", data, 0, time.Second)
	ra := NewRunningAvg(o, s.startTime, s.endTime)

	s.NotNil(ra)
	s.Equal(float64(0), ra.sum)
	s.Equal(int64(0), ra.count)
	s.NotNil(ra.average)
	s.Equal("test-project", ra.average.ProjectID)
	s.Equal(s.startTime, ra.average.StartTime)
	s.Equal(s.endTime, ra.average.EndTime)
	s.Equal(float64(0), ra.average.AverageOutput)
}

func (s *RunningAvgTestSuite) TestAdd() {
	testCases := []struct {
		name     string
		values   []float64
		expected float64
		expCount int64
		expSum   float64
	}{
		{
			name:     "single value",
			values:   []float64{10.0},
			expected: 10.0,
			expCount: 1,
			expSum:   10.0,
		},
		{
			name:     "multiple values",
			values:   []float64{10.0, 20.0, 30.0},
			expected: 20.0, // (10 + 20 + 30) / 3
			expCount: 3,
			expSum:   60.0,
		},
		{
			name:     "zero values",
			values:   []float64{0.0, 0.0},
			expected: 0.0,
			expCount: 2,
			expSum:   0.0,
		},
		{
			name:     "negative values",
			values:   []float64{-10.0, 10.0},
			expected: 0.0,
			expCount: 2,
			expSum:   0.0,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			data := []types.RealTimeDERData{{
				ID: "der1",
				DER: types.DER{
					ProjectID:         "test-project",
					DerID:             "der1",
					ContractThreshold: 0.5,
					IsOnline:          true,
					CurrentOutput:     0.0,
				},
			}}
			o := outcome.New(1, "task1", "test-project", data, 0, time.Second)
			ra := NewRunningAvg(o, s.startTime, s.endTime)

			for _, v := range tc.values {
				ra.Add(v)
			}
			s.Equal(tc.expCount, ra.count)
			s.Equal(tc.expSum, ra.sum)
			s.Equal(tc.expected, ra.average.AverageOutput)
		})
	}
}

func TestRunningAvgSuite(t *testing.T) {
	suite.Run(t, new(RunningAvgTestSuite))
}
