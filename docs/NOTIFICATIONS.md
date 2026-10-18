# Central de notificações

A central aparece pelo botão de sino no topo do workspace, com contador de não lidas. É exclusiva dos usuários com papel `member`: o admin não recebe notificações, não vê o botão e tem acesso negado às rotas da central. O indicador “Ao vivo” foi removido para todos; WebSocket continua atendendo o workspace e a geração de relatórios.

## Eventos e destinatários

- Atribuição de desenvolvimento e/ou QA em uma nova demanda, inclusive a primeira entrega criada com uma OS.
- Mudança de atribuição: novos responsáveis são avisados; quem deixa de acompanhar recebe aviso da alteração.
- Mudanças de etapa, incluindo envio para QA, devolução para correção e conclusão.
- Cenários de QA aprovados. Um teste reprovado produz o aviso de devolução, sem duplicar um aviso de reprovação.

Os destinatários são os usuários cadastrados como desenvolvedor e/ou QA da demanda. Uma pessoa nos dois papéis recebe um único aviso por evento. A conta autenticada que executa a ação não recebe aviso da própria ação, e contas admin são sempre excluídas.

As notificações são gravadas na mesma transação da demanda, alteração ou teste. Se a operação falhar ou for revertida, seus avisos também são revertidos. Os nomes exibidos são texto escapado; não há envio por email ou serviço externo.

## Leitura e atualização

O painel lateral usa Drawer do PrimeVue, Heroicons e os botões padrão do projeto. Permite alternar entre **Todas** e **Não lidas**, marcar uma ou todas como lidas, atualizar e carregar mais. Clicar em uma notificação marca a leitura e abre a demanda no drawer de detalhes. O estado de leitura fica no banco e persiste entre sessões.

A lista carrega 20 itens por página e usa cursor por data/ID para manter a navegação estável quando chegam novos avisos. Datas relativas e dicas com o horário completo usam Moment em português. Há estados de carregamento, lista vazia e erro com nova tentativa, além de suporte a teclado, foco e celular.

O evento `workspace.updated` já existente provoca uma consulta privada ao contador/lista de cada membro; nenhum conteúdo da caixa de uma pessoa é enviado aos demais. Alterações de leitura publicam `notifications.changed` somente para as conexões do próprio usuário, sincronizando outras abas. A reconexão e o retorno à aba atualizam a central. Se o WebSocket estiver desconectado, há atualização a cada 30 segundos enquanto a aba estiver visível.

## API

As rotas abaixo exigem sessão de um membro. Usuário sem sessão recebe 401; admin recebe 403. O destinatário vem da sessão, sem parâmetro de escolha de usuário.

| Método e rota | Comportamento |
| --- | --- |
| `GET /api/v1/notifications` | Lista privada; parâmetros `filter=all\|unread`, `limit` entre 1 e 50 (padrão 20), e `before` com o cursor retornado anteriormente. |
| `GET /api/v1/notifications/unread-count` | Retorna `{ "count": número }` somente para a sessão atual. |
| `POST /api/v1/notifications/{id}/read` | Marca uma notificação própria como lida e retorna `read_at`; é idempotente. ID de outra pessoa retorna 404. |
| `POST /api/v1/notifications/read-all` | Marca somente a caixa do usuário atual e retorna `read_at`. |

A listagem retorna `items`, `unread_count` e `next_cursor`. Cada item contém `id`, `demand_id`, `kind`, `title`, `message`, `read_at` e `created_at`. A consulta de página e contador usa o mesmo snapshot do banco. A central não expõe a identidade de outros destinatários.

## Migração e seed

A migração `007_notifications.sql` é aplicada automaticamente ao iniciar a API. Não transforma demandas antigas em avisos em massa: o histórico começa com as próximas ações.

Com `DEMO_SEED=true`, cada uma das cinco contas de demonstração recebe um aviso vinculado a uma entrega de exemplo. O seed usa IDs estáveis: reiniciar não duplica avisos nem apaga seu estado de leitura. Os dados e acessos fictícios estão em [WORKSTREAMS-AND-MEMBERS.md](WORKSTREAMS-AND-MEMBERS.md).

## Organização e validação

- `domain/usecase/notification/`: permissão de membro, paginação e leitura da caixa da sessão.
- `gateway/postgres/repository/notification/`: persistência e geração dos eventos dentro das transações existentes de demandas/testes.
- `gateway/api/middleware/member.go`: restrição das rotas a membros, usando as chaves de contexto existentes.
- `gateway/api/handler/notification_handler.go`: HTTP e aviso privado de atualização da leitura.
- `frontend/src/components/NotificationCenter.vue`: botão, contador, filtros e painel.

Testes cobrem privacidade entre usuários, ausência do admin, deduplicação de papéis, exclusão de ações próprias, transações revertidas, QA, cursor com novas chegadas, seed idempotente, persistência de leitura, WebSocket, abertura da demanda, erros, teclado e responsividade. Os relatórios via WebSocket continuam cobertos pelos testes existentes.
