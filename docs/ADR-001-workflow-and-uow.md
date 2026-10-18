# ADR 001 — Fluxo de demandas e Unit of Work

**Status:** aceito para o MVP 0.1  
**Contexto:** demandas passam por desenvolvimento e múltiplos ciclos de QA, gerando fatos históricos necessários para indicadores de qualidade.

## Decisões

1. As transições permitidas são invariantes no domínio Go, não apenas na interface.
2. A camada de aplicação declara as portas `UnitOfWork`, `Repositories` e `QueryRepository`.
3. O adapter PostgreSQL implementa `UnitOfWork.Do(ctx, callback)`, criando `pgx.Tx`, repassando repositórios vinculados à **mesma transação** e efetuando `Commit` somente após sucesso.
4. Cada alteração de status também insere uma transição histórica, atomicamente.
5. `GetForUpdate` executa `SELECT ... FOR UPDATE`, serializando a alteração de uma demanda.
6. Um teste reprovado insere o resultado, altera status para `changes_requested`, incrementa devoluções e gera uma transição **na mesma transação**.
7. Cada envio a QA incrementa `qa_cycle`; o comando de conclusão só considera testes no ciclo atual.
8. Relatórios usam projeções dos dados existentes, evitando acoplamento de persistência ao frontend.
9. As queries ficam em arquivos `.sql`; SQLC gera os parâmetros, resultados e chamadas pgx. Os repositories implementam as interfaces dos casos de uso, com arquivos separados para operações e mappers. Nenhum tipo gerado pelo SQLC entra no domínio.
10. Os repositories recebem o `DBTX` da UoW e constroem `sqlc.New(db)` sobre a transação correspondente. Operações de autenticação que controlam sua própria transação usam `Queries.WithTx(tx)`. Validações com locks e notificações recebem as mesmas queries transacionais. Veja [SQLC e repositories](SQLC-REPOSITORIES.md).

## Consequências positivas

- Operações de múltiplas tabelas são atômicas.
- Não há dependência de `pgx` nos casos de uso.
- Retrabalho é mensurável ao longo de diferentes tentativas.
- Regras são cobertas por testes unitários rápidos com fakes transacionais e integração real com PostgreSQL.

## Limitações conscientes

- Relatórios usam outbox transacional, Redis Streams e worker com recuperação; eventos de atualização do workspace continuam locais à instância. Veja [relatórios assíncronos](ASYNC-HTML-REPORTS.md).
- Workflow é fixo em código; configuração dinâmica fica para outra versão.
- Relatórios usam varredura de registros, suficiente para um piloto; consultas agregadas otimizadas serão necessárias para volumes grandes.
- A extensão atual acrescenta autenticação e papéis admin/member; configuração e limitações em [AUTH-AI-REPORTS.md](AUTH-AI-REPORTS.md).
