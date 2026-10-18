package release

import (
	"context"
	"strings"
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
)

type Repository interface {
	List(ctx context.Context) ([]entity.Release, error)
	Insert(ctx context.Context, item entity.Release) error
	Update(ctx context.Context, id string, input Input, now time.Time) error
	SetArchived(ctx context.Context, id string, archived bool, now time.Time) error
}

type Input struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	ReleaseDate *time.Time `json:"release_date"`
}

type UseCase struct{ repo Repository }

func New(repo Repository) *UseCase { return &UseCase{repo: repo} }

func normalize(input Input) (Input, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)

	if input.Name == "" || len(input.Name) > 80 || len(input.Description) > 20000 {
		return input, erring.ErrInvalidInput.WithMessage("informe uma versão de até 80 caracteres e descrição de até 20000")
	}

	return input, nil
}

func (s *UseCase) List(ctx context.Context) ([]entity.Release, error) {
	items, err := s.repo.List(ctx)

	return items, erring.Wrap(err, "release.List")
}

func (s *UseCase) Create(ctx context.Context, input Input) (entity.Release, error) {
	input, err := normalize(input)
	if err != nil {
		return entity.Release{}, erring.Wrap(err, "release.Create")
	}

	id, err := identifier.NewUUID()
	if err != nil {
		return entity.Release{}, erring.Wrap(err, "release.Create")
	}

	now := time.Now().UTC()

	item := entity.Release{ID: id, Name: input.Name, Description: input.Description, ReleaseDate: input.ReleaseDate, CreatedAt: now, UpdatedAt: now}
	if err = s.repo.Insert(ctx, item); err != nil {
		return entity.Release{}, erring.Wrap(err, "release.Create")
	}

	return item, nil
}

func (s *UseCase) Update(ctx context.Context, id string, input Input) error {
	if !identifier.IsUUID(id) {
		return erring.ErrInvalidInput.WithMessage("versão inválida")
	}

	input, err := normalize(input)
	if err != nil {
		return err
	}

	return erring.Wrap(s.repo.Update(ctx, id, input, time.Now().UTC()), "release.Update")
}

func (s *UseCase) SetArchived(ctx context.Context, id string, archived bool) error {
	if !identifier.IsUUID(id) {
		return erring.ErrInvalidInput.WithMessage("versão inválida")
	}

	return erring.Wrap(s.repo.SetArchived(ctx, id, archived, time.Now().UTC()), "release.SetArchived")
}
