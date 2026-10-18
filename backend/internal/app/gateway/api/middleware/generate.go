//go:generate go tool mockgen -write_command_comment=false -source=auth.go -destination=auth_mock_test.go -package=middleware_test
//go:generate go tool mockgen -write_command_comment=false -source=workspace_updates.go -destination=workspace_mock_test.go -package=middleware_test

package middleware
