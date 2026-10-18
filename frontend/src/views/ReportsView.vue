<script setup lang="ts">
import AppButton from '../components/AppButton.vue'
import DataTable from '../components/primevue/DataTable'
import Message from '../components/primevue/Message'
import {
  ArrowDownIcon,
  ArrowDownTrayIcon,
  ArrowTrendingUpIcon,
  ArrowUpIcon,
  ArrowsUpDownIcon,
  ChartBarIcon,
  ClockIcon,
  PresentationChartLineIcon,
  ShieldCheckIcon,
  TableCellsIcon,
} from '@heroicons/vue/24/outline'
import { computed, watch } from 'vue'
import { useRoute, useRouter, type LocationQueryRaw } from 'vue-router'
import Tabs from 'primevue/tabs'
import TabList from 'primevue/tablist'
import Tab from 'primevue/tab'
import TabPanels from 'primevue/tabpanels'
import TabPanel from 'primevue/tabpanel'
import { SparklesIcon, UsersIcon } from '@heroicons/vue/24/outline'
import Column from 'primevue/column'
import ProgressBar from 'primevue/progressbar'
import { useWorkspaceStore } from '../stores/workspace'
import { useCatalogStore } from '../stores/catalog'
import { percent } from '../utils/format'
import AIReports from '../components/AIReports.vue'
import ReportChart from '../components/ReportChart.vue'
import KindDistributionChart from '../components/KindDistributionChart.vue'
import { api } from '../api'
import { ref } from 'vue'
const store = useWorkspaceStore(),
  catalog = useCatalogStore()
const exportError = ref('')
const route = useRoute(),
  router = useRouter()
const tabs = [
  { value: 'overview', label: 'Visão geral', icon: ChartBarIcon },
  { value: 'flow', label: 'Fluxo e qualidade', icon: ArrowTrendingUpIcon },
  { value: 'team', label: 'OS e equipe', icon: UsersIcon },
  { value: 'ai', label: 'Relatórios com IA', icon: SparklesIcon },
]
const activeTab = computed({
  get: () =>
    route.query.report_id || route.query.demand_id
      ? 'ai'
      : tabs.some((tab) => tab.value === route.query.tab)
        ? String(route.query.tab)
        : 'overview',
  set: (value: string | number) => {
    const query: LocationQueryRaw = { ...route.query, tab: String(value) }
    if (value !== 'ai') {
      delete query.report_id
      delete query.demand_id
    }
    void router.replace({ query })
  },
})
watch(
  () => route.query.tab,
  () => window.scrollTo({ top: 0, behavior: 'smooth' }),
)
const statusChart = computed(() => ({
  title: 'Demandas por status',
  kind: 'bar' as const,
  points: [
    ['backlog', 'Backlog'],
    ['in_progress', 'Em desenvolvimento'],
    ['in_qa', 'Em QA'],
    ['changes_requested', 'Ajustes solicitados'],
    ['done', 'Concluídas'],
  ].map(([key, label]) => ({ label, value: store.metrics.by_status[key] || 0 })),
}))
const priorityChart = computed(() => ({
  title: 'Demandas por prioridade',
  kind: 'bar' as const,
  points: [
    ['critical', 'Crítica'],
    ['high', 'Alta'],
    ['medium', 'Média'],
    ['low', 'Baixa'],
  ].map(([key, label]) => ({ label, value: store.metrics.by_priority[key] || 0 })),
}))
const qaChart = computed(() => ({
  title: 'Trabalho em QA e ajustes',
  kind: 'bar' as const,
  points: [
    { label: 'Em QA', value: store.metrics.in_qa },
    { label: 'Ajustes solicitados', value: store.metrics.returned },
  ],
}))
const deadlineChart = computed(() => ({
  title: 'Demandas abertas que precisam de atenção',
  kind: 'bar' as const,
  points: [
    { label: 'Prazo vencido', value: store.metrics.overdue },
    { label: 'Sem responsável', value: store.metrics.unassigned },
  ],
}))
const ownerRows = computed(() => {
  const groups = new Map<string, typeof store.items>()
  for (const item of store.items) {
    if (!item.assignee) continue
    const key = item.assignee_id || `legacy:${item.assignee}`
    groups.set(key, [...(groups.get(key) || []), item])
  }
  return [...groups.entries()].map(([id, items]) => ({
    id,
    name: catalog.memberOptions.find((option) => option.id === id)?.label || items[0].assignee,
    total: items.length,
    done: items.filter((item) => item.status === 'done').length,
    qa: items.filter((item) => item.status === 'in_qa').length,
    returned: items.reduce((sum, item) => sum + item.return_count, 0),
  }))
})
async function exportCSV() {
  exportError.value = ''
  try {
    await api.exportCSV()
  } catch (err) {
    exportError.value = (err as Error).message
  }
}
</script>
<template>
  <div class="page-heading">
    <div>
      <div class="eyebrow">
        <span class="eyebrow-line"></span>
        INTELIGÊNCIA OPERACIONAL
      </div>
      <h1>Relatórios<span class="title-period">.</span></h1>
      <p>Dados do trabalho realizado, resultados de QA e evolução das entregas.</p>
    </div>
    <div class="heading-actions">
      <AppButton
        label="Exportar CSV"
        :icon="ArrowDownTrayIcon"
        @click="exportCSV"
      />
    </div>
  </div>
  <div class="reports-banner">
    <div class="reports-banner-icon">
      <ChartBarIcon style="width: 26px; height: 26px; flex-shrink: 0" />
    </div>
    <div>
      <h2>Transforme dados em decisões.</h2>
      <p>Indicadores calculados sobre entregas reais do fluxo, sem duplicar OS e seus itens.</p>
    </div>
    <AppButton
      variant="secondary"
      label="Exportar demandas"
      :icon="TableCellsIcon"
      @click="exportCSV"
    />
  </div>
  <Message
    v-if="exportError"
    severity="error"
    :closable="false"
    class="mb-5"
  >
    {{ exportError }}
  </Message>
  <Tabs
    v-model:value="activeTab"
    class="reports-tabs"
  >
    <TabList aria-label="Seções dos relatórios">
      <Tab
        v-for="tab in tabs"
        :key="tab.value"
        :value="tab.value"
      >
        <component
          :is="tab.icon"
          class="app-icon"
          aria-hidden="true"
        />
        <span>{{ tab.label }}</span>
      </Tab>
    </TabList>
    <TabPanels class="!bg-transparent !px-0 !pt-6">
      <TabPanel value="overview">
        <div class="reports-tab-heading">
          <div>
            <span class="overline">PANORAMA DO WORKSPACE</span>
            <h2>Onde estamos agora</h2>
            <p>Volume e distribuição das demandas, calculados diretamente sobre o fluxo.</p>
          </div>
        </div>
        <div class="report-metrics">
          <div class="report-metric">
            <span>
              <ClockIcon style="width: 18px; height: 18px; flex-shrink: 0" />
              Lead time médio
            </span>
            <strong>
              {{ (store.metrics.average_lead_hours / 24).toFixed(1) }}
              <small>dias</small>
            </strong>
            <p>Da criação à conclusão</p>
          </div>
          <div class="report-metric">
            <span>
              <ShieldCheckIcon style="width: 18px; height: 18px; flex-shrink: 0" />
              Aprovação sem devolução
            </span>
            <strong>
              {{ store.metrics.first_pass_percent.toFixed(0) }}
              <small>%</small>
            </strong>
            <p>Das entregas concluídas</p>
          </div>
          <div class="report-metric">
            <span>
              <PresentationChartLineIcon style="width: 18px; height: 18px; flex-shrink: 0" />
              Devoluções
            </span>
            <strong>{{ store.metrics.returned }}</strong>
            <p>Entregas em ajuste no momento</p>
          </div>
          <div class="report-metric">
            <span>
              <ArrowTrendingUpIcon style="width: 18px; height: 18px; flex-shrink: 0" />
              Entregues
            </span>
            <strong>{{ store.metrics.delivered }}</strong>
            <p>Demandas concluídas</p>
          </div>
        </div>
        <div class="reports-grid">
          <section class="surface-card reports-panel">
            <ReportChart
              :chart="statusChart"
              horizontal
              :localize-dates="false"
              :colors="['#94a3b8', '#8b5cf6', '#38bdf8', '#f59e0b', '#10b981']"
            />
          </section>
          <section class="surface-card reports-panel">
            <ReportChart
              :chart="priorityChart"
              :localize-dates="false"
              :colors="['#f43f5e', '#f59e0b', '#8b5cf6', '#94a3b8']"
            />
          </section>
        </div>
        <section class="surface-card reports-panel mt-6">
          <ReportChart
            :chart="{
              title: 'Entregas por dia · últimos 30 dias (UTC)',
              kind: 'line',
              points: store.metrics.daily_deliveries.map((d) => ({
                label: d.date,
                value: d.count,
              })),
            }"
          />
        </section>
      </TabPanel>
      <TabPanel value="flow">
        <div class="reports-tab-heading">
          <div>
            <span class="overline">FLUXO E QUALIDADE</span>
            <h2>Tempo, prazo e retrabalho</h2>
            <p>
              Considere o tamanho da amostra e os dados disponíveis ao interpretar os indicadores.
            </p>
          </div>
        </div>
        <div class="report-metrics">
          <div class="report-metric">
            <span>Prazo vencido</span>
            <strong>{{ store.metrics.overdue }}</strong>
            <p>Demandas abertas com prazo anterior a hoje</p>
          </div>
          <div class="report-metric">
            <span>Entregas no prazo</span>
            <strong>
              {{
                store.metrics.deliveries_with_due_date
                  ? store.metrics.on_time_percent.toFixed(0)
                  : '—'
              }}
              <small v-if="store.metrics.deliveries_with_due_date">%</small>
            </strong>
            <p>{{ store.metrics.deliveries_with_due_date }} conclusões com prazo informado</p>
          </div>
          <div class="report-metric">
            <span>Retrabalho</span>
            <strong>
              {{ store.metrics.rework_percent.toFixed(0) }}
              <small>%</small>
            </strong>
            <p>Demandas com ao menos uma devolução · {{ store.metrics.total_returns }} retornos</p>
          </div>
          <div class="report-metric">
            <span>Entregas em 30 dias</span>
            <strong>{{ store.metrics.delivered_30_days }}</strong>
            <p>Concluídas nos últimos 30 dias</p>
          </div>
        </div>
        <div class="report-metrics">
          <div class="report-metric">
            <span>Lead time mediano</span>
            <strong>
              {{ (store.metrics.median_lead_hours / 24).toFixed(1) }}
              <small>dias</small>
            </strong>
            <p>Mediana das conclusões com data</p>
          </div>
          <div class="report-metric">
            <span>Lead time P85</span>
            <strong>
              {{ (store.metrics.p85_lead_hours / 24).toFixed(1) }}
              <small>dias</small>
            </strong>
            <p>85% das entregas levaram até esse tempo</p>
          </div>
          <div class="report-metric">
            <span>Idade do trabalho aberto</span>
            <strong>
              {{ (store.metrics.average_active_age_hours / 24).toFixed(1) }}
              <small>dias</small>
            </strong>
            <p>Tempo médio desde a criação</p>
          </div>
          <div class="report-metric">
            <span>Sem responsável</span>
            <strong>{{ store.metrics.unassigned }}</strong>
            <p>Demandas abertas sem atribuição</p>
          </div>
        </div>
        <div class="reports-grid">
          <section class="surface-card reports-panel">
            <ReportChart
              :chart="qaChart"
              :localize-dates="false"
              :colors="['#38bdf8', '#f59e0b']"
            />
            <p class="card-footnote">
              Quantidade atual em cada etapa; não representa resultados de testes.
            </p>
          </section>
          <section class="surface-card reports-panel">
            <ReportChart
              :chart="deadlineChart"
              :localize-dates="false"
              :colors="['#f43f5e', '#8b5cf6']"
            />
            <p class="card-footnote">
              Uma demanda pode aparecer nos dois grupos. Os valores não devem ser somados.
            </p>
          </section>
        </div>
      </TabPanel>
      <TabPanel value="team">
        <div class="reports-tab-heading">
          <div>
            <span class="overline">CONTEXTO DAS ENTREGAS</span>
            <h2>OS, esteiras e equipe</h2>
            <p>Acompanhe a distribuição do trabalho e o progresso das OS.</p>
          </div>
        </div>
        <div class="reports-grid">
          <section class="surface-card reports-panel">
            <div class="card-heading">
              <div>
                <span class="overline">CATEGORIAS</span>
                <h2>Volume por esteira</h2>
              </div>
            </div>
            <KindDistributionChart
              :counts="store.metrics.by_kind"
              title="Volume por esteira"
            />
          </section>
          <section class="surface-card reports-panel">
            <div class="card-heading">
              <div>
                <span class="overline">ACOMPANHAMENTO</span>
                <h2>Progresso por OS</h2>
              </div>
            </div>
            <div
              v-for="pkg in store.packages"
              :key="pkg.id"
              class="report-package"
            >
              <div class="flex items-center justify-between gap-2">
                <strong>{{ pkg.reference || pkg.title }}</strong>
                <span>{{ pkg.done }} / {{ pkg.total }}</span>
              </div>
              <small>{{ pkg.title }}</small>
              <ProgressBar
                :value="percent(pkg.done, pkg.total)"
                :show-value="false"
                class="package-progress mt-2"
                :aria-label="`Progresso da OS ${pkg.reference || pkg.title}`"
              />
            </div>
            <div
              v-if="!store.packages.length"
              class="empty-inline"
            >
              Ainda não existem OS para acompanhar.
            </div>
          </section>
        </div>
        <section class="surface-card reports-people">
          <div class="card-heading">
            <div>
              <span class="overline">VISIBILIDADE DA SQUAD</span>
              <h2>Entregas por responsável</h2>
              <p>
                Volume, conclusão e devoluções para contextualizar o fluxo, não ranquear pessoas.
              </p>
            </div>
          </div>
          <DataTable
            :value="ownerRows"
            data-key="id"
            size="small"
            scrollable
          >
            <Column
              field="name"
              header="Responsável"
              sortable
            >
              <template #sorticon="{ sortOrder, class: iconClass }">
                <component
                  :is="
                    sortOrder === 1
                      ? ArrowUpIcon
                      : sortOrder === -1
                        ? ArrowDownIcon
                        : ArrowsUpDownIcon
                  "
                  :class="iconClass"
                  class="app-icon"
                />
              </template>
            </Column>
            <Column
              field="total"
              header="Itens"
              sortable
            >
              <template #sorticon="{ sortOrder, class: iconClass }">
                <component
                  :is="
                    sortOrder === 1
                      ? ArrowUpIcon
                      : sortOrder === -1
                        ? ArrowDownIcon
                        : ArrowsUpDownIcon
                  "
                  :class="iconClass"
                  class="app-icon"
                />
              </template>
            </Column>
            <Column
              field="done"
              header="Concluídos"
              sortable
            >
              <template #sorticon="{ sortOrder, class: iconClass }">
                <component
                  :is="
                    sortOrder === 1
                      ? ArrowUpIcon
                      : sortOrder === -1
                        ? ArrowDownIcon
                        : ArrowsUpDownIcon
                  "
                  :class="iconClass"
                  class="app-icon"
                />
              </template>
            </Column>
            <Column
              field="qa"
              header="Em QA"
              sortable
            >
              <template #sorticon="{ sortOrder, class: iconClass }">
                <component
                  :is="
                    sortOrder === 1
                      ? ArrowUpIcon
                      : sortOrder === -1
                        ? ArrowDownIcon
                        : ArrowsUpDownIcon
                  "
                  :class="iconClass"
                  class="app-icon"
                />
              </template>
            </Column>
            <Column
              field="returned"
              header="Devoluções"
              sortable
            >
              <template #sorticon="{ sortOrder, class: iconClass }">
                <component
                  :is="
                    sortOrder === 1
                      ? ArrowUpIcon
                      : sortOrder === -1
                        ? ArrowDownIcon
                        : ArrowsUpDownIcon
                  "
                  :class="iconClass"
                  class="app-icon"
                />
              </template>
            </Column>
            <template #empty>
              <div class="empty-inline">
                Atribua responsáveis aos tickets para visualizar a distribuição.
              </div>
            </template>
          </DataTable>
          <div class="card-footnote">
            Complexidade, dependências e natureza das demandas devem ser consideradas na
            interpretação dos indicadores.
          </div>
        </section>
      </TabPanel>
      <TabPanel value="ai"><AIReports /></TabPanel>
    </TabPanels>
  </Tabs>
</template>

<style scoped>
.reports-tabs {
  min-width: 0;
}
.reports-tabs :deep(.p-tablist-content) {
  overflow-x: auto;
}
.reports-tabs :deep(.p-tab) {
  display: inline-flex;
  align-items: center;
  gap: 0.6rem;
  white-space: nowrap;
}
.reports-tabs :deep(.p-tabpanels) {
  min-width: 0;
}
.reports-tab-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1.5rem;
}
.reports-tab-heading h2 {
  font-size: 1.45rem;
  margin: 0.4rem 0;
}
.reports-tab-heading p {
  color: var(--p-text-muted-color);
  font-size: 0.9rem;
}
.reports-grid {
  margin-top: 1.5rem;
}
@media (max-width: 600px) {
  .reports-tabs :deep(.p-tab) {
    padding: 1rem 0.8rem;
    font-size: 0.85rem;
  }
  .reports-tab-heading h2 {
    font-size: 1.25rem;
  }
}
</style>
