# Troca e recuperação interna de senha

O administrador gerencia o acesso em **Usuários**. A busca encontra pessoas por nome ou e-mail. Não é necessário configurar envio de e-mail.

## Alteração pelo administrador

Em **Alterar senha**, informe e confirme a nova senha do usuário. A operação encerra todas as sessões dessa pessoa, incluindo refresh tokens, e invalida seus links de redefinição pendentes. Se o administrador alterar a própria senha, também precisará entrar novamente.

## Recuperação sem e-mail

1. Em **Usuários**, clique no ícone de link da pessoa e depois em **Gerar link**.
2. Copie o link e entregue diretamente ao usuário pelo canal interno da equipe. O sistema não envia mensagens.
3. O usuário abre o link e informa e confirma a própria senha, sem precisar conhecer a anterior.
4. Após concluir, ele entra com a nova senha. O link não pode ser usado novamente.

O link vale por **30 minutos**. Gerar outro link invalida os anteriores. A geração preserva as sessões atuais; a redefinição bem-sucedida encerra essas sessões. Links expirados, substituídos ou usados recebem a mesma mensagem de erro. O link fica disponível na janela em que foi gerado; depois de fechá-la, gere outro se precisar.

O link **Esqueci minha senha** na tela de login abre `/reset-password`, explica o procedimento e aceita colar o link ou código recebido do administrador. Nenhum formulário público consulta se um e-mail está cadastrado. O procedimento depende de um administrador com acesso ao workspace; não cria uma recuperação alternativa do administrador por variáveis de bootstrap.

## Troca pela própria pessoa

Em **Minha conta**, qualquer usuário autenticado pode trocar sua senha informando a atual e confirmando a nova. Uma senha atual incorreta mantém a sessão. Após sucesso, as sessões são encerradas e a tela pede um novo login.

Os formulários usam Password, feedback de força, Dialog e Message do PrimeVue. O indicador Fraca/Média/Forte oferece sugestões; a validação exige pelo menos 12 caracteres Unicode e respeita o limite de 72 bytes do bcrypt, com mensagens simples na interface. Contas existentes continuam usando suas senhas atuais até uma alteração explícita.

## API e persistência

| Método | Endpoint | Acesso |
| --- | --- | --- |
| PUT | `/api/v1/users/{id}/password` | Admin; corpo `{ "password": "..." }` |
| POST | `/api/v1/users/{id}/password-reset` | Admin; retorna `token` e `expires_at` |
| PUT | `/api/v1/auth/password` | Autenticado; corpo `{ "current_password": "...", "password": "..." }` |
| POST | `/api/v1/auth/password-reset` | Público; corpo `{ "token": "...", "password": "..." }` |

As mutações mantêm proteção de origem e os headers `Content-Type: application/json` e `X-Squadflow-Client: web`. O endpoint público limita tentativas por endereço de origem a 30 a cada cinco minutos, por instância. Ele não faz login automático após a redefinição. A confirmação da senha ocorre na interface; a API recebe a senha final.

A migration `005_password_reset.sql`, aplicada ao iniciar a API, adiciona `password_reset_tokens` e `user_password_events`. Tokens aleatórios de 256 bits são armazenados como SHA-256 com domínio próprio, e só o token original retornado na geração permite redefinir. O banco não guarda senhas em texto nem tokens de recuperação originais. O histórico registra usuário, autor, ação e horário.

Troca de senha, consumo dos tokens, revogação das sessões e auditoria são transacionais, com lock no usuário. A criação de sessão usa o mesmo lock e confere se a senha verificada ainda corresponde ao hash persistido, impedindo que um login iniciado antes da troca crie uma sessão depois dela. Redefinições simultâneas do mesmo token têm no máximo um sucesso.

O frontend constrói o link com sua própria origem e leva o token no fragmento `#token=...`, que não faz parte da URL enviada ao servidor. Ao abrir a tela, move o token para a memória e remove o fragmento da barra de endereço. A página usa política de referrer `no-referrer`. O token é enviado à API no corpo da redefinição; não é salvo em localStorage ou sessionStorage e não é publicado por WebSocket. Logs HTTP registram somente o caminho e metadados da requisição.

As decisões de token aleatório, prazo, uso único, armazenamento e revogação seguem as orientações da [OWASP para recuperação de senha](https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html).

## Validação

Testes Go com PostgreSQL real cobrem permissões, alteração pelo admin e pelo usuário, senha atual incorreta, expiração, substituição, tentativa de reutilização, hash persistido, sessões/refresh revogados, auditoria, consumo concorrente e login com senha antiga em uma corrida. Os testes Playwright exercitam os mesmos fluxos pela interface e conferem que o token não aparece nas URLs de requisições.
