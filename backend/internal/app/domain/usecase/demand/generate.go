//go:generate go tool mockgen -write_command_comment=false -source=interface.go -destination=mocks/usecase_mock.go -package=mocks -mock_names UseCaseInterface=MockUseCase,BoardUseCaseInterface=MockBoardUseCase

//go:generate go tool mockgen -write_command_comment=false -source=contracts.go -destination=contracts_mock_test.go -package=demand

package demand
