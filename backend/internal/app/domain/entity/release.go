package entity

import "time"

// Release groups deliveries by code version independently of their OS.
type Release struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	ReleaseDate *time.Time `json:"release_date"`
	ArchivedAt  *time.Time `json:"archived_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Total       int        `json:"total"`
	Done        int        `json:"done"`
	InQA        int        `json:"in_qa"`
}
