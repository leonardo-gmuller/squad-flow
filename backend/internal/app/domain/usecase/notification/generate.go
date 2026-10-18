//go:generate go tool mockgen -write_command_comment=false -source=interface.go -destination=mocks/usecase_mock.go -package=mocks -mock_names UseCaseInterface=MockUseCase

//go:generate go tool mockgen -write_command_comment=false -source=notification.go -destination=repository_mock_test.go -package=notification

package notification
