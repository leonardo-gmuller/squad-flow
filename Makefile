GOLANGCI_LINT_VERSION := 1.64.8
GOLANGCI_LINT := $(CURDIR)/.tools/bin/golangci-lint
SQLC_VERSION := 1.27.0
SQLC ?= $(CURDIR)/.tools/bin/sqlc

.PHONY: up down logs test test-unit test-integration test-e2e mocks-generate mocks-check rebuild lint lint-fix lint-install lint-front lint-front-fix format-front format-front-check check-front sqlc-install sqlc-ready sqlc-generate sqlc-check
up:
	docker compose up -d --build
logs:
	docker compose logs -f --tail=100
down:
	docker compose down
rebuild:
	docker compose up -d --build --force-recreate
test: test-unit test-integration
test-unit:
	cd backend && go test -race ./...
test-integration:
	cd backend && go test -race -tags=integration -count=1 -p=2 -parallel=2 ./internal/app/gateway/api ./internal/app/gateway/postgres ./internal/app/gateway/queue/redis
mocks-generate:
	cd backend && go generate ./...
mocks-check:
	cd backend && bash scripts/check-mocks.sh
test-e2e:
	cd frontend && npm run test:e2e
lint-install:
	cd backend && GOBIN="$(CURDIR)/.tools/bin" GOTOOLCHAIN=go1.24.0 go install github.com/golangci/golangci-lint/cmd/golangci-lint@v$(GOLANGCI_LINT_VERSION)
lint:
	@test -x "$(GOLANGCI_LINT)" || $(MAKE) lint-install
	cd backend && "$(GOLANGCI_LINT)" run --config ../.golangci.yml --build-tags integration ./...
lint-fix:
	@test -x "$(GOLANGCI_LINT)" || $(MAKE) lint-install
	cd backend && "$(GOLANGCI_LINT)" run --config ../.golangci.yml --build-tags integration --fix ./...
lint-front:
	cd frontend && npm run lint
lint-front-fix:
	cd frontend && npm run lint:fix
format-front:
	cd frontend && npm run format
format-front-check:
	cd frontend && npm run format:check
check-front:
	cd frontend && npm run check
sqlc-install:
	cd backend && CGO_ENABLED=0 GOBIN="$(CURDIR)/.tools/bin" GOTOOLCHAIN=go1.24.0 go install github.com/sqlc-dev/sqlc/cmd/sqlc@v$(SQLC_VERSION)
sqlc-ready:
	@test -x "$(SQLC)" || $(MAKE) sqlc-install
	@test "$$('$(SQLC)' version)" = "v$(SQLC_VERSION)" || { echo "Use SQLC v$(SQLC_VERSION): make sqlc-install"; exit 1; }
sqlc-generate: sqlc-ready
	cd backend && "$(SQLC)" generate
sqlc-check: sqlc-ready
	cd backend && "$(SQLC)" compile
	cd backend && "$(SQLC)" vet
	cd backend && "$(SQLC)" diff
