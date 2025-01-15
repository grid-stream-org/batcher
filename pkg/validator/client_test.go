package validator

import (
	"context"
	"testing"

	pb "github.com/grid-stream-org/grid-stream-protos/gen/validator/v1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

type mockValidatorServiceClient struct {
	mock.Mock
}

func (m *mockValidatorServiceClient) ValidateAverageOutputs(ctx context.Context, req *pb.ValidateAverageOutputsRequest, opts ...grpc.CallOption) (*pb.ValidateAverageOutputsResponse, error) {
	args := m.Called(ctx, req, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.ValidateAverageOutputsResponse), args.Error(1)
}

func (m *mockValidatorServiceClient) NotifyProject(ctx context.Context, req *pb.NotifyProjectRequest, opts ...grpc.CallOption) (*pb.NotifyProjectResponse, error) {
	args := m.Called(ctx, req, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.NotifyProjectResponse), args.Error(1)
}

type ValidatorTestSuite struct {
	suite.Suite
	mockClient *mockValidatorServiceClient
	client     *validatorClient
	ctx        context.Context
}

func (s *ValidatorTestSuite) SetupTest() {
	s.mockClient = new(mockValidatorServiceClient)
	s.client = &validatorClient{
		cfg:    &Config{Host: "localhost", Port: 8080},
		client: s.mockClient,
	}
	s.ctx = context.Background()
}

func (s *ValidatorTestSuite) TestSendAverages() {
	testCases := []struct {
		name        string
		averages    []*pb.AverageOutput
		setupMock   func()
		expectError bool
		errorType   error
	}{
		{
			name: "successful validation",
			averages: []*pb.AverageOutput{
				{ProjectId: "test1"},
			},
			setupMock: func() {
				s.mockClient.On("ValidateAverageOutputs",
					s.ctx,
					&pb.ValidateAverageOutputsRequest{
						AverageOutputs: []*pb.AverageOutput{{ProjectId: "test1"}},
					},
					mock.Anything,
				).Return(&pb.ValidateAverageOutputsResponse{
					Success: true,
				}, nil)
			},
			expectError: false,
		},
		{
			name: "validation errors",
			averages: []*pb.AverageOutput{
				{ProjectId: "test2"},
			},
			setupMock: func() {
				s.mockClient.On("ValidateAverageOutputs",
					s.ctx,
					&pb.ValidateAverageOutputsRequest{
						AverageOutputs: []*pb.AverageOutput{{ProjectId: "test2"}},
					},
					mock.Anything,
				).Return(&pb.ValidateAverageOutputsResponse{
					Success: false,
					Errors: []*pb.ValidationError{
						{ProjectId: "test2", Message: "validation error"},
					},
				}, nil)
			},
			expectError: true,
			errorType:   &ValidationErrors{},
		},
		{
			name: "grpc error",
			averages: []*pb.AverageOutput{
				{ProjectId: "test3"},
			},
			setupMock: func() {
				s.mockClient.On("ValidateAverageOutputs",
					s.ctx,
					&pb.ValidateAverageOutputsRequest{
						AverageOutputs: []*pb.AverageOutput{{ProjectId: "test3"}},
					},
					mock.Anything,
				).Return(nil, errors.New("grpc error"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.setupMock()
			err := s.client.SendAverages(s.ctx, tc.averages)
			if tc.expectError {
				s.Error(err)
				if tc.errorType != nil {
					s.IsType(tc.errorType, err)
				}
			} else {
				s.NoError(err)
			}
			s.mockClient.AssertExpectations(s.T())
		})
	}
}

func (s *ValidatorTestSuite) TestNotifyProject() {
	testCases := []struct {
		name        string
		projectID   string
		setupMock   func()
		expectError bool
		errorType   error
	}{
		{
			name:      "successful notification",
			projectID: "test1",
			setupMock: func() {
				s.mockClient.On("NotifyProject",
					s.ctx,
					&pb.NotifyProjectRequest{ProjectId: "test1"},
					mock.Anything,
				).Return(&pb.NotifyProjectResponse{
					Acknowledged: true,
				}, nil)
			},
			expectError: false,
		},
		{
			name:      "notification errors",
			projectID: "test2",
			setupMock: func() {
				s.mockClient.On("NotifyProject",
					s.ctx,
					&pb.NotifyProjectRequest{ProjectId: "test2"},
					mock.Anything,
				).Return(&pb.NotifyProjectResponse{
					Acknowledged: false,
					Errors: []*pb.NotifyProjectError{
						{ProjectId: "test2", Message: "notification error"},
					},
				}, nil)
			},
			expectError: true,
			errorType:   &NotifyProjectErrors{},
		},
		{
			name:      "grpc error",
			projectID: "test3",
			setupMock: func() {
				s.mockClient.On("NotifyProject",
					s.ctx,
					&pb.NotifyProjectRequest{ProjectId: "test3"},
					mock.Anything,
				).Return(nil, errors.New("grpc error"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.setupMock()
			err := s.client.NotifyProject(s.ctx, tc.projectID)
			if tc.expectError {
				s.Error(err)
				if tc.errorType != nil {
					s.IsType(tc.errorType, err)
				}
			} else {
				s.NoError(err)
			}
			s.mockClient.AssertExpectations(s.T())
		})
	}
}

func (s *ValidatorTestSuite) TestConfigValidate() {
	testCases := []struct {
		name        string
		cfg         *Config
		expectError bool
	}{
		{
			name: "valid config without TLS",
			cfg: &Config{
				Host: "localhost",
				Port: 8080,
			},
			expectError: false,
		},
		{
			name: "valid config with TLS",
			cfg: &Config{
				Host: "localhost",
				Port: 8080,
				TLSConfig: &TLSConfig{
					Enabled:  true,
					CertPath: "cert.pem",
					KeyPath:  "key.pem",
				},
			},
			expectError: false,
		},
		{
			name: "invalid port",
			cfg: &Config{
				Host: "localhost",
				Port: 0,
			},
			expectError: true,
		},
		{
			name: "missing cert path with TLS enabled",
			cfg: &Config{
				Host: "localhost",
				Port: 8080,
				TLSConfig: &TLSConfig{
					Enabled: true,
					KeyPath: "key.pem",
				},
			},
			expectError: true,
		},
		{
			name: "missing key path with TLS enabled",
			cfg: &Config{
				Host: "localhost",
				Port: 8080,
				TLSConfig: &TLSConfig{
					Enabled:  true,
					CertPath: "cert.pem",
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := tc.cfg.Validate()
			if tc.expectError {
				s.Error(err)
			} else {
				s.NoError(err)
			}
		})
	}
}

func TestValidatorSuite(t *testing.T) {
	suite.Run(t, new(ValidatorTestSuite))
}