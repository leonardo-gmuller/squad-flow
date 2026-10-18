//go:generate go tool mockgen -write_command_comment=false -source=interface.go -destination=mocks/usecase_mock.go -package=mocks -mock_names UseCaseInterface=MockUseCase,PasswordUseCaseInterface=MockPasswordUseCase

package auth
