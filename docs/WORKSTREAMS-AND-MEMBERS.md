# Esteiras e responsáveis cadastrados

## Uso

O administrador gerencia **Esteiras** em `/workstreams`: cadastra nome e cor, edita e desativa/reativa. Somente contas admin podem alterar o catálogo; membros autenticados podem consultá-lo e usar as esteiras ativas. As quatro esteiras iniciais são Melhorias, Corretivas, Testes e Reescritas, editáveis pelo admin.

As esteiras são categorias de trabalho; as cinco etapas do Kanban e as regras de QA continuam independentes. Desativar uma esteira impede novos vínculos em demandas e OS, mantendo consultas, relatórios, métricas e histórico. Demandas existentes podem continuar sendo editadas e movimentadas.

Desenvolvedor e QA são selecionados entre usuários cadastrados, nos cadastros de demandas, na primeira entrega de uma OS e na edição do drawer. É possível deixar a atribuição como **Não atribuído**. O seletor pesquisa nomes e distingue homônimos pelo email. Movimentações e testes usam sempre a conta autenticada; os nomes enviados pelo cliente não definem a autoria.

## API

Todas as rotas exigem autenticação.

| Método e rota | Permissão | Comportamento |
| --- | --- | --- |
| `GET /api/v1/workstreams` | Membro ou admin | Lista esteiras ativas e inativas, incluindo cor e nome para exibir históricos. |
| `POST /api/v1/workstreams` | Admin | Recebe `name`, `color` (`#RRGGBB`) e `active` opcional, inicialmente `true`. |
| `PUT /api/v1/workstreams/{id}` | Admin | Edita nome/cor; `active` opcional permite desativar/reativar. |
| `GET /api/v1/members` | Membro ou admin | Diretório de atribuição com `id`, `name` e `email`, sem credenciais. |

A identidade `code` da esteira é imutável. O campo `kind` de demandas e OS usa esse código; novos códigos são gerados pelo servidor. Nomes duplicados de esteiras são rejeitados, sem diferenciar maiúsculas e espaços nas extremidades. A exclusão física não é exposta.

Criação e edição de demandas, e criação da primeira entrega de OS, recebem `assignee_id` e `reviewer_id`, com UUID do usuário ou string vazia para ausência de atribuição. Os antigos campos de texto `assignee` e `reviewer` são rejeitados nas requisições e permanecem na resposta como nomes para exibição. A validação de existência ocorre na mesma transação da gravação. As chaves estrangeiras protegem os vínculos no banco.

O filtro `owner` de `/api/v1/board` usa o ID do usuário. Relatórios por responsável agrupam por ID, preservando a distinção entre homônimos. Os gráficos usam o catálogo e incluem automaticamente novas esteiras e suas cores. Os nomes cadastrados são exibidos literalmente, mesmo quando parecem datas ou dias da semana. CSV e contexto de relatórios também incluem o nome da esteira.

Transições expõem `actor_id`; testes expõem `author_id`. O servidor preenche esses IDs pela sessão autenticada. Os textos de autoria são mantidos como registros históricos.

## Atualização de uma base existente

A migração `006_workstreams_members.sql` roda no bootstrap normal da API. Ela cria o catálogo inicial, troca os checks fixos por referências ao catálogo e acrescenta vínculos de usuários e autoria.

Nomes antigos de desenvolvedores e QA são associados automaticamente somente quando correspondem a um único usuário cadastrado, desconsiderando caixa e espaços nas extremidades. Nomes sem correspondência ou ambíguos permanecem no registro histórico, sem criar contas nem escolher pessoas arbitrariamente. O drawer avisa quando um responsável antigo ainda não tem conta vinculada; ao editar, selecione a conta correta ou **Não atribuído**. Autorias antigas que já contêm o UUID de uma conta são vinculadas a ela.

## Seed de demonstração

`DEMO_SEED=true` mantém as demandas/OS de exemplo e cadastra cinco membros fictícios:

| Nome | Email |
| --- | --- |
| Alex | `alex@example.test` |
| Rafa | `rafa@example.test` |
| Daniel | `daniel@example.test` |
| Marina | `marina@example.test` |
| Júlia | `julia@example.test` |

Senha inicial dessas contas de demonstração: `SquadFlow-Demo-2026!`. A senha é armazenada com hash; executar o seed novamente não altera senhas nem duplica usuários, OS ou demandas. Os responsáveis das nove entregas de demonstração usam essas contas. O seed pode atualizar os vínculos dos registros de demonstração já existentes, sem reatribuir demandas reais.

`DEMO_SEED=false` não cria contas ou demandas fictícias. O catálogo inicial permanece disponível. A configuração já impede habilitar o seed no ambiente de produção.

## Verificação

- Backend: `make test-unit`; para testes com banco, execute `make test-integration` com o Docker em execução. O Testcontainers cria bases descartáveis isoladas. Veja [TESTING.md](TESTING.md).
- Frontend: `cd frontend && npm run build`.
- E2E: `npm run test:e2e` com `E2E_BASE_URL`, `E2E_EMAIL` e `E2E_PASSWORD` de uma instância com dados fictícios.

Os testes cobrem permissões, categorias personalizadas nos filtros, atribuição por ID, rejeição de texto livre/usuário inexistente, autoria autenticada, desativação, edição sem perder vínculos, migração de nomes ambíguos e repetição do seed.
