# SquadFlow — Workspace de iniciativas e notas técnicas

> Especificação da próxima etapa, baseada em padrões de planejamento e entrega de software. **Este documento é genérico** e não incorpora dados, textos ou código de qualquer organização.

## Problema que resolve

Um item de uma especificação não é a mesma coisa que uma OS, e uma OS não é necessariamente uma tarefa executável. Uma mesma mudança pode envolver múltiplos produtos; uma OS pode agrupar múltiplas mudanças; e um ticket pode exigir mais de uma rodada de QA. Misturar esses conceitos gera dupla contagem e perda de rastreabilidade.

## Modelo de negócio proposto

```text
Initiative (ex.: mudança normativa, projeto interno)
  ├─ ReleaseWindow (marcos, ambientes, prazos)
  ├─ SourceDocument (origem e versão do conteúdo)
  └─ SpecificationItem [N]
      ├─ SpecificationRevision (texto original, versão, diferença)
      ├─ SolutionProposal [N] (estudo, escolhida, descartada)
      ├─ Impact [N] (produto, evento, componente)
      ├─ WorkItemLink [N] ────────┐
      └─ SourceReference [N]     │
                                  │
WorkItem (melhoria, corretiva, teste, reescrita)
  ├─ ExternalIssueRef [N] (Redmine / OS externa)
  ├─ ParentLink [0..1] / ChildLink [N]
  ├─ Assignee / QA reviewer / estimativa / prazo / complexidade
  ├─ AcceptanceCriterion [N]
  ├─ WorkflowTransition [N] (autor, data, de, para, motivo)
  ├─ TestCycle [N] -> TestCase [N] -> TestRun [N]
  ├─ Defect [N] (reaberto, corrigido, verificado)
  └─ Evidence / Attachment / Comment / Decision / CommitReference [N]
```

### Regras e cardinalidades

- `Initiative` agrupa itens de especificação; `WorkItem` é o que entra no fluxo operacional.
- A associação entre `SpecificationItem` e `WorkItem` é **N:N**, com tabela de vínculos. Isso evita duplicar tickets quando há impacto em múltiplos eventos/produtos.
- `WorkItem` pode ser pai/filho de outros itens para representar uma OS geral, OS por produto e tarefas técnicas.
- `SolutionProposal.status` **não** é `WorkItem.status`: “proposta escolhida” não significa “desenvolvimento concluído”.
- Manter `event_code` e `field_path` como dimensões consultáveis, não como parte da identidade única de um ticket.
- `ReleaseWindow` deve separar data de ambiente restrito, produção, versão e indicador de risco.
- Um ciclo QA aprovado precisa corresponder à revisão de implementação testada; ao reenviar, cria-se um novo ciclo.
- Mudanças relevantes ficam no histórico imutável, com transições atômicas via Unit of Work.

## Telas — manter o visual limpo do MVP

### 1. Iniciativas

Tabela/cartões de iniciativas (status global, responsáveis, percentual com execução, riscos, prazos). Não despejar centenas de detalhes na tela inicial.

### 2. Workspace da iniciativa

Cabeçalho com nome, marcos da release, saúde geral e links de origem. Abas:

1. **Resumo** — andamento por produto, status, alertas e métricas.
2. **Itens** — catálogo filtrável por evento, campo, prioridade, equipe, impacto.
3. **Entregas / OS** — árvore de OS principal → OS filha → tarefas executáveis.
4. **Cronograma** — janelas de homologação e produção, versões e dependências.
5. **Documentos** — referências e histórico de revisões.
6. **Decisões** — propostas técnicas, aprovação, dúvidas e esclarecimentos.

### 3. Detalhe de item / entrega

Layout com coluna central para escopo e critérios de aceite, painel lateral com responsáveis, prazo e status; abas **Contexto | De → Para | Solução | Desenvolvimento | QA | Histórico | IA**. Exibir contextos extensos de forma progressiva em abas, nunca em modal único gigantesco.

### 4. QA

Cenários versionados, pré-condições, passos, resultado esperado, evidência, execução e associação à revisão/commit. Diferenciar erro encontrado no teste de nova demanda corretiva.

## Métricas e relatórios

- Lead time e cycle time (intervalos explicitamente definidos).
- Tempo em cada estado, fila QA e gargalos por esteira.
- First-pass yield (aprovado no primeiro ciclo / enviados a QA) por período.
- Retrabalho por demanda e por ciclo, sem dupla contagem.
- Percentual de itens da especificação com entrega vinculada, com execução iniciada e homologada.
- Entregas por colaborador **com contexto de tamanho, complexidade, dependências e qualidade**; não converter contagem simples em avaliação automática de pessoas.
- Planejado vs. concluído por iniciativa, produto, ambiente, prioridade e responsável.

## Importação de referência legada

Implementar **preview de importação** e vinculação idempotente:

1. Ler um catálogo completo de itens de referência (quando disponível).
2. Ler um segundo arquivo opcional com propostas/atribuições de um subconjunto.
3. Cruzar registros por identificador estável da fonte; jamais presumir que as duas listas tenham o mesmo tamanho.
4. Gerar proposta de mapeamento e mostrar `novos`, `alterados`, `iguais`, `não mapeados` e `conflitos`.
5. Guardar `source_system`, `external_id`, `source_version`, `imported_at`, `content_hash` e referências de origem.
6. Fazer a confirmação por usuário autorizado; execução de importação atômica via UoW.
7. Nunca sobrescrever valores mais recentes criados pelos usuários sem confirmação explícita.
8. Nunca usar o estado de planejamento externo como se fosse status atual da execução.

**Segurança:** arquivos de referência internos não entram em commit, seed público, build Docker, portfólio ou telemetria de terceiros. Adotar RBAC antes de importar informações reais em ambiente compartilhado.

## Evolução para IA/MCP

Ferramentas de leitura de negócio, não SQL arbitrário:

- `get_initiative_progress`
- `get_specification_item`
- `get_related_work_items`
- `get_qa_coverage`
- `list_blocked_deliveries`
- `get_authorized_person_metrics`

Ferramentas sempre usam os casos de uso de leitura, identidade, permissões e auditoria. O chat de ticket pode recuperar descrição, revisões, critérios, decisões e QA; geração de propostas e cenários exige revisão humana antes de persistir.

## Ordem sugerida de implementação

1. Entidades `Initiative`, `SpecificationItem`, `WorkItemLink` e `ExternalIssueRef`.
2. `ImportPreview` para fontes com versionamento e confirmação.
3. Frontend com Workspace e abas, reaproveitando design do MVP.
4. Métricas por iniciativa/produto/colaborador.
5. QA e controle de revisões mais profundo.
6. MCP + LLM e base de conhecimento com permissões.
