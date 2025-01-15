package buffer

import (
	"testing"
	"time"

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
	// First addition should return false (new item)
	exists := s.cache.Add("project1", 10.0)
	s.False(exists)
	s.Len(s.cache.items, 1)

	// Second addition should return true (existing item)
	exists = s.cache.Add("project1", 20.0)
	s.True(exists)
	s.Len(s.cache.items, 1)

	// Different project should return false
	exists = s.cache.Add("project2", 30.0)
	s.False(exists)
	s.Len(s.cache.items, 2)

	// Verify the averages
	s.Equal(15.0, s.cache.items["project1"].average.AverageOutput) // (10 + 20) / 2
	s.Equal(30.0, s.cache.items["project2"].average.AverageOutput)
}

func (s *AvgCacheTestSuite) TestGetOutputs() {
	// Add some test data
	s.cache.Add("project1", 10.0)
	s.cache.Add("project1", 20.0)
	s.cache.Add("project2", 30.0)

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
	s.Equal(s.startTime, proj1.StartTime)
	s.Equal(s.endTime, proj1.EndTime)

	s.Equal("project2", proj2.ProjectID)
	s.Equal(30.0, proj2.AverageOutput)
	s.Equal(s.startTime, proj2.StartTime)
	s.Equal(s.endTime, proj2.EndTime)
}

func (s *AvgCacheTestSuite) TestGetProtoOutputs() {
	// Add some test data
	s.cache.Add("project1", 10.0)
	s.cache.Add("project1", 20.0)
	s.cache.Add("project2", 30.0)

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
	s.Equal(s.startTime.Format(time.RFC3339), proj1.StartTime)
	s.Equal(s.endTime.Format(time.RFC3339), proj1.EndTime)

	s.Equal("project2", proj2.ProjectId)
	s.Equal(30.0, proj2.AverageOutput)
	s.Equal(s.startTime.Format(time.RFC3339), proj2.StartTime)
	s.Equal(s.endTime.Format(time.RFC3339), proj2.EndTime)
}

func (s *AvgCacheTestSuite) TestReset() {
	// Add some test data
	s.cache.Add("project1", 10.0)
	s.cache.Add("project2", 20.0)
	s.Len(s.cache.items, 2)

	// Reset the cache
	s.cache.Reset()
	s.Empty(s.cache.items)

	// Add new data after reset
	s.cache.Add("project3", 30.0)
	s.Len(s.cache.items, 1)
	s.Contains(s.cache.items, "project3")
}

func TestAvgCacheSuite(t *testing.T) {
	suite.Run(t, new(AvgCacheTestSuite))
}
