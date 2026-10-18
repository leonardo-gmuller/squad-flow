# Erros e respostas HTTP

Os erros de aplicação ficam em `backend/internal/app/domain/erring`. `AppError` contém um código estável e uma mensagem pública, sem dependência de HTTP. `Wrap` acrescenta a operação para diagnóstico e preserva `nil`, `errors.Is` e `errors.As`.

```go
var ErrInvalidCredentials = NewAppError(
    "auth:invalid-credentials",
    "Email ou senha inválidos.",
).WithCause(ErrUnauthorized)
```

`WithCause` associa um erro específico à categoria existente. `WithMessage` cria uma mensagem pública contextualizada sem modificar o sentinel compartilhado. Validações e conflitos conhecidos usam essas mensagens; detalhes técnicos ficam no encadeamento de erros e nos logs.

## Contrato de erro

Credenciais incorretas retornam HTTP `401`:

```json
{
  "type": "srn:error:unauthorized",
  "code": "auth:invalid-credentials",
  "message": "Email ou senha inválidos."
}
```

Conta inexistente e senha incorreta produzem a mesma resposta. Sessão expirada usa `auth:unauthorized`, e falhas inesperadas de banco continuam `500`, sem serem apresentadas como credenciais inválidas. Um erro não mapeado retorna `oops:internal-server-error` com mensagem genérica; prefixos de operação e mensagens de drivers não entram no JSON público.

As categorias são resolvidas com `errors.Is`: entrada inválida (`400`), autenticação (`401`), autorização (`403`), recurso ausente (`404`), método inválido (`405`), conflito/transição (`409`), corpo muito grande (`413`), limite de tentativas (`429`), indisponibilidade (`503`) e falha inesperada (`500`). O request ID permanece no cabeçalho `X-Request-ID`.

## Biblioteca de response

`gateway/api/response` contém os resultados `OK`, `Created`, `Accepted`, `NoContent` e `AppError`, com `WithHeaders` e `WithCookies`. `Send` escreve o resultado; `Handle` adapta funções que retornam uma resposta ao `net/http`. O login segue este formato:

```go
mux.Post("/api/v1/auth/login", response.Handle(h.login))

func (h *AuthHandler) login(r *http.Request) *response.Response {
    var input LoginInput
    if err := response.Decode(r, &input); err != nil {
        return response.AppError(err)
    }

    result, err := h.Service.Login(r.Context(), input.Email, input.Password)
    if err != nil {
        return response.AppError(err)
    }

    return h.sessionResult(result)
}
```

O exemplo omite o controle de tentativas e limites de comprimento, que permanecem no handler real. `Decode` limita o corpo a 1 MiB, exige um objeto JSON, rejeita campos desconhecidos e documentos adicionais e não expõe o corpo recebido nas mensagens de erro.

Handlers REST e middlewares compartilham o contrato. As respostas de sucesso preservam seus payloads atuais. MCP mantém o envelope JSON-RPC, e CSV/HTML continuam usando seus tipos de conteúdo próprios.

## Login com PrimeVue Forms e Zod

`frontend/src/validation/login.ts` define o schema usado pelo resolver Zod do PrimeVue Forms. O formulário exibe erros associados aos campos, valida ao sair do campo e ao enviar, foca o primeiro campo inválido e impede envio duplicado enquanto a chamada está em andamento. A validação nativa não substitui as mensagens do formulário.

O email perde somente espaços nas extremidades; a senha é preservada integralmente. O login exige senha preenchida e respeita o limite aceito pelo backend, sem impor novamente as regras de criação de senhas. A mensagem de credenciais inválidas aparece no formulário e é removida quando o usuário começa a corrigir os dados.

`APIError` preserva `status`, `type`, `code`, `message` e `requestID`. A leitura de mensagens antigas `{ "error": "..." }` permanece compatível no cliente, inclusive para os mocks existentes. Um `401` no login não dispara renovação de sessão.

Referências: [PrimeVue Forms](https://primevue.org/forms/), [schemas Zod](https://zod.dev/api).
