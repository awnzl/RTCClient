package domain

import "time"

const (
	Join int = iota
	Leave
	Send
)

type Message struct {
	Content string    `json:"content"`
	Room    string    `json:"room"`
	Time    time.Time `json:"time"`
	Type    int       `json:"type"`
}
