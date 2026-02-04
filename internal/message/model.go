package message

import "time"

type Sender string

const (
	SenderVisitor Sender = "visitor"
	SenderOwner   Sender = "owner"
)

type Message struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Sender    Sender    `json:"sender"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
