# Persistência com SQLC

As interfaces de persistência pertencem aos casos de uso, em `backend/internal/app/domain/usecase`. As implementações PostgreSQL ficam em `backend/internal/app/gateway/postgres/repository`, separadas por módulo: autenticação, catálogo, demandas, notificações, OS, versões, relatórios, testes e transições.

## Organização

```text
backend/
  sqlc.yaml
  internal/app/
    domain/usecase/demand/contracts.go  # interfaces consumidas pela aplicação
    gateway/postgres/
      migrations/                      # schema e evolução do banco
      queries/demand.sql               # SQL nomeado, fonte de verdade
      sqlc/                            # código gerado, não editar manualmente
      repository/demand/
        demand_repository.go           # struct, constructor e contratos
        demand_create.go
        demand_get.go
        demand_update.go
        demand_mapper.go               # conversões SQLC <-> domínio
        demand_validation.go
        board.go
```

Cada implementação embute `*sqlc.Queries` e recebe `uow.DBTX`, permitindo usar tanto o pool quanto uma transação. As verificações de interface em tempo de compilação garantem a implementação dos contratos. Os casos de uso continuam recebendo interfaces e entidades; não importam SQLC nem pgx.

```go
type DemandRepository struct{ *sqlc.Queries }

var _ demand.Repository = (*DemandRepository)(nil)

func NewDemandRepository(db uow.DBTX) *DemandRepository {
    return &DemandRepository{Queries: sqlc.New(db)}
}
```

Os arquivos de operação chamam os métodos gerados com parâmetros tipados. Os mappers traduzem resultados para entidades e entidades para parâmetros de escrita. Erros do PostgreSQL são convertidos em erros de aplicação nas implementações. Listas vazias continuam sendo serializadas como `[]`.

## Gerar e verificar

Os comandos partem da raiz do projeto e usam SQLC **v1.27.0**:

```bash
make sqlc-install   # instala em .tools/bin; geração/check também instalam se necessário
make sqlc-generate  # após alterar queries ou schema
make sqlc-check     # compile, vet e diff dos arquivos gerados
make lint
cd backend && go test ./...
```

Para usar uma instalação existente da mesma versão: `make sqlc-check SQLC=/caminho/para/sqlc`. Os arquivos gerados devem acompanhar as queries no controle de versão. O CI usa `sqlc diff` para rejeitar código gerado desatualizado. O build da API usa os arquivos gerados incluídos no projeto, sem executar SQLC em produção.

A instalação usa `CGO_ENABLED=0`, selecionando o parser PostgreSQL em WebAssembly incluído no SQLC e evitando dependência do compilador C/SDK local.

O [arquivo de configuração](../backend/sqlc.yaml) lê o schema diretamente das migrations e gera código para pgx/v5. UUIDs são representados por `string`, UUIDs opcionais por `*string`, inteiros PostgreSQL `int4` por `int` e timestamps por `time.Time` ou `*time.Time`. Essas conversões ficam restritas à persistência.

## Identificadores públicos

As chaves dos recursos já são colunas PostgreSQL `uuid`. Os campos `id`, `demand_id`, `package_id`, `release_id` e os demais vínculos da API contêm UUIDs no formato de 36 caracteres; não são IDs numéricos internos. O código `SF-0001` é uma referência legível da demanda, independente da chave usada nas URLs e nos vínculos.

A geração e a validação ficam em `pkg/identifier`, usando `github.com/google/uuid`: `NewUUID` gera UUID v4 com tratamento de erro, e `IsUUID` valida o formato do contrato existente. Os nomes dos campos JSON e os UUIDs dos registros existentes permanecem os mesmos. O frontend recebe UUIDs como strings, como exige a representação JSON.

Consultas e alterações recusam identificadores malformados antes da persistência. O contrato aceita letras hexadecimais maiúsculas e minúsculas, mas não aceita formas alternativas como URN, chaves ou UUID sem hífens. Os tokens de sessão também exigem UUIDs nos identificadores de usuário e sessão. Nenhuma migração de dados é necessária.

## Transações e comportamento preservado

- A UoW cria os repositories sobre o mesmo `pgx.Tx`. Demanda, histórico, QA e notificações confirmam ou revertem juntos.
- Helpers de validação e eventos recebem as queries já vinculadas à transação. Os locks de usuários, esteiras e versões continuam no SQL.
- Autenticação controla transações de bootstrap, sessão, rotação de refresh token e senha; cada uma usa `Queries.WithTx(tx)`. Troca de senha, revogação de sessões, consumo dos links e auditoria são atômicos.
- Kanban mantém filtros, contagens e paginação por coluna numa única consulta/snapshot; notificações consultam página e total de não lidas no mesmo snapshot.
- Relatórios continuam limitados ao proprietário. Notificações continuam exclusivas dos membros.
- Seeds também usam queries geradas. O executor de migrations permanece separado: executa DDL dos arquivos SQL e controla `schema_migrations` antes de existir o schema da aplicação.

Para validar integração e concorrência, execute `make test-integration` na raiz com o Docker em execução. Os testes usam PostgreSQL descartável via Testcontainers, sem URLs externas ou testes ignorados por falta de configuração. Os testes unitários usam mocks gerados com `go.uber.org/mock`; veja [TESTING.md](TESTING.md).

Referências oficiais: [configuração e tipos](https://docs.sqlc.dev/en/v1.27.0/reference/config.html), [overrides](https://docs.sqlc.dev/en/v1.27.0/howto/overrides.html) e [transações](https://docs.sqlc.dev/en/v1.27.0/howto/transactions.html).
