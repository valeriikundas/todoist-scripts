package todoist

import (
	"time"
)

type TimeParser struct {
	time.Time
}

func (tp *TimeParser) UnmarshalJSON(b []byte) (err error) {
	time, err := time.Parse(`"2006-01-02T15:04:05.000000Z"`, string(b))
	if err != nil {
		return err
	}
	tp.Time = time
	return
}
