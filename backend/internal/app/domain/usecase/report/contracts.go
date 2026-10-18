package report

import (
	"context"
	"encoding/json"
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	demand "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/demand"
)

type Request struct {
	Prompt           string `json:"prompt"`
	DemandID         string `json:"demand_id"`
	From             string `json:"from"`
	To               string `json:"to"`
	PreviousReportID string `json:"previous_report_id"`
}
type Metric struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}
type Section struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}
type Point struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
}
type Chart struct {
	Title  string  `json:"title"`
	Kind   string  `json:"kind"`
	Points []Point `json:"points"`
}
type Table struct {
	Title   string     `json:"title"`
	Columns []string   `json:"columns"`
	Rows    [][]string `json:"rows"`
}
type Content struct {
	HTML            string    `json:"html,omitempty"`
	Title           string    `json:"title"`
	Summary         string    `json:"summary"`
	Metrics         []Metric  `json:"metrics"`
	Sections        []Section `json:"sections"`
	Charts          []Chart   `json:"charts"`
	Tables          []Table   `json:"tables"`
	Recommendations []string  `json:"recommendations"`
	MissingData     []string  `json:"missing_data"`
	SourceIDs       []string  `json:"source_ids"`
}
type Report struct {
	ID         string          `json:"id"`
	UserID     string          `json:"-"`
	DemandID   string          `json:"demand_id"`
	Prompt     string          `json:"prompt"`
	CreatedAt  time.Time       `json:"created_at"`
	Content    Content         `json:"content"`
	SourceData json.RawMessage `json:"source_data"`
}
type Repository interface {
	Save(ctx context.Context, item Report) error
	List(ctx context.Context, userID, demandID string) ([]Report, error)
	Get(ctx context.Context, userID, id string) (Report, error)
}
type Data interface {
	List(ctx context.Context, kind, status, query string) ([]entity.Demand, error)
	Detail(ctx context.Context, id string) (demand.Detail, error)
}
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}
type Toolset interface {
	Definitions() []Tool
	Call(ctx context.Context, name string, args json.RawMessage) (any, error)
	SourceIDs() []string
}
type (
	Progress  func(agent, message string)
	Generator interface {
		Generate(ctx context.Context, prompt string, tools Toolset, progress Progress) (Content, error)
	}
)
