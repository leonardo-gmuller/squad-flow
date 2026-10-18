package catalog

import (
	"context"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/auth"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
)

type Repository interface {
	List(ctx context.Context) ([]entity.Workstream, error)
	Members(ctx context.Context) ([]entity.Member, error)
	Insert(ctx context.Context, item entity.Workstream) (entity.Workstream, error)
	Update(ctx context.Context, id string, input Input) (entity.Workstream, error)
}
type Input struct {
	Name   string `json:"name"`
	Color  string `json:"color"`
	Active *bool  `json:"active"`
}
type UseCase struct{ repo Repository }

func New(repo Repository) *UseCase { return &UseCase{repo: repo} }
func (s *UseCase) List(ctx context.Context) ([]entity.Workstream, error) {
	items, err := s.repo.List(ctx)

	return items, erring.Wrap(err, "catalog.List")
}

func (s *UseCase) Members(ctx context.Context) ([]entity.Member, error) {
	items, err := s.repo.Members(ctx)

	return items, erring.Wrap(err, "catalog.Members")
}

var colorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func validate(input *Input) error {
	input.Name = strings.TrimSpace(input.Name)
	if n := utf8.RuneCountInString(input.Name); n < 2 || n > 80 || !colorPattern.MatchString(input.Color) {
		return erring.ErrInvalidInput.WithMessage("informe um nome de 2 a 80 caracteres e uma cor hexadecimal válida")
	}

	return nil
}

func (s *UseCase) Create(ctx context.Context, actor entity.User, input Input) (entity.Workstream, error) {
	if actor.Role != "admin" {
		return entity.Workstream{}, auth.ErrForbidden
	}

	if err := validate(&input); err != nil {
		return entity.Workstream{}, err
	}

	id, err := identifier.NewUUID()
	if err != nil {
		return entity.Workstream{}, erring.Wrap(err, "catalog.Create.id")
	}

	active := true
	if input.Active != nil {
		active = *input.Active
	}

	item, err := s.repo.Insert(ctx, entity.Workstream{ID: id, Code: id, Name: input.Name, Color: input.Color, Active: active})

	return item, erring.Wrap(err, "catalog.Create")
}

func (s *UseCase) Update(ctx context.Context, actor entity.User, id string, input Input) (entity.Workstream, error) {
	if actor.Role != "admin" {
		return entity.Workstream{}, auth.ErrForbidden
	}

	if !identifier.IsUUID(id) {
		return entity.Workstream{}, erring.ErrInvalidInput.WithMessage("esteira inválida")
	}

	if err := validate(&input); err != nil {
		return entity.Workstream{}, err
	}

	item, err := s.repo.Update(ctx, id, input)

	return item, erring.Wrap(err, "catalog.Update")
}
