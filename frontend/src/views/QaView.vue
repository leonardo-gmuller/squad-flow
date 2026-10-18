<script setup lang="ts">
import {
  ArrowUturnLeftIcon,
  CheckCircleIcon,
  ClipboardDocumentCheckIcon,
  FunnelIcon,
  MagnifyingGlassIcon,
  ShieldCheckIcon,
  UserCircleIcon,
  XMarkIcon,
} from '@heroicons/vue/24/outline'
import { computed, reactive, ref } from 'vue'
import Tabs from 'primevue/tabs'
import TabList from 'primevue/tablist'
import Tab from 'primevue/tab'
import TabPanels from 'primevue/tabpanels'
import TabPanel from 'primevue/tabpanel'
import Badge from 'primevue/badge'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import InputText from 'primevue/inputtext'
import AppButton from '../components/AppButton.vue'
import Select from '../components/primevue/Select'
import FormField from '../components/FormField.vue'
import DemandCard from '../components/DemandCard.vue'
import { useWorkspaceStore } from '../stores/workspace'
import { useCatalogStore } from '../stores/catalog'
import { useAuthStore } from '../stores/auth'
import { demandCode } from '../utils/format'
import { tooltip } from '../config/tooltip'

type QAStage = 'in_qa' | 'changes_requested' | 'done'
const store = useWorkspaceStore(),
  catalog = useCatalogStore(),
  auth = useAuthStore()
const tab = ref<QAStage>('in_qa')
const tabs = [
  { value: 'in_qa' as const, label: 'Aguardando QA', icon: ClipboardDocumentCheckIcon },
  { value: 'changes_requested' as const, label: 'Devolvidas', icon: ArrowUturnLeftIcon },
  { value: 'done' as const, label: 'Concluídas', icon: CheckCircleIcon },
]
const filters = reactive({ q: '', kind: '', reviewer: '', owner: '', release: '' })
type FilterKey = keyof typeof filters
const kindOptions = computed(() => [
  { value: '', label: 'Todas as esteiras' },
  ...catalog.kindOptions,
])
const peopleOptions = computed(() =>
  catalog.memberOptions
    .filter((person) => person.id)
    .map((person) => ({ value: person.id, label: person.label })),
)
const reviewerOptions = computed(() => [
  { value: '', label: 'Todos os responsáveis QA' },
  { value: 'none', label: 'Sem responsável QA' },
  ...peopleOptions.value,
])
const ownerOptions = computed(() => [
  { value: '', label: 'Todos os desenvolvedores' },
  { value: 'none', label: 'Sem desenvolvedor' },
  ...peopleOptions.value,
])
const releaseOptions = computed(() => [
  { value: '', label: 'Todas as versões' },
  { value: 'none', label: 'Sem versão' },
  ...store.releases.map((release) => ({
    value: release.id,
    label: release.name + (release.archived_at ? ' · Arquivada' : ''),
  })),
])
const mine = computed(() => Boolean(auth.user?.id && filters.reviewer === auth.user.id))
function toggleMine() {
  filters.reviewer = mine.value ? '' : auth.user?.id || ''
}
function clear() {
  Object.assign(filters, { q: '', kind: '', reviewer: '', owner: '', release: '' })
}
function normalize(value: string) {
  return value
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .toLocaleLowerCase('pt-BR')
}
const qaItems = computed(() =>
  store.items.filter((item) => tabs.some((stage) => stage.value === item.status)),
)
const filtered = computed(() => {
  const terms = normalize(filters.q).trim().split(/\s+/).filter(Boolean)
  return qaItems.value.filter((item) => {
    if (filters.kind && item.kind !== filters.kind) return false
    if (
      filters.reviewer &&
      (filters.reviewer === 'none'
        ? Boolean(item.reviewer_id)
        : item.reviewer_id !== filters.reviewer)
    )
      return false
    if (
      filters.owner &&
      (filters.owner === 'none' ? Boolean(item.assignee_id) : item.assignee_id !== filters.owner)
    )
      return false
    if (
      filters.release &&
      (filters.release === 'none' ? Boolean(item.release_id) : item.release_id !== filters.release)
    )
      return false
    const text = normalize(
      `${demandCode(item.number)} ${item.number} ${item.title} ${item.reference} ${item.product} ${item.description} ${item.event_codes.join(' ')}`,
    )
    return terms.every((term) => text.includes(term))
  })
})
const groups = computed(() => ({
  in_qa: filtered.value.filter((item) => item.status === 'in_qa'),
  changes_requested: filtered.value.filter((item) => item.status === 'changes_requested'),
  done: filtered.value.filter((item) => item.status === 'done'),
}))
const stageTotal = (stage: QAStage) => qaItems.value.filter((item) => item.status === stage).length
const activeFilters = computed(() => {
  const selected: { key: FilterKey; label: string }[] = []
  if (filters.q.trim()) selected.push({ key: 'q', label: `Busca: ${filters.q.trim()}` })
  const options = [
    { key: 'kind' as const, label: 'Esteira', values: kindOptions.value },
    { key: 'reviewer' as const, label: 'QA', values: reviewerOptions.value },
    { key: 'owner' as const, label: 'Desenvolvedor', values: ownerOptions.value },
    { key: 'release' as const, label: 'Versão', values: releaseOptions.value },
  ]
  for (const option of options) {
    if (filters[option.key])
      selected.push({
        key: option.key,
        label: `${option.label}: ${option.values.find((value) => value.value === filters[option.key])?.label || 'Indisponível'}`,
      })
  }
  return selected
})
</script>

<template>
  <div class="page-heading">
    <div>
      <div class="eyebrow">
        <span class="eyebrow-line"></span>
        QUALIDADE DE SOFTWARE
      </div>
      <h1>Central de QA<span class="title-period">.</span></h1>
      <p>Homologações, evidências e retornos para desenvolvimento em um só lugar.</p>
    </div>
  </div>
  <div class="qa-feature">
    <span class="qa-feature-icon"><ShieldCheckIcon class="app-icon" /></span>
    <div>
      <strong>Qualidade faz parte da entrega.</strong>
      <p>
        Registre cenários, aprove ou reprove dentro de cada ticket e mantenha o histórico de ciclos.
      </p>
    </div>
    <span class="qa-feature-metric">
      {{ store.metrics.in_qa }}
      <small>em homologação</small>
    </span>
  </div>

  <section
    class="surface-card qa-filter-card"
    aria-labelledby="qa-filter-title"
  >
    <div class="qa-filter-heading">
      <h2 id="qa-filter-title">
        <FunnelIcon
          class="app-icon"
          aria-hidden="true"
        />
        Filtrar demandas
        <Badge
          v-if="activeFilters.length"
          :value="activeFilters.length"
          severity="secondary"
        />
      </h2>
      <AppButton
        label="Limpar filtros"
        variant="ghost"
        size="small"
        :icon="XMarkIcon"
        :disabled="!activeFilters.length"
        @click="clear"
      />
    </div>
    <div class="qa-filter-search">
      <IconField>
        <InputIcon><MagnifyingGlassIcon class="app-icon" /></InputIcon>
        <InputText
          v-model="filters.q"
          aria-label="Buscar demandas de QA"
          placeholder="Título, código, OS ou evento..."
          maxlength="300"
          fluid
        />
      </IconField>
      <AppButton
        v-tooltip.top="tooltip('Mostrar demandas em que você é o responsável QA')"
        label="Minhas revisões"
        :variant="mine ? 'secondary' : 'ghost'"
        :icon="UserCircleIcon"
        :aria-pressed="mine"
        :disabled="!auth.user?.id"
        @click="toggleMine"
      />
    </div>
    <div class="qa-filter-fields">
      <FormField
        id="qa-filter-kind"
        label="Esteira"
      >
        <Select
          v-model="filters.kind"
          input-id="qa-filter-kind"
          aria-labelledby="qa-filter-kind-label"
          :options="kindOptions"
          placeholder="Todas as esteiras"
          option-label="label"
          option-value="value"
          filter
          fluid
        />
      </FormField>
      <FormField
        id="qa-filter-reviewer"
        label="Responsável QA"
      >
        <Select
          v-model="filters.reviewer"
          input-id="qa-filter-reviewer"
          aria-labelledby="qa-filter-reviewer-label"
          :options="reviewerOptions"
          placeholder="Todos os responsáveis QA"
          option-label="label"
          option-value="value"
          filter
          fluid
        />
      </FormField>
      <FormField
        id="qa-filter-owner"
        label="Desenvolvedor"
      >
        <Select
          v-model="filters.owner"
          input-id="qa-filter-owner"
          aria-labelledby="qa-filter-owner-label"
          :options="ownerOptions"
          placeholder="Todos os desenvolvedores"
          option-label="label"
          option-value="value"
          filter
          fluid
        />
      </FormField>
      <FormField
        id="qa-filter-release"
        label="Versão"
      >
        <Select
          v-model="filters.release"
          input-id="qa-filter-release"
          aria-labelledby="qa-filter-release-label"
          :options="releaseOptions"
          placeholder="Todas as versões"
          option-label="label"
          option-value="value"
          filter
          fluid
        />
      </FormField>
    </div>
    <div
      v-if="activeFilters.length"
      class="qa-active-filters"
      aria-label="Filtros ativos"
    >
      <AppButton
        v-for="filter in activeFilters"
        :key="filter.key"
        :label="filter.label"
        :aria-label="`Remover filtro ${filter.label}`"
        variant="secondary"
        size="small"
        :icon="XMarkIcon"
        class="qa-filter-chip"
        @click="filters[filter.key] = ''"
      />
    </div>
  </section>

  <Tabs
    v-model:value="tab"
    class="qa-status-tabs"
  >
    <TabList aria-label="Etapas de QA">
      <Tab
        v-for="stage in tabs"
        :key="stage.value"
        :value="stage.value"
      >
        <component
          :is="stage.icon"
          class="app-icon"
          aria-hidden="true"
        />
        <span class="qa-tab-label">{{ stage.label }}</span>
        <Badge
          :value="groups[stage.value].length"
          severity="secondary"
        />
      </Tab>
    </TabList>
    <TabPanels class="!bg-transparent !px-0 !pt-5">
      <TabPanel
        v-for="stage in tabs"
        :key="stage.value"
        :value="stage.value"
      >
        <p
          v-if="store.loading && !store.loaded"
          class="qa-result-summary"
          role="status"
        >
          Carregando demandas de QA…
        </p>
        <p
          v-else
          class="qa-result-summary"
          role="status"
          aria-live="polite"
        >
          <strong>{{ groups[stage.value].length }}</strong>
          {{ groups[stage.value].length === 1 ? 'demanda encontrada' : 'demandas encontradas' }}
          <span v-if="activeFilters.length">de {{ stageTotal(stage.value) }} nesta etapa</span>
        </p>
        <div class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
          <DemandCard
            v-for="item in groups[stage.value]"
            :key="item.id"
            :item="item"
          />
          <div
            v-if="store.loaded && !groups[stage.value].length"
            class="surface-card empty-qa qa-empty-state"
          >
            <ShieldCheckIcon
              class="app-icon"
              aria-hidden="true"
            />
            <h3>
              {{
                activeFilters.length
                  ? 'Nenhuma demanda corresponde aos filtros'
                  : 'Nenhuma demanda nesta etapa'
              }}
            </h3>
            <p>
              {{
                activeFilters.length
                  ? 'Ajuste a busca ou remova algum filtro para ampliar os resultados.'
                  : 'Quando houver movimentações de QA, as demandas aparecerão aqui.'
              }}
            </p>
            <AppButton
              v-if="activeFilters.length"
              label="Limpar filtros"
              variant="secondary"
              size="small"
              :icon="XMarkIcon"
              @click="clear"
            />
          </div>
        </div>
      </TabPanel>
    </TabPanels>
  </Tabs>
</template>

<style scoped>
.qa-feature-icon > svg {
  width: 27px;
  height: 27px;
}
.qa-filter-card {
  padding: 20px;
  margin-bottom: 22px;
}
.qa-filter-heading,
.qa-filter-heading h2,
.qa-filter-search {
  display: flex;
  align-items: center;
  gap: 12px;
}
.qa-filter-heading {
  justify-content: space-between;
  margin-bottom: 18px;
}
.qa-filter-heading h2 {
  font-size: 13px;
  font-weight: 600;
  color: var(--sf-text);
}
.qa-filter-heading h2 > svg {
  color: var(--sf-accent);
}
.qa-filter-search {
  margin-bottom: 18px;
}
.qa-filter-search > .p-iconfield {
  flex: 1;
  min-width: 0;
}
.qa-filter-fields {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
}
.qa-filter-fields :deep(label) {
  color: var(--sf-muted);
  font-size: 11px;
}
.qa-filter-fields :deep(.p-select) {
  font-size: 12px;
  min-width: 0;
}
.qa-active-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding-top: 16px;
  margin-top: 18px;
  border-top: 1px solid var(--sf-border);
}
.qa-filter-chip.app-button {
  max-width: 100%;
  min-width: 0;
  text-align: left;
}
.qa-filter-chip :deep(.p-button-label) {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.qa-status-tabs {
  min-width: 0;
}
.qa-status-tabs :deep(.p-tablist),
.qa-status-tabs :deep(.p-tablist-content) {
  background: transparent;
}
.qa-status-tabs :deep(.p-tablist-tab-list) {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
  background: transparent;
  border: 0;
}
.qa-status-tabs :deep(.p-tablist-active-bar) {
  display: none;
}
.qa-status-tabs :deep(.p-tab) {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 10px;
  min-width: 0;
  padding: 16px;
  border: 1px solid var(--sf-border);
  border-radius: 12px;
  background: var(--sf-soft);
  color: var(--sf-muted);
  font-size: 12px;
}
.qa-status-tabs :deep(.p-tab[aria-selected='true']) {
  border-color: var(--sf-accent);
  background: var(--sf-accent-soft);
  color: var(--sf-accent);
}
.qa-status-tabs :deep(.p-tab:focus-visible) {
  outline: 2px solid var(--sf-accent);
  outline-offset: -4px;
}
.qa-status-tabs :deep(.p-tab .p-badge) {
  margin-left: auto;
  flex-shrink: 0;
}
.qa-tab-label {
  min-width: 0;
  white-space: normal;
}
.qa-result-summary {
  color: var(--sf-muted);
  font-size: 12px;
  line-height: 1.6;
  margin-bottom: 16px;
}
.qa-result-summary strong {
  color: var(--sf-text);
}
.qa-empty-state {
  padding: 44px 24px;
  gap: 12px;
}
.qa-empty-state > svg {
  width: 38px;
  height: 38px;
  color: var(--sf-accent);
}
.qa-empty-state h3 {
  margin: 0;
}
.qa-empty-state p {
  color: var(--sf-muted);
  line-height: 1.7;
}
@media (max-width: 1200px) {
  .qa-filter-fields {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
@media (max-width: 640px) {
  .qa-filter-card {
    padding: 16px;
  }
  .qa-filter-heading {
    align-items: flex-start;
    gap: 8px;
  }
  .qa-filter-heading h2 {
    gap: 8px;
    font-size: 12px;
    flex-wrap: wrap;
  }
  .qa-filter-search {
    flex-direction: column;
    align-items: stretch;
    gap: 10px;
  }
  .qa-filter-search > .app-button {
    align-self: flex-start;
  }
  .qa-filter-fields {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 14px;
  }
  .qa-status-tabs :deep(.p-tablist-tab-list) {
    gap: 8px;
  }
  .qa-status-tabs :deep(.p-tab) {
    flex-direction: column;
    justify-content: center;
    gap: 8px;
    padding: 12px 5px;
    text-align: center;
    font-size: 11px;
  }
  .qa-status-tabs :deep(.p-tab .p-badge) {
    margin-left: 0;
  }
}
@media (max-width: 360px) {
  .qa-filter-fields {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
