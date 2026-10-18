# Changelog

## Atualização — glass e profundidade visual

- Fundo com luzes difusas e gradientes suaves, adaptado aos temas claro e escuro.
- Glass discreto na navegação e no login; painéis e colunas com transparência leve e reflexos nas bordas.
- Cards com transparência mais evidente, reflexo diagonal, bordas luminosas e desfoque do fundo, inclusive sem hover.
- Campos com base firme para preservar leitura, blur reduzido no celular e alternativa sem transparência.

## Atualização — cards arrastáveis no Kanban

- Movimentação por mouse, toque na alça ou seleção de etapa por teclado.
- Destinos permitidos destacados, prévia do cartão, rolagem nas bordas e cancelamento por Escape.
- Devolução com motivo obrigatório e conclusão validada pelo QA, mantendo autor e histórico na API.
- Atualização após confirmação do servidor e testes de interação e persistência com PostgreSQL.

## Atualização — Heroicons

- Heroicons em menus, ações, indicadores e campos, substituindo Lucide e PrimeIcons.
- Ícones internos do PrimeVue personalizados por slots, incluindo tabelas, calendários, seleção e notificações.
- SVGs com cor herdada do tema e carregamento por rota, sem fonte de ícones.

## Atualização — senhas e recuperação interna

- Admin pode alterar senhas e gerar links internos de recuperação, com expiração em 30 minutos e uso único.
- Tela pública de redefinição sem e-mail e Minha conta para troca mediante senha atual.
- Formulários PrimeVue com indicador de força e confirmação da nova senha; busca de usuários para localizar acessos.
- Migration `005_password_reset.sql`, hashes dos tokens e auditoria, com revogação transacional das sessões e links.
- Lock compartilhado com a criação de sessões para impedir login com credenciais antigas em uma corrida com a troca.
- Testes Go/PostgreSQL e Playwright de permissões, troca, recuperação e uso único.

## Atualização — middlewares e contexto da API

- Middlewares próprios extraídos para `gateway/api/middleware`, combinados com `RequestID`, `Recoverer` e `NoCache` do Chi.
- Library `pkg/ctxkey` com chaves privadas tipadas e funções Get/Put para headers, request ID, usuário, transporte e agente.
- Contexto e correlação compartilhados entre HTTP, WebSocket e OpenAI; respostas JSON e erros centralizados em `api/response`.
- Testes de credenciais, origem/CSRF, isolamento de contexto, eventos e correlação na integração.

## Atualização — visual do workspace e Kanban

- Identidade inspirada no TaskWhiz, com superfícies em carvão, destaque violeta e navegação lateral compacta.
- Tema escuro como padrão inicial, preservando a preferência salva e a opção de tema claro.
- Kanban com abas de navegação, filtros expansíveis e cartões com etiquetas, responsáveis, prazo e informações reais de QA.
- Cinco etapas e paginação preservadas, com rolagem por coluna em desktop e navegação adaptada para celular.

## Atualização — versões de entrega e arquivo do Kanban

- Cadastro de versões com notas e data opcional, vínculo nas demandas e consulta do conteúdo entregue.
- Arquivamento/restauração por versão com confirmação de demandas abertas e preservação de QA, histórico e métricas.
- Kanban consultado no servidor, filtros na URL, 20 cartões por coluna e carregamento adicional.
- Versão identificada nos cartões, detalhes, lista de demandas e CSV; migração preserva demandas existentes sem versão.
- Testes de integração com PostgreSQL real e fluxos Playwright de versionamento e paginação.

## Atualização — autenticação, relatórios com IA e tempo real

- Sidebar simplificado, login e tela de usuários exclusiva do único admin.
- JWT e refresh token em cookies HttpOnly, rotação transacional e revogação.
- Indicadores de prazo, retrabalho, throughput diário e lead time mediano/P85.
- OpenAI Responses com agentes de coleta/análise/revisão e JSON estruturado.
- Relatórios de demandas concluídas com testes, movimentações e snapshot persistido.
- Servidor MCP de leitura autenticado e WebSocket para atualizações e geração.
- Testes de backend/API/PostgreSQL/OpenAI simulada e E2E de autenticação e relatórios.


## v0.4 — SPA profissional com Vue, PrimeVue e Tailwind

- Refatorado o frontend em **SPA com Vue Router**, preservando a API e o backend Go no padrão do Bolão.
- Adicionado **Pinia** para centralizar estado, recarregamento, seleção de tickets e mutações.
- Substituídos controles manuais por **PrimeVue 4 (Aura)**: Dialog, Drawer, DataTable, Select, Inputs, Textarea, Button, Checkbox e Toast.
- Adicionado **Tailwind CSS 4** com componentes e estilos de identidade próprios.
- Implementadas rotas de dashboard, lista de tickets, Kanban, OS e iniciativas, QA e relatórios.
- Criados formulários modais de OS e entrega, drawer editável com transições, critérios de aceite, ciclos de QA e histórico.
- Implementada busca global por ticket, tema claro/escuro, navegação mobile e responsividade.
- Atualizados testes Playwright para a nova interface.
- Ajustado README com instruções SPA e limitações de validação.

**Validação nesta sessão:** testes unitários Go das camadas de domínio e usecases executados; análise sintática dos arquivos TS/Vue, tags HTML e CSS. Instalação npm, build Vue, Docker e E2E não foram executados por indisponibilidade de acesso ao registry npm e do runtime Docker.

---

# Changelog

## 0.3.0 — OS, itens compartilhados e corretivas avulsas

- Adicionada entidade `WorkPackage`, separada de `Demand`: OS agrupadora não entra no cálculo das métricas de execução.
- Nova migration `002_work_packages.sql`, sem apagar os dados anteriores.
- API `GET/POST /packages`, criação transacional opcional da primeira entrega e `PUT /demands/{id}` para edição de metadados.
- Backend mantém Clean/Hexagonal no padrão estrutural do Bolão, com `gateway`, `domain`, `usecase`, `app.go` e registry de UoW com escopo por transação.
- Corrigido bug herdado em que o caso de uso `AddTest` estava em `add_test.go` (arquivo de testes Go), renomeado para `add_test_case.go`.
- Novo frontend com navegação OS e iniciativas, OS dividida ou individual, tickets corretivos avulsos, filtro por responsável e relatórios por OS/colaborador.
- Seeds inteiramente fictícios demonstrando os três cenários e testes de pacote/UoW, edição e E2E.

## 0.2.0 — Refatoração arquitetural

- Reorganização para o padrão do projeto Bolão; composição de dependências em `app.go`, gateways e casos de uso.
- `TxScope` isolado por chamada, sem estado de transação compartilhado.

## 0.1.0 — MVP inicial

- Esteiras, Kanban, homologação QA, devolução automática, dashboard, Docker e CI.
