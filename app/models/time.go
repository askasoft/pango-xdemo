package models

import "time"

const (
	DateFormat = "2006-01-02"
	TimeFormat = "2006-01-02 15:04:05"
)

func FormatDate(t time.Time) string {
	return t.Local().Format(DateFormat)
}

func FormatTime(t time.Time) string {
	return t.Local().Format(TimeFormat)
}
