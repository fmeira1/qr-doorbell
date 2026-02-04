package session

import "time"

type Status string

const (
	StatusOpen   Status = "open"
	StatusClosed Status = "closed"
)

type Session struct {
	ID         string    `json:"id"`
	DoorbellID string    `json:"doorbell_id"`
	UnitID     string    `json:"unit_id"`
	Status     Status    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}
