# Versões de entrega e arquivo do Kanban

A tela **Versões** permite criar e editar versões com nome livre (por exemplo `2.6.0`, `2026.10` ou um nome de release), descrição e data de entrega opcional. O nome é único no workspace, sem diferenciar maiúsculas/minúsculas nem espaços nas pontas. Não há incremento automático de versão.

Cada demanda tem uma versão opcional. O vínculo é informado ao criar ou editar a demanda, ou ao criar a primeira entrega de uma OS. A versão pertence à demanda: uma OS pode reunir entregas de versões diferentes. A tela de versões apresenta totais, concluídas, itens abertos, QA e a lista paginada das demandas vinculadas. Cards, detalhes, lista de demandas e CSV também identificam a versão.

## Arquivamento

- Arquivar uma versão a remove da visão **Versões ativas** do Kanban. Demandas sem versão continuam nessa visão.
- A confirmação apresenta a quantidade de demandas e avisa quando ainda existem entregas abertas.
- Arquivar não conclui demandas nem apaga informações. Estados, casos de teste, movimentações, datas e métricas históricas permanecem disponíveis.
- As visões **Versões arquivadas** e **Todas, incluindo arquivadas** permitem consultar o histórico. A lista de demandas, QA, OS, dashboard e relatórios mantêm o histórico completo.
- Uma versão arquivada pode ser restaurada. Suas demandas voltam ao quadro ativo nos estados anteriores.
- Novas vinculações a versões arquivadas são recusadas pela API. É permitido editar ou movimentar demandas já vinculadas, ou removê-las da versão. A atribuição e o arquivamento são coordenados por locks transacionais no PostgreSQL.

## API

Todas as rotas exigem sessão e seguem as proteções de origem/CSRF do workspace. Membros autenticados podem gerenciar versões, assim como OS e demandas.

- `GET /api/v1/releases`: versões ativas e arquivadas, com totais (`total`, `done`, `in_qa`).
- `POST /api/v1/releases`: `{ "name": "2.6.0", "description": "Notas", "release_date": null }`.
- `PUT /api/v1/releases/{id}`: os mesmos campos de criação; não altera vínculos nem arquivo.
- `POST /api/v1/releases/{id}/archive`: `{ "archived": true }` para arquivar ou `false` para restaurar. Repetir o arquivamento preserva a data original.
- `POST /api/v1/demands`: aceita `release_id` opcional. String vazia significa sem versão.
- `PUT /api/v1/demands/{id}`: `release_id` omitido preserva o vínculo; string vazia o remove.
- `POST /api/v1/packages`: `release_id` aplica-se somente à primeira entrega automática, quando solicitada.

`GET /api/v1/board` aceita:

| Parâmetro | Comportamento |
|---|---|
| `archive` | `active` (padrão), `archived` ou `all` |
| `release_id` | UUID de versão, `none` para sem versão, ou ausente para todas |
| `kind`, `owner`, `q` | Esteira, responsável exato e busca textual |
| `status` | Opcional; restringe a uma coluna |
| `limit` | Cartões por coluna, padrão 20, máximo 100 |
| `offset` | Padrão 0; páginas seguintes exigem `status` |

A resposta contém `columns: [{ status, total, items }]` e `owners`. As contagens e as páginas usam o mesmo snapshot de consulta. Os cartões são resumos: descrição limitada a 220 caracteres e sem critérios de aceite; os detalhes completos vêm de `/demands/{id}`. A ordem é atualização decrescente, com número da demanda como desempate. A interface mantém os filtros na URL, descarta respostas de filtros anteriores e recarrega o quadro após atualizações do workspace.

## Persistência e validação

A migração `004_releases.sql` cria versões, a chave estrangeira opcional e índices para consulta por versão/etapa. Dados existentes são preservados e começam sem versão. A migração é versionada e aplicada no startup da API.

Os testes de integração usam PostgreSQL real para validar unicidade, criação/edição, paginação, demandas sem versão, arquivamento, restauração, histórico de QA e manutenção das métricas. Os testes Playwright de `releases.spec.ts` cobrem o fluxo completo e um quadro com 42 cartões, além da ausência de download da lista completa no Kanban. `board-drag.spec.ts` cobre arraste por mouse e toque, destinos permitidos, cancelamento, motivo, falhas da API e teclado; `board-drag-api.spec.ts` percorre o ciclo de correção e aprovação de QA pela API real e verifica autor e histórico persistidos.
