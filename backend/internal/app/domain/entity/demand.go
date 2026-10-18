package entity

import (
	"regexp"
	"strings"
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
)

type (
	Status   string
	Kind     string
	Priority string
)

const (
	Backlog          Status = "backlog"
	InProgress       Status = "in_progress"
	InQA             Status = "in_qa"
	ChangesRequested Status = "changes_requested"
	Done             Status = "done"
	Improvement      Kind   = "improvement"
	Corrective       Kind   = "corrective"
	Testing          Kind   = "testing"
	Rewrite          Kind   = "rewrite"
)

var transitions = map[Status]map[Status]bool{
	Backlog:          {InProgress: true},
	InProgress:       {InQA: true},
	InQA:             {ChangesRequested: true, Done: true},
	ChangesRequested: {InProgress: true},
}

func CanTransition(from, to Status) bool { return transitions[from][to] }

var kindCodePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)

func IsKind(k Kind) bool { return kindCodePattern.MatchString(string(k)) }
func IsStatus(s Status) bool {
	return s == Backlog || s == InProgress || s == InQA || s == ChangesRequested || s == Done
}

func IsPriority(p Priority) bool {
	return p == "low" || p == "medium" || p == "high" || p == "critical"
}

type Demand struct {
	KindName           string     `json:"kind_name"`
	AssigneeID         string     `json:"assignee_id"`
	ReviewerID         string     `json:"reviewer_id"`
	ReleaseID          string     `json:"release_id"`
	ReleaseName        string     `json:"release_name"`
	ID                 string     `json:"id"`
	Number             int64      `json:"number"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	Kind               Kind       `json:"kind"`
	Priority           Priority   `json:"priority"`
	Status             Status     `json:"status"`
	Assignee           string     `json:"assignee"`
	Reviewer           string     `json:"reviewer"`
	Reference          string     `json:"reference"`
	PackageID          string     `json:"package_id"`
	Product            string     `json:"product"`
	AcceptanceCriteria string     `json:"acceptance_criteria"`
	DueDate            *time.Time `json:"due_date"`
	EventCodes         []string   `json:"event_codes"`
	QACycle            int        `json:"qa_cycle"`
	ReturnCount        int        `json:"return_count"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	CompletedAt        *time.Time `json:"completed_at"`
}

func (d Demand) Validate() error {
	if len(strings.TrimSpace(d.Title)) < 5 || len(d.Title) > 180 {
		return erring.ErrInvalidInput.WithMessage("O título deve conter de 5 a 180 caracteres.")
	}

	if len(d.Description) > 20000 {
		return erring.ErrInvalidInput.WithMessage("A descrição excede o tamanho permitido.")
	}

	if !IsKind(d.Kind) {
		return erring.ErrInvalidInput.WithMessage("Esteira inválida.")
	}

	if !IsPriority(d.Priority) {
		return erring.ErrInvalidInput.WithMessage("Prioridade inválida.")
	}

	if len(d.Assignee) > 120 || len(d.Reviewer) > 120 || len(d.Reference) > 120 || len(d.Product) > 120 {
		return erring.ErrInvalidInput.WithMessage("Um dos campos excede o tamanho permitido.")
	}

	if len(d.AcceptanceCriteria) > 20000 {
		return erring.ErrInvalidInput.WithMessage("Os critérios de aceite excedem o tamanho permitido.")
	}

	if len(d.EventCodes) > 30 {
		return erring.ErrInvalidInput.WithMessage("Quantidade de códigos de evento acima do limite.")
	}

	for _, event := range d.EventCodes {
		if len(event) > 30 {
			return erring.ErrInvalidInput.WithMessage("Código de evento inválido.")
		}
	}

	return nil
}

type TestCase struct {
	AuthorID  string    `json:"author_id"`
	ID        string    `json:"id"`
	DemandID  string    `json:"demand_id"`
	Cycle     int       `json:"cycle"`
	Title     string    `json:"title"`
	Steps     string    `json:"steps"`
	Expected  string    `json:"expected"`
	Actual    string    `json:"actual"`
	Result    string    `json:"result"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

func (t TestCase) Validate() error {
	if len(strings.TrimSpace(t.Title)) < 3 || len(t.Title) > 200 {
		return erring.ErrInvalidInput.WithMessage("O título do teste deve conter de 3 a 200 caracteres.")
	}

	if t.Result != "passed" && t.Result != "failed" {
		return erring.ErrInvalidInput.WithMessage("Resultado de teste inválido.")
	}

	if len(t.Steps) > 10000 || len(t.Expected) > 10000 || len(t.Actual) > 10000 || len(t.Author) > 120 {
		return erring.ErrInvalidInput.WithMessage("Um dos campos do teste excede o tamanho permitido.")
	}

	if t.Result == "failed" && strings.TrimSpace(t.Actual) == "" {
		return erring.ErrInvalidInput.WithMessage("Informe o resultado obtido no teste reprovado.")
	}

	return nil
}

type Transition struct {
	ActorID   string    `json:"actor_id"`
	ID        string    `json:"id"`
	DemandID  string    `json:"demand_id"`
	From      Status    `json:"from_status"`
	To        Status    `json:"to_status"`
	Actor     string    `json:"actor"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}
