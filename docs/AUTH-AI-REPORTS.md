# Autenticação, relatórios com IA e atualizações

Os relatórios agora usam fila persistente e documentos HTML. O fluxo, a operação e os contratos atuais estão em [Relatórios assíncronos e HTML](ASYNC-HTML-REPORTS.md).

## Configuração

Preencha as novas variáveis no `.env` existente, usando `.env.example` como referência. Não substitua um `.env` que já contenha sua configuração de banco.

| Variável | Uso |
| --- | --- |
| `ENVIRONMENT` | `local` no desenvolvimento. `production` exige cookies Secure, origens HTTPS e seed demonstrativo desativado. |
| `DB_USER`, `DB_NAME`, `DB_PASSWORD`, `DB_HOST`, `DB_PORT`, `DB_SSLMODE` | Conexão da API local ao PostgreSQL. O exemplo corresponde ao banco de `docker-compose.dev.yaml`. |
| `DATABASE_URL` | Alternativa aos campos `DB_*`, com prioridade quando preenchida. O Compose completo define a URL internamente com host `db`. |
| `POSTGRES_PASSWORD` | Senha do PostgreSQL no ambiente completo de `compose.yaml`, que usa o usuário `squadflow`. |
| `JWT_SECRET` | Segredo com pelo menos 32 bytes. Gere com `openssl rand -hex 32`. |
| `ADMIN_EMAIL`, `ADMIN_PASSWORD` | Primeiro administrador. Senha de pelo menos 12 caracteres, limitada a 72 bytes pelo bcrypt. |
| `ADMIN_NAME` | Nome do administrador. |
| `OPENAI_API_KEY` | Chave somente no backend. Sem ela, indicadores e autenticação funcionam; a geração com IA fica indisponível. |
| `OPENAI_MODEL` | Modelo com function calling e Structured Outputs. Padrão configurável: `gpt-4o-mini`. |
| `COOKIE_SECURE` | `false` em HTTP local; `true` com HTTPS em produção. |
| `ALLOWED_ORIGINS` | Origens exatas do frontend, separadas por vírgula, sem barra final. |
| `ACCESS_TOKEN_TTL`, `REFRESH_TOKEN_TTL` | Padrões: 15 minutos e 7 dias. |
| `AI_REPORT_TIMEOUT`, `SERVER_WRITE_TIMEOUT` | Padrões: 210 segundos por tentativa da LLM e 30 segundos por escrita HTTP, independentes. |
| `PROCESS_ROLE`, `REDIS_URL`, `REPORT_WORKER_CONCURRENCY` | API/worker e conexão da fila; veja [configuração completa](ASYNC-HTML-REPORTS.md). |

O Compose repassa as credenciais ao backend e mantém API/banco na rede interna. Em desenvolvimento sem Compose, execute o backend a partir da raiz para carregar o `.env`, ou exporte as variáveis antes de `cd backend && go run ./cmd/api`. O Vite encaminha `/api` e WebSocket para a porta 8090.

Na primeira inicialização, a migration cria as tabelas e o bootstrap cadastra o administrador. Um lock transacional e um índice único parcial impedem um segundo admin. Nas inicializações seguintes, o admin existente é reutilizado; mudar as variáveis de bootstrap não altera sua senha. As variáveis de bootstrap podem ser removidas depois da criação. Não existe cadastro público nem promoção de usuário para admin.

Entre com o administrador e abra **Usuários** para cadastrar membros. Todos os usuários autenticados consultam e trabalham nas demandas/OS do workspace. Somente o administrador consulta a lista de usuários e cadastra novos acessos. O histórico de relatórios gerados é pessoal; as demandas e os indicadores continuam compartilhados.

O administrador também pode alterar senhas ou gerar links de recuperação internos, válidos por 30 minutos e uso único. Qualquer usuário conectado pode alterar sua senha em **Minha conta**. O login oferece **Esqueci minha senha**, com orientação para solicitar o link ao admin. Não há envio de e-mail. Veja [troca e recuperação de senha](PASSWORD-RECOVERY.md).

## Sessão e permissões

- JWT HS256 com issuer, audience e expiração verificados. Cookie `sf_access`, HttpOnly, SameSite Strict, path `/api`.
- Refresh token aleatório de 256 bits. Cookie `sf_refresh`, HttpOnly, SameSite Strict, path `/api/v1/auth`. O banco guarda somente SHA-256.
- Rotação transacional do refresh: o token anterior é consumido. Reutilizá-lo revoga toda a sessão, inclusive os JWTs já emitidos. A sessão mantém prazo absoluto de 7 dias.
- Logout revoga a sessão no banco e remove os cookies. Cada acesso autenticado verifica a sessão persistida, permitindo revogação imediata.
- Senhas bcrypt, proteção de origem e requisição JSON com `X-SquadFlow-Client: web` nas mutações. Limitação local de tentativas de login.
- Nenhum token vai para localStorage/sessionStorage ou para a URL do WebSocket. O localStorage existente guarda apenas a preferência de tema.
- A API substitui `actor` e `author` fornecidos pelo cliente pelo nome e ID do usuário autenticado. Os registros antigos preservam seus autores originais.

## Casos de uso e adapters

`domain/usecase/auth` declara portas para usuários/sessões, senhas e tokens. `gateway/security` implementa bcrypt/JWT; `gateway/postgres` implementa persistência, bootstrap e rotação. As rotas protegidas e cookies pertencem a `gateway/api`.

Os middlewares próprios de autenticação, origem/CSRF, contexto, logging e atualização do workspace ficam em `gateway/api/middleware`, combinados com `RequestID`, `Recoverer` e `NoCache` do Chi. A library `pkg/ctxkey` centraliza os valores compartilhados por HTTP, WebSocket e geração de relatórios. Veja [ordem de execução e contratos](API-MIDDLEWARES.md).

`domain/usecase/report` declara contratos para dados, geração e persistência. O caso de uso aplica escopo, exige demanda concluída para relatórios individuais, valida a resposta e salva o relatório com os dados consultados. `gateway/openai` implementa a Responses API. Domínio e casos de uso não importam HTTP, PostgreSQL ou o SDK da OpenAI.

O processo usa três papéis de agentes, em sequência:

1. **Coleta:** consulta as ferramentas permitidas e reúne as evidências.
2. **Análise:** gera um documento HTML autossuficiente, título, resumo e fontes sob JSON Schema estrito, com CSS, indicadores, tabelas e gráficos SVG.
3. **Revisão:** confere HTML e metadados contra as evidências e devolve o documento revisado.

Título e resumo possuem instruções explícitas no schema para evitar campos vazios. Para até 250 fontes, o schema restringe os IDs às demandas consultadas; a validação do caso de uso confere as fontes em todos os tamanhos de workspace. Se a revisão devolver componentes inválidos, há uma única chamada adicional de correção antes da validação final e do salvamento. A validação permanece obrigatória.

Não se expõe SQL, execução de código ou alterações de dados ao modelo. Contexto e pedidos são tratados como dados não confiáveis. As chamadas usam `store: false` por padrão (`OPENAI_STORE_RESPONSES=false`); os dados necessários ao relatório são enviados à OpenAI quando o usuário solicita a geração. O modelo continua sujeito a erros de interpretação: a revisão por IA e a validação de estrutura/fontes não garantem exatidão de todos os números gerados. A tela permite consultar as demandas e baixar o JSON com o snapshot para conferir a análise.

## Relatórios e indicadores

Em uma demanda finalizada, clique **Gerar relatório desta demanda**. O contexto inclui descrição atual, critérios de aceite, responsáveis, referências, todos os ciclos/testes e todo o histórico de movimentações. O relatório é gerado sob demanda depois da conclusão, podendo ser consultado novamente sem chamar a LLM.

A aba Relatórios aceita pedidos livres, demanda finalizada específica, período e uma pergunta de continuação baseada no resumo de um relatório anterior do próprio usuário. Cada pergunta consulta dados atuais e gera um novo relatório. O período filtra conclusão para demandas finalizadas e criação para as demais, em dias UTC.

Os indicadores fixos são calculados em Go, independentemente da IA:

- Lead time médio, mediano e P85: criação até conclusão, usando somente conclusões com timestamp. P85 usa o método nearest-rank.
- Primeiro passe: entregas finalizadas sem devolução / entregas finalizadas.
- Retrabalho: demandas com pelo menos uma devolução / todas as demandas. Total de devoluções soma todos os retornos, inclusive em demandas já finalizadas.
- Prazo vencido: demandas abertas com prazo em dia anterior a hoje. Entrega no mesmo dia do prazo conta como pontual.
- Entregas no prazo: finalizadas no prazo / finalizadas com prazo e data de conclusão.
- Throughput dos últimos 30 dias e série diária, incluindo dias sem entregas.
- Idade média das demandas abertas, trabalho sem responsável e distribuição por esteira/status/prioridade.

Limites da geração: uma solicitação ativa por usuário no banco, concorrência configurável por worker, até três tentativas automáticas, até oito rodadas de coleta com consultas em lote e até 24 consultas únicas de ferramentas, lista com até 2000 demandas e contexto limitado. Se necessário, reduza o período ou selecione uma demanda. Consultas repetidas reutilizam resultados da mesma geração, sem consultar novamente o banco. Se o agente repetir apenas consultas já respondidas, ou atingir o orçamento de rodadas/consultas, a coleta termina e a análise segue com as evidências obtidas. O relatório registra essa limitação em missing_data. A análise e a revisão usam tool_choice=none, encerrando o ciclo de ferramentas. Listas truncadas informam esse limite ao modelo. O histórico lista os 50 relatórios mais recentes; relatórios antigos podem ser consultados pelo ID.

## API e WebSocket

| Método | Endpoint | Uso |
| --- | --- | --- |
| POST | `/api/v1/auth/login` | Login; define cookies. |
| POST | `/api/v1/auth/refresh` | Rotação; define novos cookies. |
| POST | `/api/v1/auth/logout` | Revoga sessão e remove cookies. |
| GET | `/api/v1/auth/me` | Identidade atual. |
| GET, POST | `/api/v1/users` | Consulta/cadastro exclusivo do admin. |
| GET | `/api/v1/reports/status` | Disponibilidade da geração. |
| POST | `/api/v1/reports/generate` | Registra solicitação e responde `202 Accepted`. |
| GET | `/api/v1/reports/jobs` | Solicitações pessoais persistidas. |
| GET | `/api/v1/reports/jobs/{id}` | Estado, tentativas e etapas. |
| POST | `/api/v1/reports/jobs/{id}/retry` | Reenvio após falha definitiva. |
| GET | `/api/v1/reports/{id}/html?download=1` | Documento HTML para download. |
| GET | `/api/v1/reports?demand_id=<uuid>` | Histórico pessoal; filtro opcional. |
| GET | `/api/v1/reports/{id}` | Relatório pessoal e snapshot. |
| GET | `/api/v1/ws` | Upgrade autenticado. |
| POST | `/api/v1/mcp` | Ferramentas MCP de leitura. |

Mensagem WebSocket para solicitar análise:

```json
{"type":"report.generate","request_id":"pergunta-1","request":{"prompt":"Analise as devoluções de QA e recomende melhorias","demand_id":"","from":"","to":"","previous_report_id":""}}
```

Eventos de conexão: `connected` e `workspace.updated`. Solicitações legadas por WebSocket recebem `report.queued` com o job. O frontend registra o pedido por HTTP e acompanha o estado persistido, inclusive após recarregar a tela ou desconectar. A outbox e o gateway Redis substituem a geração vinculada à conexão. Veja [fila, recuperação e documentos HTML](ASYNC-HTML-REPORTS.md).

## Diagnóstico da geração

A criação retorna imediatamente `202` e um UUID de solicitação. O histórico de etapas fica em `report_events`, e o progresso atual em `report_jobs`. Logs do worker identificam job e tentativa; falhas do broker preservam os eventos para reenvio. A interface mostra fila, progresso e falha definitiva, com nova tentativa manual.

Cada chamada ao provedor registra `openai request started` e `openai request completed`, com modelo, etapa, request ID da OpenAI, response ID, tokens de entrada/saída e duração. Falhas registram o status HTTP e o código do provedor. Chaves, prompts, dados das demandas e respostas completas não são registrados. A interface mostra uma notificação de erro além da mensagem no formulário.

Um erro durante a geração provoca nova tentativa; após o limite, a solicitação fica com falha e mensagem ao usuário. Erros de chave, permissão, saldo/cota e limite de requisições possuem mensagens específicas. Um relatório salvo pode ser reaberto pelo histórico sem fazer novas chamadas à OpenAI.

Com `OPENAI_STORE_RESPONSES=true`, novas chamadas podem ser inspecionadas na plataforma OpenAI, no projeto correspondente à chave. Essa opção armazena os inputs/outputs no provedor. Recrie os processos após alterar o ambiente. Metadata identifica aplicação, job e etapa; isso não habilita tracing do Agents SDK.

## MCP

Servidor Streamable HTTP sem estado, com respostas JSON e compatibilidade com versões `2025-03-26`, `2025-06-18` e `2025-11-25`. Implementa `initialize`, `ping`, `tools/list`, `tools/call` e notificações de inicialização/cancelamento. Não anuncia recursos, prompts, SSE ou ferramentas de escrita.

Ferramentas disponíveis: `list_demands`, `get_metrics`, `get_demand_context`. Um cliente MCP deve enviar cookies de sessão ou `Authorization: Bearer <JWT de acesso>`, além de `Content-Type: application/json` e `X-SquadFlow-Client: web`. Um cliente com Bearer sem Origin é aceito; se fornecer Origin, ela precisa estar na lista permitida. Tokens expiram normalmente; não há token MCP permanente ou OAuth de terceiros nesta implementação.

Como a OpenAI não acessa a rede local do Compose, os agentes usam function calling: o backend executa as mesmas ferramentas que o servidor MCP oferece. Não é necessário expor o banco ou o endpoint local publicamente para gerar relatórios. Conectar clientes MCP externos ao servidor e usar um MCP remoto diretamente pela OpenAI são configurações distintas; não há MCP remoto público configurado aqui.

## Validação

```bash
make test-unit
make test-integration # requer Docker; Testcontainers cria PostgreSQL e Redis

cd frontend
npm run build
E2E_EMAIL='admin-de-teste@example.com' E2E_PASSWORD='senha-do-admin-de-teste' npm run test:e2e
```

Os testes de integração criam contêineres descartáveis e não acessam o banco da aplicação. Para E2E, use contas descartáveis. Veja [TESTING.md](TESTING.md). As chamadas OpenAI são simuladas nos testes, sem enviar dados ou consumir créditos. Para verificar a integração real, configure sua chave no backend e gere um relatório pela tela.

Referências: [OpenAI Structured Outputs](https://developers.openai.com/api/docs/guides/structured-outputs?api-mode=responses), [OpenAI MCP](https://developers.openai.com/api/docs/guides/tools-connectors-mcp), [MCP Streamable HTTP](https://modelcontextprotocol.io/specification/2025-06-18/basic/transports), [OWASP Session Management](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html).
