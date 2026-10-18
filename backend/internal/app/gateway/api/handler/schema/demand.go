package schema

import (
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	demand "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/demand"
)

type CreateDemandRequest struct {
	AssigneeID         string          `json:"assignee_id"`
	ReviewerID         string          `json:"reviewer_id"`
	Title              string          `json:"title"`
	Description        string          `json:"description"`
	Kind               entity.Kind     `json:"kind"`
	Priority           entity.Priority `json:"priority"`
	Reference          string          `json:"reference"`
	PackageID          string          `json:"package_id"`
	ReleaseID          string          `json:"release_id"`
	Product            string          `json:"product"`
	AcceptanceCriteria string          `json:"acceptance_criteria"`
	DueDate            *time.Time      `json:"due_date"`
	EventCodes         []string        `json:"event_codes"`
}

func (r CreateDemandRequest) ToInput() demand.CreateInput {
	return demand.CreateInput(r)
}

type MoveDemandRequest struct {
	Status entity.Status `json:"status"`
	Actor  string        `json:"actor"`
	Reason string        `json:"reason"`
}

func (r MoveDemandRequest) ToInput() demand.MoveInput {
	return demand.MoveInput{Status: r.Status, Actor: r.Actor, Reason: r.Reason}
}

type TestCaseRequest struct {
	Title    string `json:"title"`
	Steps    string `json:"steps"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	Result   string `json:"result"`
	Author   string `json:"author"`
}

func (r TestCaseRequest) ToInput() demand.AddTestInput {
	return demand.AddTestInput{Title: r.Title, Steps: r.Steps, Expected: r.Expected, Actual: r.Actual, Result: r.Result, Author: r.Author}
}
