# Testes do backend

O backend separa testes unitários de testes de integração pela build tag `integration`. Ambos usam `-race`. O Go continua na versão 1.24.

## Execução

Na raiz do repositório:

```bash
make test-unit         # todos os testes sem infraestrutura externa
make test-integration  # API, PostgreSQL/SQLC e Redis; requer Docker em execução
make test              # as duas suítes
make mocks-generate    # gera novamente os mocks
make mocks-check       # verifica a geração sem modificar os arquivos
make lint              # inclui os testes com a tag integration
```

Para executar um cenário específico:

```bash
cd backend
go test -race -run TestMoveEnforcesWorkflowAndCurrentQACycle ./internal/app/domain/usecase/demand
go test -race -tags=integration -count=1 -run TestPostgresQACycleWithRejectionAndFinalApproval ./internal/app/gateway/postgres
```

O comando de integração limita a execução a dois pacotes e dois testes paralelos por pacote para controlar o consumo do Docker. `-count=1` impede que resultados em cache substituam a execução contra os contêineres.

## Integração com Testcontainers

`internal/testinfra/containers.go` centraliza os módulos oficiais de PostgreSQL e Redis do Testcontainers. Cada teste recebe um contêiner próprio, iniciado com as mesmas versões principais do ambiente da aplicação: `postgres:17-alpine` e `redis:7.4-alpine`.

As portas são atribuídas dinamicamente e a conexão é obtida do contêiner. O PostgreSQL aguarda a inicialização completa e recebe as migrations reais da aplicação. O Testcontainers registra a limpeza no ciclo de vida do teste; pools, servidores HTTP e clientes são encerrados antes da remoção do contêiner.

Não há dependência de `TEST_DATABASE_URL`, `TEST_REDIS_URL`, Compose ou da `.env`. Sem Docker acessível, a suíte de integração falha; ela não ignora cenários silenciosamente. A primeira execução pode demorar mais pelo download das imagens, incluindo o Ryuk, responsável pela coleta de recursos do Testcontainers.

A suíte verifica rollback real da unidade de trabalho, rollback de OS quando a primeira demanda falha, fluxo completo de reprovação e aprovação em QA, migrations, versões, permissões, autenticação, reset de senha, notificações, relatórios assíncronos, outbox, leases e recuperação de consumidores Redis. Os testes da LLM usam respostas locais controladas, sem chave OpenAI ou consumo de créditos.

## Mocks gerados com Uber GoMock

A dependência é `go.uber.org/mock`, com `mockgen` declarado na seção `tool` do `go.mod`. Não é necessário instalar um gerador global: `go generate` usa a versão fixada pelo módulo.

As diretivas ficam em `generate.go`. Os mocks de repositórios usados somente no próprio pacote têm sufixo `_mock_test.go`. Os contratos dos casos de uso ficam em `interface.go`, no domínio, e geram `mocks/usecase_mock.go`: um pacote reutilizável que pode ser importado por testes de handlers, workers e outros consumidores. Os arquivos gerados devem ser versionados. Altere a interface e execute `make mocks-generate`; não edite o código gerado manualmente.

Os casos de uso de autenticação, catálogo, demandas, notificações, OS, versões e relatórios expõem `UseCaseInterface`. O quadro Kanban, as solicitações assíncronas e o worker de relatórios têm contratos próprios: `BoardUseCaseInterface`, `JobUseCaseInterface` e `WorkerInterface`. Cada implementação possui uma verificação de compatibilidade em compilação, como `var _ UseCaseInterface = (*UseCase)(nil)`.

Handlers e fábricas da aplicação recebem ou expõem os contratos do domínio. As interfaces duplicadas que ficavam nos handlers de demandas e OS foram removidas. Os consumidores de relatórios usam métodos como `Get`, `List`, `Validate` e `Build`; não acessam campos dos repositórios através do caso de uso.

Exemplo de expectativa:

```go
controller := gomock.NewController(t)
repository := NewMockQueryRepository(controller)
repository.EXPECT().Detail(t.Context(), demandID).Return(detail, nil)
service := NewUseCase(NewMockUnitOfWork(controller), repository)
```

`gomock.NewController(t)` verifica as expectativas automaticamente no encerramento do teste. Use argumentos exatos e `gomock.InOrder` quando a ordem fizer parte da regra. `gomock.Cond` permite conferir entidades que recebem UUIDs ou datas durante a execução. Sem uma expectativa, qualquer chamada ao mock falha: os cenários de entrada inválida ou acesso proibido verificam assim que a persistência não foi consultada.

Para testar um handler com um caso de uso simulado, importe o pacote `mocks` correspondente. Por exemplo, os testes de demandas usam `demandmock.NewMockUseCase(controller)` e passam o resultado para `handler.New`. Os testes de login e relatórios usam o mesmo padrão. Os testes de relatórios também verificam que o usuário autenticado chega ao contrato e que a solicitação retorna `202` sem chamar `Build` no processo HTTP.

Mocks de unidade de trabalho executam o callback com os repositórios gerados para testar as regras e a propagação de erros. A atomicidade é verificada pela suíte com PostgreSQL real.

## CI

O job de backend verifica SQLC, mocks gerados, lint, `go vet` e as duas suítes. O runner disponibiliza Docker; os testes criam e removem a infraestrutura necessária via Testcontainers. Não são usados serviços PostgreSQL ou Redis compartilhados no job.
