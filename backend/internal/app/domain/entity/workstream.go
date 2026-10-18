package entity

import "time"

// Workstream is a configurable category. Workflow statuses remain independent.
type Workstream struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Member is the directory exposed for assignment, without credentials.
type Member struct {
	Email string `json:"email"`
	ID    string `json:"id"`
	Name  string `json:"name"`
}
