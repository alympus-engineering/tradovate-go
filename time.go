package tradovate

import (
	"fmt"
	"strings"
	"time"
)

type Time struct {
	time.Time
}

func (ct *Time) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		ct.Time = time.Time{}
		return
	}
	ct.Time, err = time.Parse("2006-01-02T15:04:05", s)
	return
}

func (ct *Time) MarshalJSON() ([]byte, error) {
	if ct.Time.UnixNano() == 0 {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", ct.Time.UTC().Format("2006-01-02T15:04:05"))), nil
}