package app

import (
	"context"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/config"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/auth"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/catalog"
	demand "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/demand"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/notification"
	workpackage "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/package"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/release"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/openai"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres"
	authrepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/auth"
	catalogrepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/catalog"
	demandrepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/demand"
	notificationrepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/notification"
	packagerepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/package"
	releaserepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/release"
	reportrepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/report"
	reportjobrepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/reportjob"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/reporthtml"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/security"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
)

type App struct {
	DB         *postgres.Client
	UnitOfWork *postgres.UnitOfWork
	Auth       auth.UseCaseInterface
	Reports    *report.Service
	ReportJobs *report.JobService
}

func (a *App) Configure(ctx context.Context, cfg config.Config) error {
	authRepo := authrepo.NewAuthRepository(a.DB.Pool)

	a.Auth = auth.New(authRepo, authRepo, security.Passwords{}, security.Tokens{Secret: []byte(cfg.Auth.JWTSecret)}, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL)
	if err := a.Auth.Bootstrap(ctx, cfg.Auth.AdminName, cfg.Auth.AdminEmail, cfg.Auth.AdminPassword); err != nil {
		return erring.Wrap(err, "app.Configure")
	}

	llm := openai.New(cfg.OpenAI.APIKey, cfg.OpenAI.Model)
	llm.StoreResponses = cfg.OpenAI.StoreResponses

	a.Reports = &report.Service{Data: a.NewDemandUseCase(), Repository: reportrepo.NewReportRepository(a.DB.Pool), Generator: llm, Renderer: reporthtml.Renderer{}}
	a.ReportJobs = &report.JobService{Repository: reportjobrepo.NewReportJobRepository(a.DB.Pool), Reports: a.Reports}

	return nil
}

func New(db *postgres.Client) (*App, error) {
	unitOfWork, err := postgres.NewUnitOfWork(db)
	if err != nil {
		return nil, erring.Wrap(err, "app.New")
	}

	return &App{DB: db, UnitOfWork: unitOfWork}, nil
}

func (a *App) NewDemandUseCase() demand.UseCaseInterface {
	return demand.NewUseCase(a.UnitOfWork, demandrepo.NewDemandRepository(a.DB.Pool))
}

func (a *App) NewPackageUseCase() workpackage.UseCaseInterface {
	return workpackage.New(a.UnitOfWork, packagerepo.NewPackageRepository(a.DB.Pool), identifier.NewUUID)
}

func (a *App) NewReleaseUseCase() release.UseCaseInterface {
	return release.New(releaserepo.NewReleaseRepository(a.DB.Pool))
}

func (a *App) NewBoardUseCase() demand.BoardUseCaseInterface {
	return demand.NewBoardUseCase(demandrepo.NewDemandRepository(a.DB.Pool))
}

func (a *App) NewCatalogUseCase() catalog.UseCaseInterface {
	return catalog.New(catalogrepo.NewCatalogRepository(a.DB.Pool))
}

func (a *App) NewNotificationUseCase() notification.UseCaseInterface {
	return notification.New(notificationrepo.NewNotificationRepository(a.DB.Pool))
}
