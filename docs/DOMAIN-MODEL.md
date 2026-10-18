# Modelo de trabalho: esteiras, OS e entregas

Uma **iniciativa** é um agrupamento lógico em texto, não uma Nota Técnica obrigatória. Ex.: modernização, compliance, evolução da API, NT. Uma **OS** (`work_packages`) é uma pasta que agrupa entregas e não entra nos contadores de produtividade/QA. Uma **entrega** (`work_items`) é a unidade de execução e de contabilização; pertence a zero ou uma OS. **Corretivas** podem ser tickets independentes com `reference=BUG-123` e `package_id` vazio. As quatro esteiras (`kind`) continuam no item: melhoria, corretiva, teste e reescrita.

## Dois cenários que a aplicação atende

- OS-100 — 1 item "Implementação completa" atribuído a Alex. Desenvolvedor e QA próprios.
- OS-200 — itens "Alteração de processamento" atribuído a Alex e "Validação de XML" atribuído a Daniel. Cada um tem status e QA independentes.
- BUG-500 — corretiva avulsa sem OS, com suas transições e testes.

Uma OS recém-criada pode ter **zero itens**, ou criar a primeira entrega na mesma transação. Adicionar mais itens posteriormente pela aba OS. Não há exigência de Nota Técnica. Somente itens entram no relatório de throughput, lead time e retrabalho; progresso de OS é contagem de itens fechados dividida pelo número de itens.

## Versões de entrega

Uma versão (`releases`) agrupa demandas pelo código entregue. Cada demanda pode ter zero ou uma versão (`release_id`), independentemente da OS. A mesma OS pode ser entregue em versões distintas. Os nomes são livres e únicos sem diferenciar maiúsculas/minúsculas.

Arquivar uma versão registra `archived_at` e altera a visibilidade padrão do Kanban. Não altera status, datas de conclusão, testes, transições nem métricas históricas. É possível consultar e restaurar versões arquivadas; novas vinculações exigem uma versão ativa. Veja [RELEASES.md](RELEASES.md).

## UoW

`domain/usecase/package` depende apenas dos contratos de repositório e UoW. O adapter PostgreSQL implementa a transação `DoPackage` em `gateway/postgres/unit_of_work.go`. Se falhar o primeiro item, faz rollback da OS. Transações ficam em scopes por requisição, diferentemente da transação mutável compartilhada do projeto de referência.

## Pontos futuros

Autenticação, identidade verificável em transições/testes, indicadores e relatórios com OpenAI estão implementados; veja [AUTH-AI-REPORTS.md](AUTH-AI-REPORTS.md). Edição com histórico próprio, status configuráveis, artefatos, Redmine e SVN permanecem fora do escopo atual.
