package types

import (
	"strings"
	"time"
)

type NillableTime struct {
	time.Time
}

func (nt *NillableTime) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == "null" || s == `""` {
		nt.Time = time.Now()
		return nil
	}

	s = strings.Trim(s, `"`)

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	nt.Time = t
	return nil
}
