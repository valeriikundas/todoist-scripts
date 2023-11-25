package todoist_utils

import (
	"io"
	"log"
	"os"
	"strings"
	"time"
)

func ReadApiTokenFromDotenv() string {
	// FIXME: rewrite with `github.com/joho/godotenv`
	file, err := os.Open(".env")
	if err != nil {
		log.Fatal(err)
	}
	b, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	text := string(b)

	split := strings.Split(text, "=")
	apiToken := split[1]
	return apiToken
}

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
