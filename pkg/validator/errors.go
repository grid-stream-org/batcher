package validator

import (
	"fmt"
	"strings"

	pb "github.com/grid-stream-org/grid-stream-protos/gen/validator/v1"
)

type ValidationErrors struct {
	Errors []*pb.ValidationError
}

func (ve *ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve.Errors {
		messages = append(messages, fmt.Sprintf("project %s: %s", err.ProjectId, err.Message))
	}
	return "validation failed: " + strings.Join(messages, "; ")
}

type NotifyProjectErrors struct {
	Errors []*pb.NotifyProjectError
}

func (ne *NotifyProjectErrors) Error() string {
	var messages []string
	for _, err := range ne.Errors {
		messages = append(messages, fmt.Sprintf("project %s: %s", err.ProjectId, err.Message))
	}
	return "failed to notify validator: " + strings.Join(messages, "; ")
}
