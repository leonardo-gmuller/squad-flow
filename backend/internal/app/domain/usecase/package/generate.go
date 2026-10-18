//go:generate go tool mockgen -write_command_comment=false -source=interface.go -destination=mocks/usecase_mock.go -package=mocks -mock_names UseCaseInterface=MockUseCase

//go:generate go tool mockgen -write_command_comment=false -source=contracts.go -destination=contracts_mock_test.go -package=workpackage
//go:generate go tool mockgen -write_command_comment=false -source=../../../domain/usecase/demand/contracts.go -destination=demand_mock_test.go -package=workpackage -mock_names Repository=MockDemandRepository,TestRepository=MockDemandTestRepository,TransitionRepository=MockDemandTransitionRepository,Repositories=MockDemandRepositories,UnitOfWork=MockDemandUnitOfWork,QueryRepository=MockDemandQueryRepository

package workpackage
