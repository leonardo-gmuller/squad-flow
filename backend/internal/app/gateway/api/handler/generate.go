//go:generate go tool mockgen -write_command_comment=false -source=../../../domain/usecase/auth/auth.go -destination=auth_mock_test.go -package=handler

package handler
