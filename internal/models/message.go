package models

import "time"

// Message represents an inter-agent message
type Message struct {
	ID           string
	Sender       string
	Recipient    string
	Subject      string
	Body         string
	Timestamp    time.Time
	Read         bool
	CommissionID string
}
