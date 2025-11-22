package domain

import "time"

type LinkStatus string

const (
	StatusAvailable    LinkStatus = "available"
	StatusNotAvailable LinkStatus = "not available"
	StatusUnknown      LinkStatus = "unknown"
)

type Link struct {
	URL    string
	Status LinkStatus
}

type LinkCheckTask struct {
	ID        int64
	Links     []Link
	CreatedAt time.Time
}
