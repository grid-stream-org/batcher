package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type NillableTimeTestSuite struct {
	suite.Suite
}

func (s *NillableTimeTestSuite) TestUnmarshalJSON() {
	testCases := []struct {
		name        string
		input       []byte
		expectError bool
		validate    func(*NillableTime)
	}{
		{
			name:        "valid RFC3339 time",
			input:       []byte(`"2024-01-13T15:04:05Z"`),
			expectError: false,
			validate: func(nt *NillableTime) {
				expected := time.Date(2024, 1, 13, 15, 4, 5, 0, time.UTC)
				s.Equal(expected, nt.Time)
			},
		},
		{
			name:        "null string",
			input:       []byte("null"),
			expectError: false,
			validate: func(nt *NillableTime) {
				// Since we use time.Now() for null values, we can only check
				// that the time is close to now
				diff := time.Since(nt.Time)
				s.Less(diff.Abs(), time.Second)
			},
		},
		{
			name:        "empty string in quotes",
			input:       []byte(`""`),
			expectError: false,
			validate: func(nt *NillableTime) {
				diff := time.Since(nt.Time)
				s.Less(diff.Abs(), time.Second)
			},
		},
		{
			name:        "invalid time format",
			input:       []byte(`"2024-13-01"`),
			expectError: true,
			validate: func(nt *NillableTime) {
				// No validation needed for error case
			},
		},
		{
			name:        "invalid JSON",
			input:       []byte(`invalid`),
			expectError: true,
			validate: func(nt *NillableTime) {
				// No validation needed for error case
			},
		},
		{
			name:        "with timezone",
			input:       []byte(`"2024-01-13T15:04:05+02:00"`),
			expectError: false,
			validate: func(nt *NillableTime) {
				expected := time.Date(2024, 1, 13, 15, 4, 5, 0, time.FixedZone("", 7200))
				s.Equal(expected, nt.Time)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var nt NillableTime
			err := nt.UnmarshalJSON(tc.input)

			if tc.expectError {
				s.Error(err)
			} else {
				s.NoError(err)
				tc.validate(&nt)
			}
		})
	}
}

func TestNillableTimeSuite(t *testing.T) {
	suite.Run(t, new(NillableTimeTestSuite))
}
