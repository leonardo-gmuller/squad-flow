//go:generate go tool mockgen -write_command_comment=false -source=../../domain/usecase/report/contracts.go -destination=report_mock_test.go -package=openai

package openai
