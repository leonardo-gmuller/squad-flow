package entity

import (
	"strings"
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
)

// WorkPackage is a delivery container, not a workflow item. It has no QA cycle.
// The actual execution is always recorded in one or more Demand records.
type WorkPackage struct {
	ID          string    `json:"id"`
	Reference   string    `json:"reference"` // External OS or epic number; optional for internal packages
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Initiative  string    `json:"initiative"` // e.g. a technical note or independent improvement project
	Kind        Kind      `json:"kind"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Total       int       `json:"total"`
	Done        int       `json:"done"`
	InQA        int       `json:"in_qa"`
}

func (p WorkPackage) Validate() error {
	if len(strings.TrimSpace(p.Title)) < 5 || len(p.Title) > 180 {
		return erring.ErrInvalidInput.WithMessage("O título da OS deve conter de 5 a 180 caracteres.")
	}

	if !IsKind(p.Kind) {
		return erring.ErrInvalidInput.WithMessage("Esteira da OS inválida.")
	}

	if len(p.Reference) > 120 || len(p.Initiative) > 160 || len(p.Description) > 20000 {
		return erring.ErrInvalidInput.WithMessage("Um dos campos da OS excede o tamanho permitido.")
	}

	return nil
}
