package entity

import "time"

type Notification struct {
	ID        string     `json:"id"`
	UserID    string     `json:"-"`
	DemandID  string     `json:"demand_id"`
	Kind      string     `json:"kind"`
	Title     string     `json:"title"`
	Message   string     `json:"message"`
	ReadAt    *time.Time `json:"read_at"`
	CreatedAt time.Time  `json:"created_at"`
}
