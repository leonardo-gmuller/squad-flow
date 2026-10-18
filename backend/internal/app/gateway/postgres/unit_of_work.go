package postgres

import (
	"context"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	demand "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/demand"
	workpackage "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/package"
	demandrepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/demand"
	packagerepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/package"
	testcaserepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/testcase"
	transitionrepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/transition"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/uow"
)

const (
	RepositoryName           = "demands"
	PackageRepositoryName    = "packages"
	TestRepositoryName       = "test_cases"
	TransitionRepositoryName = "transitions"
)

type UnitOfWork struct{ engine *uow.PGX }

type txRepositories struct {
	demands     demand.Repository
	tests       demand.TestRepository
	transitions demand.TransitionRepository
}

func (r *txRepositories) Demands() demand.Repository               { return r.demands }
func (r *txRepositories) Tests() demand.TestRepository             { return r.tests }
func (r *txRepositories) Transitions() demand.TransitionRepository { return r.transitions }

func NewUnitOfWork(client *Client) (*UnitOfWork, error) {
	engine := uow.NewPGX(client.Pool)
	for _, reg := range []struct {
		name    string
		factory uow.RepoFactory
	}{
		{RepositoryName, func(db uow.DBTX) any { return demandrepo.NewDemandRepository(db) }},
		{PackageRepositoryName, func(db uow.DBTX) any { return packagerepo.NewPackageRepository(db) }},
		{TestRepositoryName, func(db uow.DBTX) any { return testcaserepo.NewTestCaseRepository(db) }},
		{TransitionRepositoryName, func(db uow.DBTX) any { return transitionrepo.NewTransitionRepository(db) }},
	} {
		if err := engine.Register(reg.name, reg.factory); err != nil {
			return nil, erring.Wrap(err, "postgres.NewUnitOfWork")
		}
	}

	return &UnitOfWork{engine: engine}, nil
}

func (u *UnitOfWork) Do(ctx context.Context, fn func(context.Context, demand.Repositories) error) error {
	return erring.Wrap(u.engine.Do(ctx, func(ctx context.Context, scope *uow.TxScope) error {
		demands, err := uow.RepoTx[demand.Repository](scope, RepositoryName)
		if err != nil {
			return err
		}

		tests, err := uow.RepoTx[demand.TestRepository](scope, TestRepositoryName)
		if err != nil {
			return err
		}

		transitions, err := uow.RepoTx[demand.TransitionRepository](scope, TransitionRepositoryName)
		if err != nil {
			return err
		}

		return fn(ctx, &txRepositories{demands: demands, tests: tests, transitions: transitions})
	}), "postgres.Do")
}

// Package writes use the very same transaction engine as demand transitions.
type packageTxRepos struct {
	packages workpackage.Repository
	demands  demand.Repository
}

func (r *packageTxRepos) Packages() workpackage.Repository { return r.packages }
func (r *packageTxRepos) Demands() demand.Repository       { return r.demands }
func (u *UnitOfWork) DoPackage(ctx context.Context, fn func(context.Context, workpackage.Repositories) error) error {
	return erring.Wrap(u.engine.Do(ctx, func(ctx context.Context, scope *uow.TxScope) error {
		packages, err := uow.RepoTx[workpackage.Repository](scope, PackageRepositoryName)
		if err != nil {
			return err
		}

		demands, err := uow.RepoTx[demand.Repository](scope, RepositoryName)
		if err != nil {
			return err
		}

		return fn(ctx, &packageTxRepos{packages: packages, demands: demands})
	}), "postgres.DoPackage")
}
