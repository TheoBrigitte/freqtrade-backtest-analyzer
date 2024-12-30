package main

import (
	"strings"
	"time"
)

type CustomTime struct {
	time.Time
}

const dateTimeFormat = "2006-01-02 15:04:05"

func (t *CustomTime) UnmarshalJSON(b []byte) (err error) {
	value := strings.Trim(string(b), `"`)
	if value == "" || value == "null" {
		return nil
	}

	date, err := time.Parse(dateTimeFormat, value)
	if err != nil {
		return err
	}
	t.Time = date
	return
}
