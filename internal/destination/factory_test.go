package destination

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/stretchr/testify/suite"
)

type FactoryTestSuite struct {
	suite.Suite
	ctx context.Context
	log *slog.Logger
}

func (s *FactoryTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.log = slog.Default()
}

func (s *FactoryTestSuite) TestNewDestination() {
	testCases := []struct {
		name        string
		cfg         *config.Destination
		expectError bool
	}{
		{
			name: "stdout destination",
			cfg: &config.Destination{
				Type: "stdout",
				Buffer: &config.Buffer{
					StartTime: time.Now(),
					Interval:  time.Minute,
					Offset:    time.Second * 30,
				},
			},
			expectError: false,
		},
		{
			name: "invalid type",
			cfg: &config.Destination{
				Type: "invalid",
				Buffer: &config.Buffer{
					StartTime: time.Now(),
					Interval:  time.Minute,
					Offset:    time.Second * 30,
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			dest, err := NewDestination(s.ctx, tc.cfg, s.log)
			if tc.expectError {
				s.Error(err)
				s.Nil(dest)
				s.Contains(err.Error(), "invalid destination type")
			} else {
				s.NoError(err)
				s.NotNil(dest)

				// Test specific type
				switch tc.cfg.Type {
				case "stdout":
					s.IsType(&stdoutDestination{}, dest)
				}
			}
		})
	}
}

func TestFactorySuite(t *testing.T) {
	suite.Run(t, new(FactoryTestSuite))
}
