package destination

import (
	"context"

	"github.com/grid-stream-org/batcher/internal/outcome"
	pb "github.com/grid-stream-org/grid-stream-protos/gen/validator/v1"
	"github.com/stretchr/testify/mock"
)

type MockValidatorClient struct {
	mock.Mock
}

func (m *MockValidatorClient) SendAverages(ctx context.Context, averages []*pb.AverageOutput) error {
	args := m.Called(ctx, averages)
	return args.Error(0)
}

func (m *MockValidatorClient) NotifyProject(ctx context.Context, projectID string) error {
	args := m.Called(ctx, projectID)
	return args.Error(0)
}

func (m *MockValidatorClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockBuffer struct {
	mock.Mock
}

func (m *MockBuffer) Start(ctx context.Context) {
	m.Called(ctx)
}

func (m *MockBuffer) Stop() {
	m.Called()
}

func (m *MockBuffer) Add(ctx context.Context, outcome *outcome.Outcome) {
	m.Called(ctx, outcome)
}

func (m *MockBuffer) Flush(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
