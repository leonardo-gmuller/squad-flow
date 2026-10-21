# Middlewares e contexto da API

`backend/internal/app/gateway/api/api.go` compõe o roteador Chi. Os middlewares próprios ficam em `gateway/api/middleware`; handlers recebem a requisição já preparada e consultam os valores compartilhados por `pkg/ctxkey`.

## Ordem de execução

1. **Chi `RequestID`:** preserva `X-Request-ID` recebido ou gera um identificador.
2. **`RequestContext`:** disponibiliza esse identificador, `Authorization`, `Idempotency-Key` e o transporte HTTP por `ctxkey`; devolve o identificador em `X-Request-ID` na resposta.
3. **`Observability`:** extrai contexto W3C, cria span HTTP e registra métricas por template da rota. Expõe `X-Trace-Id`; health checks e WebSocket ficam fora da medida de latência HTTP. Veja [OBSERVABILITY.md](OBSERVABILITY.md).
4. **`RequestLogger`:** usa `slog` e o response writer do Chi para registrar método, caminho, status, duração e request ID, sem registrar credenciais.
5. **Chi `Recoverer`:** trata panics em todas as rotas, inclusive autenticação e relatórios, com HTTP 500 quando a resposta ainda pode ser escrita. O corpo de um panic segue o comportamento do Chi; erros tratados usam o contrato JSON de `api/response`.
6. **Chi `NoCache`:** desabilita cache das respostas da API.
7. **`OriginPolicy.Protect`:** valida origens exatas e mantém a exigência de JSON com `X-Squadflow-Client: web` para mutações. Clientes Bearer podem omitir Origin, mas precisam desses headers.

Login, refresh, logout e health check são públicos. O grupo protegido recebe `RequireAuth`, que valida Bearer ou cookie `sf_access`, consulta a sessão e coloca o usuário autenticado no contexto. Bearer tem precedência sobre o cookie; um Bearer inválido não usa o cookie como alternativa. Permissões de administrador continuam nos casos de uso.

O endpoint público `POST /api/v1/auth/password-reset` consome links internos de recuperação. Ele recebe `RateLimiter` de `api/middleware`, com 30 tentativas por endereço de origem a cada cinco minutos e armazenamento limitado por instância. O endereço vem da conexão, sem confiar em headers de proxy. Alteração de senha própria e gerenciamento de senhas por admin pertencem ao grupo protegido.

`WorkspaceUpdates` envolve as rotas de mutação do workspace e publica `workspace.updated` após POST/PUT bem-sucedidos. Usa o response writer do Chi para observar o primeiro status efetivamente enviado. O WebSocket recebe a mesma política de origem e mantém a revalidação da sessão.

## Library `ctxkey`

Localização: `backend/internal/app/pkg/ctxkey`. As chaves são privadas, tipadas e distintas; consumidores usam funções `Put...` e `Get...`, sem acessar `context.WithValue` diretamente. Os getters retornam `(valor, bool)`, distinguindo ausência de um valor vazio.

```go
ctx = ctxkey.PutRequestID(ctx, "request-123")
requestID, ok := ctxkey.GetRequestID(ctx)
```

Valores disponíveis: `AuthorizationHeader`, `IdempotencyKey`, `RequestID`, `User`, `Transport` e `Agent`. Cada setter retorna um contexto derivado e preserva cancelamento, prazo e valores do contexto original.

O header `Idempotency-Key` é somente propagado. Não há deduplicação, armazenamento de respostas ou garantia de execução única nas mutações.

Cada geração por WebSocket deriva seu próprio contexto, usa o `request_id` da mensagem e define transporte `websocket`, sem alterar o contexto da conexão. Relatórios HTTP mantêm o identificador da requisição. O adapter OpenAI lê o mesmo request ID e deriva o nome do agente em `ctxkey`, mantendo a correlação dos logs.

Respostas JSON e o mapeamento de erros ficam em `gateway/api/response`, compartilhados por handlers e middlewares sem dependência entre esses dois pacotes. A limitação de tentativas de login permanece no fluxo de login, que considera o endereço e o e-mail já decodificado.

## Validação

Os testes de middleware cobrem isolamento dos metadados por requisição, IDs gerados/recebidos, precedência de credenciais, bloqueio de autenticação, regras de origem/CSRF e publicação de eventos conforme o status enviado. O teste de integração da API também verifica a correlação e a identidade autenticada na geração HTTP e WebSocket, usando PostgreSQL real e geração de IA simulada.

## Respostas de erro

Autenticação, autorização, origem/CSRF e limites de tentativas usam `gateway/api/response` e o envelope `{ type, code, message }`. O mapeamento preserva erros encadeados com `errors.Is` e impede que mensagens internas sejam enviadas ao cliente. O contrato completo e os exemplos estão em [API-ERRORS.md](API-ERRORS.md).
