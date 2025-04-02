package buffer

import (
	"testing"
	"time"

	"github.com/grid-stream-org/batcher/internal/outcome"
	"github.com/grid-stream-org/batcher/internal/types"
	pb "github.com/grid-stream-org/grid-stream-protos/gen/validator/v1"
	"github.com/stretchr/testify/suite"
)

type AvgCacheTestSuite struct {
	suite.Suite
	startTime time.Time
	endTime   time.Time
	cache     *AvgCache
}

func (s *AvgCacheTestSuite) SetupTest() {
	s.startTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	s.endTime = time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC)
	s.cache = NewAvgCache(s.startTime, s.endTime)
}

func (s *AvgCacheTestSuite) TestNewAvgCache() {
	s.NotNil(s.cache)
	s.NotNil(s.cache.items)
	s.Equal(s.startTime, s.cache.startTime)
	s.Equal(s.endTime, s.cache.endTime)
	s.Empty(s.cache.items)
}

func (s *AvgCacheTestSuite) TestAdd() {
	// First addition (new item)
	data1 := []types.RealTimeDERData{{
		ID: "der1",
		DER: types.DER{
			ProjectID:         "project1",
			DerID:             "der1",
			ContractThreshold: 0.5,
			IsOnline:          true,
			CurrentOutput:     10.0,
		},
	}}
	o1 := outcome.New(1, "task1", "project1", data1, 10.0, time.Second)
	s.cache.Add(o1)
	s.Len(s.cache.items, 1)

	// Second addition (existing item)
	data2 := []types.RealTimeDERData{{
		ID: "der1",
		DER: types.DER{
			ProjectID:         "project1",
			DerID:             "der1",
			ContractThreshold: 0.5,
			IsOnline:          true,
			CurrentOutput:     20.0,
		},
	}}
	o2 := outcome.New(1, "task1", "project1", data2, 20.0, time.Second)
	s.cache.Add(o2)
	s.Len(s.cache.items, 1)

	// Different project
	data3 := []types.RealTimeDERData{{
		ID: "der2",
		DER: types.DER{
			ProjectID:         "project2",
			DerID:             "der2",
			ContractThreshold: 0.75,
			IsOnline:          true,
			CurrentOutput:     30.0,
		},
	}}
	o3 := outcome.New(1, "task1", "project2", data3, 30.0, time.Second)
	s.cache.Add(o3)
	s.Len(s.cache.items, 2)

	// Verify the averages
	s.Equal(15.0, s.cache.items["project1"].average.AverageOutput) // (10 + 20) / 2
	s.Equal(30.0, s.cache.items["project2"].average.AverageOutput)
}

func (s *AvgCacheTestSuite) TestGetOutputs() {
	// Add some test data
	data1 := []types.RealTimeDERData{{
		ID: "der1",
		DER: types.DER{
			ProjectID:         "project1",
			DerID:             "der1",
			ContractThreshold: 0.5,
			IsOnline:          true,
			CurrentOutput:     10.0,
		},
	}}
	o1 := outcome.New(1, "task1", "project1", data1, 10.0, time.Second)
	s.cache.Add(o1)

	data2 := []types.RealTimeDERData{{
		ID: "der1",
		DER: types.DER{
			ProjectID:         "project1",
			DerID:             "der1",
			ContractThreshold: 0.5,
			IsOnline:          true,
			CurrentOutput:     20.0,
		},
	}}
	o2 := outcome.New(1, "task1", "project1", data2, 20.0, time.Second)
	s.cache.Add(o2)

	data3 := []types.RealTimeDERData{{
		ID: "der2",
		DER: types.DER{
			ProjectID:         "project2",
			DerID:             "der2",
			ContractThreshold: 0.75,
			IsOnline:          true,
			CurrentOutput:     30.0,
		},
	}}
	o3 := outcome.New(1, "task1", "project2", data3, 30.0, time.Second)
	s.cache.Add(o3)

	outputs := s.cache.GetOutputs()
	s.Len(outputs, 2)

	// Sort outputs by project ID for consistent testing
	var proj1, proj2 types.AverageOutput
	for _, out := range outputs {
		if out.ProjectID == "project1" {
			proj1 = out
		} else {
			proj2 = out
		}
	}

	s.Equal("project1", proj1.ProjectID)
	s.Equal(15.0, proj1.AverageOutput)
	s.Equal(0.5, proj1.ContractThreshold)
	s.Equal(s.startTime, proj1.StartTime)
	s.Equal(s.endTime, proj1.EndTime)

	s.Equal("project2", proj2.ProjectID)
	s.Equal(30.0, proj2.AverageOutput)
	s.Equal(0.75, proj2.ContractThreshold)
	s.Equal(s.startTime, proj2.StartTime)
	s.Equal(s.endTime, proj2.EndTime)
}

func (s *AvgCacheTestSuite) TestGetProtoOutputs() {
	// Add some test data
	data1 := []types.RealTimeDERData{{
		ID: "der1",
		DER: types.DER{
			ProjectID:         "project1",
			DerID:             "der1",
			ContractThreshold: 0.5,
			IsOnline:          true,
			CurrentOutput:     10.0,
		},
	}}
	o1 := outcome.New(1, "task1", "project1", data1, 10.0, time.Second)
	s.cache.Add(o1)

	data2 := []types.RealTimeDERData{{
		ID: "der1",
		DER: types.DER{
			ProjectID:         "project1",
			DerID:             "der1",
			ContractThreshold: 0.5,
			IsOnline:          true,
			CurrentOutput:     20.0,
		},
	}}
	o2 := outcome.New(1, "task1", "project1", data2, 20.0, time.Second)
	s.cache.Add(o2)

	data3 := []types.RealTimeDERData{{
		ID: "der2",
		DER: types.DER{
			ProjectID:         "project2",
			DerID:             "der2",
			ContractThreshold: 0.75,
			IsOnline:          true,
			CurrentOutput:     30.0,
		},
	}}
	o3 := outcome.New(1, "task1", "project2", data3, 30.0, time.Second)
	s.cache.Add(o3)

	outputs := s.cache.GetProtoOutputs()
	s.Len(outputs, 2)

	// Sort outputs by project ID for consistent testing
	var proj1, proj2 *pb.AverageOutput
	for _, out := range outputs {
		if out.ProjectId == "project1" {
			proj1 = out
		} else {
			proj2 = out
		}
	}

	s.Equal("project1", proj1.ProjectId)
	s.Equal(15.0, proj1.AverageOutput)
	s.Equal(0.5, proj1.ContractThreshold)
	s.Equal(s.startTime.Format(time.RFC3339), proj1.StartTime)
	s.Equal(s.endTime.Format(time.RFC3339), proj1.EndTime)

	s.Equal("project2", proj2.ProjectId)
	s.Equal(30.0, proj2.AverageOutput)
	s.Equal(0.75, proj2.ContractThreshold)
	s.Equal(s.startTime.Format(time.RFC3339), proj2.StartTime)
	s.Equal(s.endTime.Format(time.RFC3339), proj2.EndTime)
}

func (s *AvgCacheTestSuite) TestReset() {
	// Add some test data
	data1 := []types.RealTimeDERData{{
		ID: "der1",
		DER: types.DER{
			ProjectID:         "project1",
			DerID:             "der1",
			ContractThreshold: 0.5,
			IsOnline:          true,
			CurrentOutput:     10.0,
		},
	}}
	o1 := outcome.New(1, "task1", "project1", data1, 10.0, time.Second)
	s.cache.Add(o1)

	data2 := []types.RealTimeDERData{{
		ID: "der2",
		DER: types.DER{
			ProjectID:         "project2",
			DerID:             "der2",
			ContractThreshold: 0.75,
			IsOnline:          true,
			CurrentOutput:     20.0,
		},
	}}
	o2 := outcome.New(1, "task1", "project2", data2, 20.0, time.Second)
	s.cache.Add(o2)
	s.Len(s.cache.items, 2)

	// Reset the cache
	s.cache.Reset(s.startTime, s.endTime)
	s.Empty(s.cache.items)

	// Add new data after reset
	data3 := []types.RealTimeDERData{{
		ID: "der3",
		DER: types.DER{
			ProjectID:         "project3",
			DerID:             "der3",
			ContractThreshold: 0.6,
			IsOnline:          true,
			CurrentOutput:     30.0,
		},
	}}
	o3 := outcome.New(1, "task1", "project3", data3, 30.0, time.Second)
	s.cache.Add(o3)
	s.Len(s.cache.items, 1)
	s.Contains(s.cache.items, "project3")
}

func TestAvgCacheSuite(t *testing.T) {
	suite.Run(t, new(AvgCacheTestSuite))
}
