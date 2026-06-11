package adapter

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

func rawString(value json.RawMessage) string {
	value = bytes.TrimSpace(value)
	if len(value) == 0 || bytes.Equal(value, []byte("null")) {
		return ""
	}
	var text string
	if err := json.Unmarshal(value, &text); err == nil {
		return strings.TrimSpace(text)
	}
	var number json.Number
	if err := json.Unmarshal(value, &number); err == nil {
		return number.String()
	}
	return ""
}

func rawFloat(value json.RawMessage) *float64 {
	text := rawString(value)
	if text == "" {
		return nil
	}
	parsed, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return nil
	}
	return &parsed
}

func rawInt(value json.RawMessage) *int {
	text := rawString(value)
	if text == "" {
		return nil
	}
	parsed, err := strconv.Atoi(text)
	if err != nil {
		return nil
	}
	return &parsed
}

func parseTimeOrNow(value string, now time.Time) time.Time {
	if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return parsed
	}
	return now.UTC()
}

func rankPointer(rank int) *int {
	if rank <= 0 {
		return nil
	}
	return &rank
}
