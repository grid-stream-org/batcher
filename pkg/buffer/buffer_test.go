package buffer

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type BufferTestSuite struct {
	suite.Suite
	ctx    context.Context
	cancel context.CancelFunc
}

func (s *BufferTestSuite) SetupTest() {
	s.ctx, s.cancel = context.WithCancel(context.Background())
}

func (s *BufferTestSuite) TearDownTest() {
	s.cancel()
}

func (s *BufferTestSuite) TestNew() {
	testCases := []struct {
		name        string
		opts        []Option[int]
		expectError bool
	}{
		{
			name: "Basic buffer creation",
		},
		{
			name: "With valid flush function",
			opts: []Option[int]{
				WithFlushFunc(func(ctx context.Context, data []int) error {
					return nil
				}),
			},
		},
		{
			name: "With nil flush function",
			opts: []Option[int]{
				WithFlushFunc[int](nil),
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			buf, err := New(slog.Default(), tc.opts...)
			if tc.expectError {
				s.Error(err)
				s.Nil(buf)
			} else {
				s.NoError(err)
				s.NotNil(buf)
			}
		})
	}
}

func (s *BufferTestSuite) TestAddAndFlush() {
	var flushedData []string
	flushFunc := func(ctx context.Context, data []string) error {
		flushedData = make([]string, len(data))
		copy(flushedData, data)
		return nil
	}

	buf, err := New(slog.Default(), WithFlushFunc(flushFunc))
	s.NoError(err)

	testData := []string{"test1", "test2", "test3"}
	for _, data := range testData {
		buf.Add(data)
	}

	// First flush should contain our test data
	err = buf.Flush(context.Background())
	s.NoError(err)
	s.Equal(testData, flushedData)

	// Reset flushedData
	flushedData = nil

	// Second flush should be empty since buffer was cleared
	err = buf.Flush(context.Background())
	s.NoError(err)
	s.Empty(flushedData)
}

func (s *BufferTestSuite) TestStartStop() {
	var flushedData []int
	flushFunc := func(ctx context.Context, data []int) error {
		flushedData = data
		return nil
	}

	buf, err := New(slog.Default(), WithFlushFunc(flushFunc))
	s.NoError(err)

	buf.Start(s.ctx, 100*time.Millisecond)

	testData := []int{1, 2, 3, 4, 5}
	for _, d := range testData {
		buf.Add(d)
	}

	time.Sleep(150 * time.Millisecond)
	buf.Stop()

	s.ElementsMatch(testData, flushedData)

	// Verify no more flushes after stop
	buf.Add(6)
	time.Sleep(150 * time.Millisecond)
	s.NotContains(flushedData, 6)
}

func (s *BufferTestSuite) TestConcurrentAccess() {
	var (
		flushedData []int
		mu          sync.Mutex
	)

	flushFunc := func(ctx context.Context, data []int) error {
		mu.Lock()
		flushedData = append(flushedData, data...)
		mu.Unlock()
		return nil
	}

	buf, err := New(slog.Default(), WithFlushFunc(flushFunc))
	s.NoError(err)

	buf.Start(s.ctx, 100*time.Millisecond)

	var wg sync.WaitGroup
	numWriters := 10
	itemsPerWriter := 100

	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for j := 0; j < itemsPerWriter; j++ {
				buf.Add(writerID*itemsPerWriter + j)
				time.Sleep(time.Microsecond)
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(200 * time.Millisecond)
	buf.Stop()

	mu.Lock()
	s.Equal(numWriters*itemsPerWriter, len(flushedData))
	mu.Unlock()
}

func (s *BufferTestSuite) TestDoubleStartStop() {
	buf, err := New[int](slog.Default())
	s.NoError(err)

	buf.Start(s.ctx, time.Second)
	buf.Start(s.ctx, time.Second) // Should not panic
	buf.Stop()
	buf.Stop() // Should not panic
}

func TestBufferSuite(t *testing.T) {
	suite.Run(t, new(BufferTestSuite))
}
