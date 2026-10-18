# Frontend SPA — SquadFlow v0.4

## Objetivo

A interface é um workspace com **esteiras cadastradas pelo admin**, inicialmente Melhoria, Corretiva, Teste e Reescrita. Não há uma tela específica para NT: Notas Técnicas são apenas iniciativas, e uma OS pode ter vários itens de trabalho. Tickets corretivos podem existir sem OS.

## Organização

```text
frontend/src/
├── App.vue                      # composição da SPA
├── main.ts                      # plugins: Pinia, PrimeVue, Toast, Router
├── router/index.ts              # rotas SPA
├── stores/workspace.ts          # estado e mutações na API
├── config/catalog.ts           # rótulos de status/prioridade e consulta de esteiras
├── utils/format.ts             # formatadores puros
├── components/
│   ├── AppShell.vue             # navegação global, layout responsivo
│   ├── MemberSelect.vue         # seleção de usuários cadastrados
│   ├── NotificationCenter.vue   # sino e caixa privada de notificações para membros
│   ├── AppButton.vue            # ações padronizadas sobre Button do PrimeVue
│   ├── CreateDemandDialog.vue   # criação de itens com vínculo opcional
│   ├── CreatePackageDialog.vue  # cadastro de OS e primeiro item
│   ├── TaskDrawer.vue           # detalhe, transição, edição e QA
│   ├── DemandCard.vue           # cartão reutilizável
│   ├── PasswordForm.vue         # troca de senha com confirmação
│   └── StatCard.vue
└── views/
    ├── DashboardView.vue
    ├── TicketsView.vue
    ├── BoardView.vue
    ├── ReleasesView.vue
    ├── PackagesView.vue
    ├── QaView.vue
    ├── AccountView.vue
    ├── ResetPasswordView.vue
    └── ReportsView.vue
```

## Decisões

- Vue Router com history mode e fallback Nginx (`try_files ... /index.html`) para abrir qualquer rota diretamente.
- Pinia centraliza o estado compartilhado da API; o Kanban mantém localmente suas páginas de cartões e filtros na URL. Não há dados ou contagens fictícias codificados nas telas.
- PrimeVue 4 para formulários, tabela, diálogo, drawer, notificações, gráficos (`Chart` com Chart.js) e progresso de OS (`ProgressBar`); Tailwind 4 para o layout e CSS próprio para identidade visual. Tema Aura com cor primária violeta.
- Ícones SVG da biblioteca oficial [Heroicons para Vue](https://github.com/tailwindlabs/heroicons/tree/master/vue), variante `24/outline`, em toda a interface. `components/primevue/` fornece os ícones internos pelos slots públicos dos componentes (carregamento, paginação, calendário, seleção e fechamento), preservando os slots das telas. Importe esses componentes individualmente para manter o carregamento por rota. A classe `app-icon` padroniza dimensões e cor herdada; indicadores de carregamento respeitam movimento reduzido. Lucide e a fonte PrimeIcons foram removidos.
- Identidade inspirada no [TaskWhiz](https://dribbble.com/shots/20769637-TaskWhiz-Task-Management-Kanban-Dashboard): superfícies em carvão, navegação lateral com opção compacta e cartões com etiquetas e responsáveis. O menu inicia expandido, com ícones e nomes das telas; o botão de recolher/expandir salva a preferência local. No celular, os nomes ficam sempre visíveis ao abrir o menu. O tema escuro é o padrão inicial; a preferência salva por claro/escuro é respeitada. `styles/workspace.css` concentra essa apresentação e adapta a navegação para celular.
- `styles/atmosphere.css` adiciona glass com fundo fixo de luzes difusas em violeta, teal e rosa. Cards e painéis têm transparência perceptível, reflexo diagonal, borda luminosa e desfoque do fundo, inclusive sem hover. O Kanban usa blur de 10 px e os painéis de 14 px, reduzidos para 8 px no celular; campos e formulários preservam suas próprias superfícies para leitura. Os reflexos são estáticos e o efeito adapta-se aos dois temas, com alternativa sólida quando não há suporte ou a pessoa prefere menos transparência. A camada decorativa ignora cliques e leitores de tela, sem interferir no arraste e nos overlays.
- O quadro mantém as cinco etapas, oferece atalhos para demandas, versões e relatórios e deixa busca e versão sempre visíveis. O botão **Filtros** expande arquivo, esteira e responsável; as colunas têm rolagem própria em desktop e navegação horizontal em telas menores.
- O sidebar anima largura, textos e perfil em 240 ms. A troca de telas usa fade com deslocamento discreto (100 ms de saída e 180 ms de entrada); login e backdrop mobile usam fade. As transições de tela usam o caminho da rota como chave, preservando a tela e o foco ao mudar filtros na URL. `prefers-reduced-motion` desativa essas animações.
- Dicas usam a diretiva global `v-tooltip` do PrimeVue. `config/tooltip.ts` padroniza atraso e conteúdo textual; `config/primevue.ts` define cores, espaçamento e cantos para os dois temas. `styles/workspace.css` ajusta tipografia, quebra de texto, largura no celular e movimento reduzido. O menu recolhido, controles por ícone, avatares e prazos usam essas dicas, mantendo seus nomes acessíveis.
- No cadastro de usuários, `PasswordInput` habilita o feedback do Password do PrimeVue: indicador Fraca/Média/Forte e sugestões atualizadas durante a digitação. O texto usa caracteres e o mínimo de 12, com orientações de maiúsculas, minúsculas, números e símbolos. A força é uma indicação visual; a API mantém sua validação. O login usa o mesmo componente com feedback desativado. O cadastro verifica comprimento Unicode e o limite técnico antes de enviar, com mensagens sem termos de armazenamento.
- O Kanban consulta `/api/v1/board`, carrega 20 resumos por coluna e oculta versões arquivadas por padrão. O workspace evita baixar a lista completa de demandas nessa rota.
- A tela `/releases?release_id=<id>` cadastra versões, consulta entregas e permite arquivar/restaurar com histórico preservado.
- Kanban é uma projeção dos itens de trabalho. Os cartões podem ser arrastados entre as etapas permitidas, com prévia, destaque de destinos, rolagem nas bordas e cancelamento por Escape. No celular, arraste pela alça para preservar a rolagem normal do cartão. Clicar na alça ou acioná-la pelo teclado abre um Dialog do PrimeVue com as etapas disponíveis. A devolução exige motivo; a conclusão exige aprovação de QA. A API registra autor e histórico, e o quadro só muda após a confirmação do servidor. O drawer continua oferecendo as mesmas movimentações. A posição dentro de uma coluna segue a ordenação da API.
- Dashboard e relatórios baseiam-se em métricas da API e agrupamentos dos itens, sem duplicar OS.
- Gráficos mantêm bordas na cor original e preenchimentos translúcidos, com a mesma paleta nos dois temas. Barras usam contorno de 2 px; linhas usam traço de 3 px e área suave até o eixo zero. `utils/chartStyle.ts` centraliza a transparência, inclusive para o gráfico circular e os relatórios gerados.
- O módulo de OS usa query `?os=<id>` e oferece URL profunda para o agrupamento selecionado.
- Login por cookies HttpOnly, gestão de usuários exclusiva do admin, geração de relatórios com IA e atualização por WebSocket. Veja [AUTH-AI-REPORTS.md](AUTH-AI-REPORTS.md).
- **Minha conta** (`/account`) permite trocar a própria senha com a senha atual. Em **Usuários**, o admin pode alterá-la diretamente ou gerar um link interno válido por 30 minutos. A tela pública `/reset-password`, acessível pelo link ou por **Esqueci minha senha**, recebe o código e confirma a nova senha. Os formulários compartilham o Password do PrimeVue. Veja [PASSWORD-RECOVERY.md](PASSWORD-RECOVERY.md).

## Notificações

`NotificationCenter` substitui o indicador “Ao vivo” por um sino com contador para membros; o admin não renderiza a central nem consulta suas rotas. O painel usa Drawer, SelectButton, Heroicons e AppButton, com filtros, paginação, marcação de leitura e abertura da demanda. O estado fica no banco e as consultas são privadas por sessão. WebSocket continua funcionando e atualiza o contador/lista; há nova consulta ao voltar à aba e fallback de 30 segundos quando desconectado. Veja eventos, API e migração em [NOTIFICATIONS.md](NOTIFICATIONS.md).

## Filtros da Central de QA

A rota `/qa` combina busca por título, código da demanda, OS, produto, descrição e evento com filtros por esteira, responsável QA, desenvolvedor e versão. A busca ignora acentos e maiúsculas; cada termo deve aparecer nos dados da demanda. Responsáveis são comparados por ID, inclusive no atalho **Minhas revisões**, e há opções para demandas sem responsável ou versão. Versões arquivadas continuam disponíveis para consulta.

Os contadores das três etapas refletem a combinação de filtros. Os filtros ativos podem ser removidos individualmente ou limpos em conjunto sem mudar a etapa selecionada. Abrir uma demanda e retornar mantém a seleção. O painel se adapta ao celular e diferencia uma etapa vazia de uma busca sem resultados.

## Esteiras e responsáveis

`stores/catalog.ts` carrega esteiras e usuários cadastrados. A rota `/workstreams`, restrita ao admin, gerencia nome, cor e ativação das esteiras. Formulários de demandas e OS usam somente esteiras ativas e o `MemberSelect` para atribuição por ID. Filtros, etiquetas e gráficos consultam os nomes e cores do catálogo, incluindo novas categorias. Movimentações e testes têm autoria definida pela sessão no servidor. Veja API, migração e seed em [WORKSTREAMS-AND-MEMBERS.md](WORKSTREAMS-AND-MEMBERS.md).

## Botões

Use `components/AppButton.vue` para ações da interface. Ele encapsula o Button do PrimeVue e centraliza aparência e estados em `styles/buttons.css`:

- `variant`: `primary` (padrão), `secondary` (contornado), `ghost` (ação discreta), `danger` (ação de risco) ou `link` (abrir conteúdo relacionado).
- `size`: `small` (32 px), `normal` (40 px, padrão) ou `large` (44 px). Botões só com ícone ficam quadrados automaticamente e precisam de `aria-label` e tooltip.
- `icon`: componente Heroicons; o slot `icon` continua disponível para personalização.
- `loading` mostra o indicador e bloqueia a ação; `disabled` desabilita sem carregamento. `fluid` ocupa a largura disponível.
- O tipo padrão é `button`. Informe `type="submit"` para enviar formulários; `form`, atributos ARIA, `data-testid`, eventos e `v-tooltip` são encaminhados ao botão do PrimeVue.

```vue
<AppButton label="Nova entrega" :icon="PlusIcon" @click="openCreate" />
<AppButton label="Cancelar" variant="ghost" @click="close" />
<AppButton label="Salvar" type="submit" form="delivery-form" :loading="busy" />
<AppButton :icon="PencilSquareIcon" variant="ghost" size="small"
  aria-label="Editar entrega" v-tooltip.top="tooltip('Editar entrega')" @click="edit" />
```

Evite redefinir altura, padding, fonte ou cores por tela. CSS local deve cuidar do posicionamento. `components/primevue/SurfaceButton.ts` fica reservado aos cards e linhas clicáveis do Kanban e das OS, que têm conteúdo e dimensões próprios.

## Checklist técnico

1. `npm ci && npm run check && npm run build` no frontend.
2. `docker compose up -d --build` na raiz do projeto.
3. Conferir dashboard e rotas `/tickets`, `/board`, `/releases`, `/packages`, `/qa`, `/reports`.
4. Cadastrar uma OS com primeiro item, adicionar outro com responsável diferente e abrir os detalhes.
5. Cadastrar corretiva sem OS; iniciar, enviar ao QA, reprovar, devolver, corrigir e aprovar no segundo ciclo.
6. `cd frontend && npx playwright install chromium && npm run test:e2e`.

## Segurança / desenvolvimento futuro

As rotas do workspace exigem login; login e redefinição de senha são públicos. O backend aplica autenticação e restringe o cadastro de usuários ao único administrador. Transições e testes usam a identidade autenticada. Tokens de sessão ficam em cookies HttpOnly, com proteção de origem e rotação de refresh. O token de recuperação fica somente em memória e é consumido uma vez. A configuração e os limites de implantação estão em [AUTH-AI-REPORTS.md](AUTH-AI-REPORTS.md).

**Fontes:** Google Fonts são carregadas de CDN público no exemplo. Para uma instalação sem acesso externo, remova essas tags ou configure o uso corporativo de fontes licenciadas/autorizadas.

## Relatórios assíncronos em HTML

A solicitação responde `202` e fica em `stores/reportJobs.ts`, que recupera o estado do servidor e acompanha a fila globalmente enquanto há sessão. `ReportJobsPanel` mostra etapas, tentativas e falhas; sair da rota ou recarregar não cancela o trabalho. O frontend solicita por HTTP, sem depender de uma conexão WebSocket longa.

`ReportRenderer` mostra `content.html` em iframe com sandbox vazio, permite apresentar em dialog PrimeVue maximizável e baixar HTML ou JSON. Os documentos são independentes de Vue e usam CSS/SVG internos. Relatórios antigos mantêm seus componentes de visualização. `DashboardReport` abre a última análise pessoal dentro do dashboard. Veja [arquitetura e execução](ASYNC-HTML-REPORTS.md).

### Organização dos relatórios

A tela usa quatro abas PrimeVue: Visão geral, Fluxo e qualidade, OS e equipe, Relatórios com IA. A URL mantém a seção selecionada em `tab`; links de demandas ou relatórios abrem diretamente a aba de IA. Os painéis preservam o formulário ao alternar e a fila continua independente da aba ativa. Novos gráficos padrão usam métricas reais, PrimeVue Chart e a paleta com bordas sólidas e preenchimento translúcido.

### Qualidade e formatação

`eslint.config.js` combina as regras recomendadas de JavaScript, TypeScript e Vue 3, com ambientes de browser para a aplicação e Node para testes/configurações. `eslint-config-prettier` remove conflitos entre lint e formatação. `npm run lint` falha com qualquer aviso; `npm run lint:fix` aplica correções automáticas.

`.prettierrc.json` usa dois espaços, aspas simples, sem ponto e vírgula, largura de 100 caracteres e um atributo Vue por linha. `npm run format` aplica o padrão e `npm run format:check` verifica sem escrever. `npm run check` reúne lint, formatação e TypeScript. O CI executa essas verificações, e o VS Code usa o Prettier local ao salvar.

### Validação do login

A tela de login usa `@primevue/forms` com resolver Zod, definido em `validation/login.ts`. Os inputs são registrados por `name`, as mensagens ficam associadas aos campos por `aria-describedby` e o envio valida o formulário antes de chamar a API. O cliente preserva o código do erro HTTP em `APIError`, e credenciais incorretas aparecem como mensagem do formulário. Veja [o contrato de erros](API-ERRORS.md).
