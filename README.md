# SquadFlow — Gestão do desenvolvimento de software

MVP full-stack para organizar o desenvolvimento de software em **esteiras cadastradas pelo administrador**, com Melhorias, Corretivas, Testes e Reescritas como catálogo inicial, incluindo OS grandes distribuídas entre vários desenvolvedores, OS executadas por uma única pessoa e tickets corretivos independentes. Projeto **demonstrativo/portfólio**, com dados totalmente fictícios, sem código ou dados internos da LG.

## Como o domínio está modelado

- **Iniciativa:** nome livre de agrupamento, como uma NT, projeto de modernização ou melhoria rotineira. Não existe obrigatoriedade de vincular a uma NT.
- **OS / work package:** registro agrupador com número externo, título, descrição e iniciativa; pode ter **1 ou N entregas**, inclusive com responsáveis distintos. É um contêiner, não uma tarefa de desenvolvimento.
- **Entrega / work item:** unidade executável com responsável, QA, prioridade, produto, prazo, critérios de aceite, eventos, status, testes, devoluções e histórico. Pode ter `package_id` ou ser um **ticket avulso** (corretiva sem OS).
- **Métricas:** contam somente entregas, não OS + entregas; progresso da OS = itens concluídos / itens registrados. Observe que uma OS vazia não representa entrega concluída.

Veja [docs/DOMAIN-MODEL.md](docs/DOMAIN-MODEL.md) para o domínio e [docs/FRONTEND-SPA.md](docs/FRONTEND-SPA.md) para as decisões de interface.

## Stack

- **Go 1.24**, Clean Architecture / Hexagonal, monólito modular com `domain/entity`, `domain/usecase`, gateways e composição em `app.go`.
- **UoW (pgx/v5):** registry de repositórios, cada operação com `TxScope` privado, rollback em erro, transações sem estado global compartilhado.
- **SQLC v1.27.0:** SQL em `gateway/postgres/queries`, código gerado em `gateway/postgres/sqlc` e implementações em `gateway/postgres/repository`, com interfaces nos casos de uso. Veja [persistência e geração](docs/SQLC-REPOSITORIES.md).
- **UUIDs:** chaves PostgreSQL `uuid`, geração e validação com `github.com/google/uuid` em `pkg/identifier`. Os campos `id` da API contêm UUIDs; `SF-0001` é apenas o código legível da demanda.
- **API Chi:** middlewares próprios em `gateway/api/middleware`, `RequestID`, `Recoverer` e `NoCache` do Chi e library `pkg/ctxkey` para contexto tipado. Veja [middlewares e contexto](docs/API-MIDDLEWARES.md).
- **PostgreSQL 17:** migrations com versionamento; tabelas `work_packages`, `work_items`, `test_cases`, `work_item_transitions`.
- **Vue 3 + TypeScript + Vite**, **Vue Router (SPA)**, **Pinia**, **PrimeVue 4 (Aura)** e **Tailwind CSS 4**, com componentes Lucide.
- **Relatórios assíncronos:** eventos/outbox em PostgreSQL, gateway Redis Streams, worker com retomada e tentativas; documentos HTML independentes para dashboard, apresentação e download. Veja [arquitetura e execução](docs/ASYNC-HTML-REPORTS.md).
- **Docker Compose:** banco, Redis, API, worker e frontend com Nginx. **Testes:** Go unitários (inclui pacote/OS transacional), integração PostgreSQL sob demanda e Playwright E2E. GitHub Actions já configurado.

## Observabilidade

OpenTelemetry exporta traces, métricas e logs por OTLP HTTP/protobuf, com correlação entre API, PostgreSQL, outbox, Redis, worker e chamadas da LLM. A exportação é opcional (`OTEL_ENABLED=false` por padrão). `make observability-up` inicia o ambiente completo com Collector e Grafana local em `http://127.0.0.1:3000`, com um dashboard provisionado; `make observability-local` sobe apenas a stack para processos Go locais. Veja [OBSERVABILITY.md](docs/OBSERVABILITY.md) para configuração, consultas e limites.

## Rodar localmente

É necessário ter Docker Engine ou Docker Desktop com Compose v2 e acesso à internet para baixar imagens e dependências.

```bash
cp .env.example .env
# Configure JWT_SECRET, ADMIN_EMAIL e ADMIN_PASSWORD antes de iniciar.
# JWT_SECRET: gere com openssl rand -hex 32. Senha do admin: mínimo de 12 caracteres, máximo de 72 bytes.
# Para gerar relatórios com IA, configure também OPENAI_API_KEY.
docker compose up -d --build
# http://127.0.0.1:8080
```

Apenas o frontend é publicado no loopback `127.0.0.1:8080`. API e PostgreSQL são internos à rede do Compose.

```bash
docker compose logs -f api
docker compose down     # mantém os dados
docker compose down -v  # apaga o banco local!
```

Com `DEMO_SEED=true`, serão cadastradas cinco contas fictícias com responsáveis vinculados, demandas fictícias, **uma OS com dois responsáveis**, uma OS individual e um ticket corretivo avulso. Use `DEMO_SEED=false` para iniciar sem usuários/demandas fictícios. O catálogo inicial de esteiras continua disponível. Veja contas e funcionamento em [Esteiras e responsáveis](docs/WORKSTREAMS-AND-MEMBERS.md).

Usuários membros têm uma **central de notificações** no botão de sino do topo, com atribuições, mudanças de etapa, QA e controle de leitura. O admin não recebe nem acessa a central. O indicador “Ao vivo” foi substituído; a conexão em tempo real continua funcionando. Veja [Notificações](docs/NOTIFICATIONS.md).

## Frontend SPA — arquitetura e telas

O frontend usa componentes e rotas protegidas por sessão. A SPA funciona com deep links por módulo graças ao fallback de Nginx para `index.html`.

| Rota | Tela |
|---|---|
| `/dashboard` | Indicadores, esteiras, andamento e entregas recentes |
| `/tickets` | Lista paginada e ordenável com filtros |
| `/workstreams` | Cadastro, edição e desativação de esteiras pelo admin |
| `/board` | Kanban paginado por etapa, com filtros de versão, arquivo, esteira e responsável |
| `/releases` | Versões de entrega, conteúdo por versão, arquivamento e restauração |
| `/packages` | Visão mestre/detalhe de OS e suas entregas |
| `/qa` | Demandas em QA, devolvidas e concluídas |
| `/reports` | Métricas, volumes por responsável, progresso por OS e CSV |
| `/login` | Autenticação por cookies HttpOnly |
| `/users` | Gestão de usuários, senhas e links internos pelo administrador |
| `/account` | Troca da própria senha com confirmação da senha atual |
| `/reset-password` | Redefinição por link interno entregue pelo admin, sem e-mail |

Componentes globais de criação e edição de entregas usam `Dialog` e `Drawer` do PrimeVue. Os dados carregam de `stores/workspace.ts`; `stores/catalog.ts` carrega as esteiras e o diretório de usuários da API; `config/catalog.ts` concentra os rótulos de status/prioridade e consulta esse catálogo. Desenvolvedor e QA usam seletores de usuários cadastrados, e a API valida os vínculos por ID. O visual é inspirado no [TaskWhiz](https://dribbble.com/shots/20769637-TaskWhiz-Task-Management-Kanban-Dashboard), com navegação lateral compacta, superfícies em carvão e destaque violeta. O tema escuro é o padrão inicial; a escolha de tema claro/escuro fica salva como preferência visual local.

Os formulários usam controles PrimeVue (`InputText`, `Textarea`, `Select`, `DatePicker`, `Password`, `Checkbox` e `Message`), com grids e espaçamento em Tailwind. `config/primevue.ts` centraliza o tema Aura, a tradução e a ordem das camadas CSS; `FormField`, `CalendarDateInput` e `PasswordInput` compartilham rótulos, datas e senhas. Moment com locale `pt-br`, centralizado em `config/moment.ts`, formata as datas e horários da interface. Os calendários exibem `dd/mm/aaaa` e preservam o dia ao enviar dados à API. Nos gráficos, eixos e tooltips mostram meses e dias em português, inclusive ao abrir relatórios com datas ISO ou rótulos em inglês; as séries diárias mantêm os dias UTC calculados pela API. Tabelas usam `DataTable` e a Central de QA usa `Tabs`. Todos os gráficos usam `Chart` do PrimeVue com Chart.js: barras por esteira, rosca de progresso, linhas de entregas diárias e gráficos dos relatórios de IA. Os gráficos acompanham o tema claro/escuro e oferecem descrições acessíveis; as barras de progresso das OS usam `ProgressBar`.

**Autenticação, IA e tempo real:** JWT e refresh token com rotação em cookies HttpOnly, um único administrador, workspace compartilhado, três agentes de relatórios usando OpenAI Responses, endpoint MCP autenticado e WebSocket. Os indicadores fixos são calculados na API Go. Veja [configuração e arquitetura](docs/AUTH-AI-REPORTS.md) para variáveis, contratos, limites e testes.

## Funcionalidades da interface

- **Visão geral:** métricas, distribuição por esteira, andamento e entregas recentes.
- **Senhas e recuperação:** admin altera senhas ou gera links internos de uso único, válidos por 30 minutos; usuários trocam a própria senha em Minha conta. Cada troca encerra as sessões e invalida links pendentes. Veja [o fluxo completo](docs/PASSWORD-RECOVERY.md).
- **Versões de código:** cadastro com nome livre, descrição e data de entrega opcional; vínculo opcional por demanda, inclusive a primeira entrega de uma OS; consulta das entregas e dos totais por versão. Uma OS pode ter entregas em versões diferentes.
- **Quadro Kanban:** cinco etapas, cartões arrastáveis entre etapas permitidas, com etiquetas, responsáveis e prazos. O arraste destaca destinos, rola nas bordas e pode ser cancelado com Escape; a alça também permite mover por toque ou teclado. Devoluções pedem motivo e conclusões respeitam a aprovação de QA, com histórico e autor registrados na API. Busca e versão ficam sempre visíveis; o botão **Filtros** expande arquivo, esteira e responsável. Carregamento de 20 cartões por coluna com contagem total e botão para carregar mais. A consulta ocorre no servidor e a tela não baixa a lista completa de demandas.
- **Arquivo de versões:** retira suas demandas da visão ativa do Kanban, preserva estados, QA, histórico e métricas, e permite restauração. O filtro de arquivadas mostra o histórico. Novas vinculações a versões arquivadas são recusadas; demandas sem versão seguem visíveis no quadro ativo.
- **OS e iniciativas:** cards de OS, contadores e progresso; possibilidade de criar OS com primeira entrega na mesma transação (ou sem ela), adicionar itens por pessoa e abrir um item para ver a homologação.
- **Nova entrega:** pode ser avulsa ou vinculada a uma OS; campos de produto, eventos, critérios de aceite e prazo. Uma entrega existente pode ser editada, sem reiniciar seu ciclo de QA.
- **QA:** casos aprovados/reprovados, reprovação devolve automaticamente ao desenvolvimento na mesma transação. Aprovação final exige pelo menos um teste aprovado e nenhum reprovado **no ciclo atual**.
- **Relatórios:** throughput diário, lead time médio/mediano/P85, primeiro passe, retrabalho, prazos, trabalho aberto, distribuição por responsável/OS e CSV. Pedidos livres à IA entram em uma fila persistente e geram documentos HTML com contexto, testes e movimentações de demandas finalizadas, que podem ser abertos no dashboard, apresentados e baixados. O trabalho continua ao sair da tela.

## REST API

| Método | Rota | Uso |
|---|---|---|
| GET | `/healthz` | Health check |
| GET, POST | `/api/v1/packages` | Consultar / criar OS |
| GET | `/api/v1/packages/{id}` | Detalhes resumidos da OS |
| GET, POST | `/api/v1/demands` | Listar / cadastrar entregas |
| GET, PUT | `/api/v1/demands/{id}` | Consultar / editar entrega |
| POST | `/api/v1/demands/{id}/move` | Transição de status |
| POST | `/api/v1/demands/{id}/tests` | Registrar teste QA |
| GET | `/api/v1/metrics` | Indicadores consolidados |
| GET | `/api/v1/board` | Colunas paginadas do Kanban |
| GET, POST | `/api/v1/releases` | Listar / cadastrar versões |
| PUT | `/api/v1/releases/{id}` | Editar versão |
| POST | `/api/v1/releases/{id}/archive` | Arquivar / restaurar com `archived: true/false` |
| GET | `/api/v1/reports/demands.csv` | Relatório CSV |

Todas as rotas de dados exigem sessão. A migração `004_releases.sql` é aplicada ao iniciar a API; demandas existentes continuam sem versão até serem vinculadas. Veja [versões e Kanban](docs/RELEASES.md) para regras, filtros e paginação. Exemplo usando um cookie jar obtido após login:

```bash
curl -X POST http://127.0.0.1:8080/api/v1/packages \
  -b cookies.txt \
  -H 'Content-Type: application/json' \
  -H 'X-SquadFlow-Client: web' \
  -d '{"reference":"OS-9001","title":"Otimizar validações","initiative":"Evolução da plataforma","kind":"improvement","create_first_item":true,"assignee":"Alex"}'
```

A resposta inclui o UUID `id` da OS. Cadastre outros itens através da interface "Novo item" ou POST `/api/v1/demands` com `package_id` igual ao UUID da OS, incluindo título, esteira e prioridade.

## Padrão de erros e validação de login

A API usa erros de aplicação com códigos estáveis e mensagens públicas em português, por exemplo `auth:invalid-credentials` para email ou senha incorretos. A biblioteca `gateway/api/response` centraliza status HTTP, payloads de erro, respostas de sucesso, cookies e headers. Os erros REST seguem `{ type, code, message }`; detalhes internos ficam nos logs. O login usa PrimeVue Forms com resolver Zod e erros por campo. Veja [contratos e exemplos](docs/API-ERRORS.md).

## Lint e formatação do frontend

O frontend usa ESLint com regras recomendadas para Vue 3 e TypeScript, e Prettier para padronizar Vue, TypeScript, CSS, JSON e testes. O lint exige zero avisos; regras de formatação conflitantes são desativadas com `eslint-config-prettier`, conforme a [integração recomendada pelo Prettier](https://prettier.io/docs/integrating-with-linters.html). Use Node `20.19+`, `22.13+` ou `24+` nas respectivas linhas suportadas.

```bash
cd frontend
npm ci
npm run lint          # verifica código e templates
npm run lint:fix      # aplica correções automáticas do ESLint
npm run format       # formata os arquivos com Prettier
npm run format:check # verifica formatação sem modificar arquivos
npm run check        # lint + formatação + TypeScript
npm run build        # valida TypeScript e gera a SPA
```

Na raiz: `make lint-front`, `make lint-front-fix`, `make format-front`, `make format-front-check` e `make check-front`. O CI verifica lint e formatação antes do build. As configurações em `.vscode` recomendam as extensões Vue, ESLint e Prettier e habilitam formatação ao salvar nos arquivos do frontend. Arquivos gerados, resultados de testes e o lockfile ficam fora da formatação.

## Testes

Após alterar queries ou migrations, regenere os tipos com `make sqlc-generate`. Execute `make sqlc-check` para validar o SQL e verificar se os arquivos gerados estão atualizados. Os comandos instalam a versão fixada do SQLC em `.tools/bin`, se necessário; o CI executa a mesma verificação.

O padrão de lint do backend está em [`.golangci.yml`](.golangci.yml). Os comandos abaixo partem da raiz e analisam o módulo `backend`, incluindo os testes:

```bash
make lint          # instala a versão fixada da ferramenta, se necessário, e verifica
make lint-fix      # aplica as correções automáticas disponíveis
make lint-install  # reinstala o golangci-lint v1.64.8 em .tools/bin
```

A mesma versão roda no CI. A configuração mantém o formato v1 do modelo, com os imports locais do SquadFlow, `shadow` no `govet` e as opções renomeadas atualizadas. `mnd` substitui `gomnd`; `usetesting` substitui `tenv`. As palavras do produto em português ficam em `misspell.ignore-words`. O teste que captura o logger global roda em série, com uma justificativa local para `paralleltest`.

```bash
# backend — a partir da raiz
make test-unit         # sem Docker
make test-integration  # PostgreSQL e Redis descartáveis via Testcontainers; requer Docker
make test              # ambas as suítes, com detector de corrida
make mocks-generate    # gera mocks com go.uber.org/mock/mockgen fixado no go.mod
make mocks-check       # verifica se os mocks estão atualizados

# frontend — com a aplicação em localhost:8080
cd frontend
npm install
npx playwright install chromium
E2E_EMAIL=admin-de-teste@example.com E2E_PASSWORD=senha-de-teste npm run test:e2e
```

Os E2E foram atualizados para a navegação SPA e componentes PrimeVue: fluxo de reprovação/reenvio/aprovação, cadastro de OS, item adicional, corretiva avulsa e navegação por rotas. Os testes de integração usam a tag `integration` e criam seus próprios contêineres com portas dinâmicas e limpeza automática. Não precisam de banco configurado na `.env`; sem Docker disponível, falham explicitamente. Veja [TESTING.md](docs/TESTING.md) para o padrão de mocks e execução.

## Configuração e limites atuais

Login, papéis admin/member, identidade nas movimentações e testes, OpenAI, MCP e WebSocket estão implementados. O registro de edições com histórico próprio, integrações Redmine/SVN e importação de NT continuam fora do escopo. O WebSocket distribui eventos de uma única instância. Veja [AUTH-AI-REPORTS.md](docs/AUTH-AI-REPORTS.md) para configurar HTTPS/cookies, origens, credenciais e limites da geração.

### Validação nesta implementação

Passaram o build Vue com checagem TypeScript, lint Go, testes Go com `-race` e PostgreSQL real descartável, além de 40 cenários Playwright: autenticação, senhas, permissões, navegação, calendários, datas em português, gráficos, OS, fluxo QA, relatórios, versões, paginação e arraste no Kanban por mouse, teclado e toque. As telas de relatórios também foram verificadas em desktop/mobile. Os testes OpenAI usam respostas simuladas; uma chamada real requer sua chave no backend.

## Desenvolvimento local do frontend

```bash
cd frontend
npm install
npm run dev     # http://localhost:5173 com proxy para http://localhost:8090
npm run build   # valida TypeScript e gera a SPA
```

O backend precisa estar ativo na porta 8090 durante o desenvolvimento. Em Docker Compose o frontend é servido por Nginx em `http://127.0.0.1:8080`.

A tipografia usa Google Fonts (DM Sans e Manrope) com fallback para fontes do sistema. Para distribuição em ambiente corporativo, remova ou hospede essas fontes internamente conforme as políticas de rede e segurança.
