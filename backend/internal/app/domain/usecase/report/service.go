package report

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
)

var ErrInvalidReport = errors.New("a IA retornou um relatório inválido")

type Service struct {
	Data       Data
	Repository Repository
	Generator  Generator
	Renderer   DocumentRenderer
}

func ValidateRequest(input Request) error {
	if len(strings.TrimSpace(input.Prompt)) < 5 || len(input.Prompt) > 4000 {
		return erring.ErrInvalidInput.WithMessage("descreva o relatório em 5 a 4000 bytes")
	}

	if input.DemandID != "" && !identifier.IsUUID(input.DemandID) {
		return erring.ErrInvalidInput
	}

	if input.PreviousReportID != "" && !identifier.IsUUID(input.PreviousReportID) {
		return erring.ErrInvalidInput
	}

	for _, date := range []string{input.From, input.To} {
		if date != "" {
			if _, err := time.Parse("2006-01-02", date); err != nil {
				return erring.ErrInvalidInput.WithMessage("data inválida")
			}
		}
	}

	if input.From != "" && input.To != "" && input.From > input.To {
		return erring.ErrInvalidInput.WithMessage("período inválido")
	}

	return nil
}

func (s *Service) Build(ctx context.Context, userID, reportID string, input Request, progress Progress) (Report, error) {
	if err := ValidateRequest(input); err != nil {
		return Report{}, err
	}

	if progress == nil {
		progress = func(string, string) {}
	}

	tools := NewTools(s.Data, input)
	prompt := input.Prompt

	if input.PreviousReportID != "" {
		previous, err := s.Repository.Get(ctx, userID, input.PreviousReportID)
		if err != nil {
			return Report{}, erring.Wrap(err, "report.Generate")
		}

		prompt = "Relatório anterior: " + previous.Content.Title + "\nResumo anterior (não substitui consultas atuais): " + previous.Content.Summary + "\nPedido atual: " + prompt
	}

	progress("coleta", "Consultando os dados do workspace")

	if _, err := tools.Call(ctx, "get_metrics", json.RawMessage(`{}`)); err != nil {
		return Report{}, err
	}

	if err := s.collectContext(ctx, tools, input); err != nil {
		return Report{}, err
	}

	seed, err := json.Marshal(tools.Snapshots)
	if err != nil {
		return Report{}, erring.Wrap(err, "report.Generate.seed")
	}

	content, err := s.Generator.Generate(ctx, "Pedido: "+prompt+"\nEscopo obrigatório: demanda="+input.DemandID+", de="+input.From+", até="+input.To+"\nDados consultados: "+string(seed), tools, progress)
	if err != nil {
		return Report{}, erring.Wrap(err, "report.Generate")
	}

	if generateErr := ValidateContent(content, tools.Sources); generateErr != nil {
		return Report{}, generateErr
	}

	snapshot, err := json.Marshal(tools.Snapshots)
	if err != nil {
		return Report{}, erring.Wrap(err, "report.Generate")
	}

	if s.Renderer == nil {
		return Report{}, fmt.Errorf("%w: renderizador HTML indisponível", ErrInvalidReport)
	}

	content.HTML, err = s.Renderer.Render(content)
	if err != nil {
		return Report{}, erring.Wrap(err, "report.Build.html")
	}

	result := Report{ID: reportID, UserID: userID, DemandID: input.DemandID, Prompt: input.Prompt, CreatedAt: time.Now().UTC(), Content: content, SourceData: snapshot}

	progress("revisão", "Relatório revisado; preparando o documento HTML")

	return result, nil
}

func ValidateContent(content Content, sources map[string]entity.Demand) error {
	invalid := func(reason string) error { return fmt.Errorf("%w: %s", ErrInvalidReport, reason) }
	if len(content.HTML) > 1024*1024 {
		return invalid("HTML acima do limite de 1 MiB")
	}

	if strings.TrimSpace(content.Title) == "" || strings.TrimSpace(content.Summary) == "" {
		return invalid("título ou resumo ausente")
	}

	if len(content.Metrics) > 30 || len(content.Charts) > 12 || len(content.Tables) > 12 {
		return invalid("quantidade de componentes acima do limite")
	}

	for _, id := range content.SourceIDs {
		if _, ok := sources[id]; !ok {
			return invalid("source_ids contém uma referência que não foi consultada")
		}
	}

	for _, chart := range content.Charts {
		if chart.Kind != "bar" && chart.Kind != "line" {
			return invalid("tipo de gráfico não suportado")
		}

		if len(chart.Points) > 120 {
			return invalid("gráfico acima do limite de 120 pontos")
		}
	}

	for _, table := range content.Tables {
		if len(table.Columns) == 0 || len(table.Columns) > 20 || len(table.Rows) > 500 {
			return invalid("tabela deve conter de 1 a 20 colunas e até 500 linhas")
		}

		for _, row := range table.Rows {
			if len(row) != len(table.Columns) {
				return invalid("linha da tabela possui quantidade de células diferente das colunas")
			}
		}
	}

	return nil
}

func (s *Service) collectContext(ctx context.Context, tools *Tools, input Request) error {
	if input.DemandID == "" {
		_, err := tools.Call(ctx, "list_demands", json.RawMessage(`{"status":""}`))

		return err
	}

	detail, err := s.Data.Detail(ctx, input.DemandID)
	if err != nil {
		return erring.Wrap(err, "report.collectContext")
	}

	if detail.Demand.Status != entity.Done {
		return erring.ErrConflict.WithMessage("o relatório individual exige uma demanda finalizada")
	}

	args, err := json.Marshal(map[string]string{"id": input.DemandID})
	if err != nil {
		return erring.Wrap(err, "report.collectContext.arguments")
	}

	_, err = tools.Call(ctx, "get_demand_context", args)

	return err
}

// Validate checks the request scope and ownership before an asynchronous job is queued.
func (s *Service) Validate(ctx context.Context, userID string, input Request) error {
	if err := ValidateRequest(input); err != nil {
		return err
	}

	if input.PreviousReportID != "" {
		if _, err := s.Get(ctx, userID, input.PreviousReportID); err != nil {
			return erring.Wrap(err, "report.Validate.previous")
		}
	}

	if input.DemandID != "" {
		detail, err := s.Data.Detail(ctx, input.DemandID)
		if err != nil {
			return erring.Wrap(err, "report.Validate.demand")
		}

		if detail.Demand.Status != entity.Done {
			return erring.ErrConflict.WithMessage("o relatório individual exige uma demanda finalizada")
		}
	}

	return nil
}

func (s *Service) List(ctx context.Context, userID, demandID string) ([]Report, error) {
	if demandID != "" && !identifier.IsUUID(demandID) {
		return nil, erring.ErrInvalidInput.WithMessage("demanda inválida")
	}

	items, err := s.Repository.List(ctx, userID, demandID)

	return items, erring.Wrap(err, "report.List")
}

func (s *Service) Get(ctx context.Context, userID, id string) (Report, error) {
	if !identifier.IsUUID(id) {
		return Report{}, erring.ErrInvalidInput.WithMessage("relatório inválido")
	}

	item, err := s.Repository.Get(ctx, userID, id)

	return item, erring.Wrap(err, "report.Get")
}
