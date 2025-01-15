package destination

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/grid-stream-org/batcher/internal/buffer"
	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/internal/outcome"
	"github.com/grid-stream-org/batcher/internal/types"
	"github.com/stretchr/testify/suite"
)

type FileDestinationTestSuite struct {
	suite.Suite
	ctx      context.Context
	dest     *fileDestination
	mockVC   *MockValidatorClient
	log      *slog.Logger
	tempDir  string
	filePath string
}

func (s *FileDestinationTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.mockVC = new(MockValidatorClient)
	s.log = slog.Default()

	// Create temp directory for test files
	tempDir, err := os.MkdirTemp("", "file-destination-test-*")
	s.NoError(err)
	s.tempDir = tempDir
	s.filePath = filepath.Join(tempDir, "test.json")

	cfg := &config.Destination{
		Type: "file",
		Path: s.filePath,
		Buffer: &config.Buffer{
			StartTime: time.Now(),
			Interval:  time.Minute,
			Offset:    time.Second * 30,
		},
	}

	dest, err := newFileDestination(s.ctx, cfg, s.mockVC, s.log)
	s.NoError(err)
	s.dest = dest.(*fileDestination)
}

func (s *FileDestinationTestSuite) TearDownTest() {
	if s.dest != nil {
		s.dest.Close()
	}
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
	}
}

func (s *FileDestinationTestSuite) TestNewFileDestination() {
	testCases := []struct {
		name        string
		setupPath   func() string
		expectError bool
	}{
		{
			name: "valid path",
			setupPath: func() string {
				return filepath.Join(s.tempDir, "test1.json")
			},
			expectError: false,
		},
		{
			name: "nested path creation",
			setupPath: func() string {
				return filepath.Join(s.tempDir, "nested", "dir", "test2.json")
			},
			expectError: false,
		},
		{
			name: "invalid path",
			setupPath: func() string {
				return string([]byte{0x00}) // invalid path character
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cfg := &config.Destination{
				Type: "file",
				Path: tc.setupPath(),
				Buffer: &config.Buffer{
					StartTime: time.Now(),
					Interval:  time.Minute,
					Offset:    time.Second * 30,
				},
			}

			dest, err := newFileDestination(s.ctx, cfg, s.mockVC, s.log)
			if tc.expectError {
				s.Error(err)
				s.Nil(dest)
			} else {
				s.NoError(err)
				s.NotNil(dest)
				fd := dest.(*fileDestination)
				s.NotNil(fd.file)
				s.NotNil(fd.encoder)
				fd.Close()
			}
		})
	}
}

func (s *FileDestinationTestSuite) TestAdd() {
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
				s.mockVC.On("NotifyProject", s.ctx, "project1").Return(nil).Once()
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

func (s *FileDestinationTestSuite) TestFlushFunc() {
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

	// Test the flush function
	err := s.dest.flushFunc(s.ctx, testData)
	s.NoError(err)

	// Verify file contents
	err = s.dest.file.Sync()
	s.NoError(err)

	contents, err := os.ReadFile(s.filePath)
	s.NoError(err)

	var decoded buffer.FlushOutcome
	err = json.Unmarshal(contents, &decoded)
	s.NoError(err)

	s.Equal(testData.Outcomes[0].Success, decoded.Outcomes[0].Success)
	s.Equal(testData.Outcomes[0].WorkerID, decoded.Outcomes[0].WorkerID)
	s.Equal(testData.Outcomes[0].TaskID, decoded.Outcomes[0].TaskID)
	s.Equal(testData.Outcomes[0].ProjectID, decoded.Outcomes[0].ProjectID)

	s.Equal(testData.AvgOutputs[0].ProjectID, decoded.AvgOutputs[0].ProjectID)
	s.Equal(testData.AvgOutputs[0].AverageOutput, decoded.AvgOutputs[0].AverageOutput)
}

func (s *FileDestinationTestSuite) TestClose() {
	// Write something to ensure file is open
	err := s.dest.flushFunc(s.ctx, &buffer.FlushOutcome{})
	s.NoError(err)

	// Test close
	err = s.dest.Close()
	s.NoError(err)

	// Verify file is closed by trying to write again
	err = s.dest.flushFunc(s.ctx, &buffer.FlushOutcome{})
	s.Error(err) // Should fail because file is closed
}

func TestFileDestinationSuite(t *testing.T) {
	suite.Run(t, new(FileDestinationTestSuite))
}
