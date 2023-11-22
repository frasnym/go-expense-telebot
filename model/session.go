package model

import (
	"time"
)

type Session struct {
	Action    string
	ChatID    int64
	MessageID int
	StartTime time.Time
}
