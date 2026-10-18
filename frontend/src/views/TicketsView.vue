<script setup lang="ts">
import { useCatalogStore } from '../stores/catalog'
import AppButton from '../components/AppButton.vue'
import Select from '../components/primevue/Select'
import DataTable from '../components/primevue/DataTable'
import {
  ArrowDownIcon,
  ArrowUpIcon,
  ArrowsUpDownIcon,
  FunnelIcon,
  MagnifyingGlassIcon,
  PlusIcon,
} from '@heroicons/vue/24/outline'
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import InputText from 'primevue/inputtext'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import Column from 'primevue/column'
import { useWorkspaceStore } from '../stores/workspace'
import { stages, kindLabel, kindStyle, priorityLabel, statusLabel } from '../config/catalog'
import { dateLabel, demandCode, initials, isLate } from '../utils/format'
import type { Kind, Status } from '../types'
const catalog = useCatalogStore()
const store = useWorkspaceStore(),
  route = useRoute()
const q = ref(String(route.query.q ?? '')),
  kind = ref<Kind | ''>((route.query.kind as Kind) || ''),
  status = ref<Status | ''>(''),
  owner = ref(''),
  release = ref(String(route.query.release_id || ''))
watch(
  () => route.query,
  (query) => {
    q.value = String(query.q ?? '')
    kind.value = (query.kind as Kind) || ''
    release.value = String(query.release_id || '')
  },
)
const filtered = computed(() =>
  store.items.filter(
    (d) =>
      (!release.value ||
        (release.value === 'none' ? !d.release_id : d.release_id === release.value)) &&
      (!kind.value || d.kind === kind.value) &&
      (!status.value || d.status === status.value) &&
      (!owner.value || d.assignee_id === owner.value) &&
      (!q.value ||
        `${d.title} ${d.reference} ${d.product} ${d.event_codes.join(' ')} ${d.description}`
          .toLowerCase()
          .includes(q.value.toLowerCase())),
  ),
)
function clear() {
  q.value = ''
  kind.value = ''
  status.value = ''
  owner.value = ''
  release.value = ''
}
</script>
<template>
  <div class="page-heading">
    <div>
      <div class="eyebrow">
        <span class="eyebrow-line"></span>
        OPERAÇÃO / DEMANDAS
      </div>
      <h1>Demandas<span class="title-period">.</span></h1>
      <p>Todas as unidades de trabalho, com responsáveis e histórico de homologação.</p>
    </div>
    <div class="heading-actions">
      <AppButton
        label="Nova entrega"
        data-testid="new-demand"
        :icon="PlusIcon"
        @click="store.openCreate()"
      />
    </div>
  </div>
  <div class="surface-card list-card">
    <div class="filter-head">
      <div>
        <span class="overline">GESTÃO DE TICKETS</span>
        <h2>
          Lista de demandas
          <span class="light-count">{{ filtered.length }}</span>
        </h2>
      </div>
      <span class="text-xs text-slate-400">Clique em uma linha para abrir a demanda</span>
    </div>
    <div class="filter-toolbar">
      <IconField class="min-w-0 flex-1">
        <InputIcon><MagnifyingGlassIcon class="app-icon" /></InputIcon>
        <InputText
          v-model="q"
          placeholder="Pesquisar título, evento ou referência..."
          aria-label="Buscar demandas"
          data-testid="search-demand"
          fluid
        />
      </IconField>
      <Select
        v-model="kind"
        aria-label="Filtrar por esteira"
        :options="[{ label: 'Todas as esteiras', value: '' }, ...catalog.kindOptions]"
        option-label="label"
        option-value="value"
        placeholder="Todas as esteiras"
        class="filter-select"
      />
      <Select
        v-model="status"
        placeholder="Todos os status"
        aria-label="Filtrar por status"
        :options="[{ label: 'Todos os status', value: '' }, ...stages]"
        option-label="label"
        option-value="value"
        class="filter-select"
      />
      <Select
        v-model="owner"
        placeholder="Todos os responsáveis"
        aria-label="Filtrar por responsável"
        :options="[
          { label: 'Todos os responsáveis', value: '' },
          ...catalog.memberOptions
            .filter((o) => o.id)
            .map((o) => ({ label: o.label, value: o.id })),
        ]"
        option-label="label"
        option-value="value"
        class="filter-select"
      />
      <Select
        v-model="release"
        aria-label="Filtrar por versão"
        placeholder="Todas as versões"
        :options="[
          { label: 'Todas as versões', value: '' },
          { label: 'Sem versão', value: 'none' },
          ...store.releases.map((r) => ({
            label: r.name + (r.archived_at ? ' · Arquivada' : ''),
            value: r.id,
          })),
        ]"
        option-label="label"
        option-value="value"
        filter
      />
      <AppButton
        variant="ghost"
        label="Limpar"
        :icon="FunnelIcon"
        @click="clear"
      />
    </div>
    <DataTable
      :value="filtered"
      data-key="id"
      paginator
      :rows="10"
      :rows-per-page-options="[10, 20, 50]"
      removable-sort
      sort-field="updated_at"
      :sort-order="-1"
      class="work-table"
      :pt="{ bodyRow: { class: 'cursor-pointer' } }"
      @row-click="(e) => store.selectDemand(e.data.id)"
    >
      <Column
        field="number"
        header="DEMANDA"
        sortable
        style="min-width: 280px"
      >
        <template #sorticon="{ sortOrder, class: iconClass }">
          <component
            :is="
              sortOrder === 1 ? ArrowUpIcon : sortOrder === -1 ? ArrowDownIcon : ArrowsUpDownIcon
            "
            :class="iconClass"
            class="app-icon"
          />
        </template>
        <template #body="{ data }">
          <div class="ticket-row-name">
            <strong>{{ data.title }}</strong>
            <small>{{ demandCode(data.number) }} · {{ data.reference || 'Sem referência' }}</small>
          </div>
        </template>
      </Column>
      <Column
        field="kind"
        header="ESTEIRA"
        sortable
        style="min-width: 135px"
      >
        <template #sorticon="{ sortOrder, class: iconClass }">
          <component
            :is="
              sortOrder === 1 ? ArrowUpIcon : sortOrder === -1 ? ArrowDownIcon : ArrowsUpDownIcon
            "
            :class="iconClass"
            class="app-icon"
          />
        </template>
        <template #body="{ data }">
          <span
            class="kind-pill"
            :style="kindStyle(data.kind)"
          >
            {{ kindLabel(data.kind) }}
          </span>
        </template>
      </Column>
      <Column
        field="release_id"
        header="VERSÃO"
      >
        <template #body="{ data }">
          {{ store.releaseById(data.release_id)?.name || data.release_name || 'Sem versão' }}
        </template>
      </Column>
      <Column
        field="assignee"
        header="RESPONSÁVEL"
        sortable
        style="min-width: 165px"
      >
        <template #sorticon="{ sortOrder, class: iconClass }">
          <component
            :is="
              sortOrder === 1 ? ArrowUpIcon : sortOrder === -1 ? ArrowDownIcon : ArrowsUpDownIcon
            "
            :class="iconClass"
            class="app-icon"
          />
        </template>
        <template #body="{ data }">
          <div class="flex items-center gap-2">
            <span class="avatar-mini">{{ initials(data.assignee) }}</span>
            <span class="text-sm">{{ data.assignee || 'Não atribuído' }}</span>
          </div>
        </template>
      </Column>
      <Column
        field="status"
        header="STATUS"
        sortable
        style="min-width: 155px"
      >
        <template #sorticon="{ sortOrder, class: iconClass }">
          <component
            :is="
              sortOrder === 1 ? ArrowUpIcon : sortOrder === -1 ? ArrowDownIcon : ArrowsUpDownIcon
            "
            :class="iconClass"
            class="app-icon"
          />
        </template>
        <template #body="{ data }">
          <span
            class="status-pill"
            :class="data.status"
          >
            <span></span>
            {{ statusLabel(data.status) }}
          </span>
        </template>
      </Column>
      <Column
        field="priority"
        header="PRIORIDADE"
        sortable
      >
        <template #sorticon="{ sortOrder, class: iconClass }">
          <component
            :is="
              sortOrder === 1 ? ArrowUpIcon : sortOrder === -1 ? ArrowDownIcon : ArrowsUpDownIcon
            "
            :class="iconClass"
            class="app-icon"
          />
        </template>
        <template #body="{ data }">
          <span
            class="priority-badge"
            :class="data.priority"
          >
            <span></span>
            {{ priorityLabel(data.priority) }}
          </span>
        </template>
      </Column>
      <Column
        field="due_date"
        header="PRAZO"
        sortable
      >
        <template #sorticon="{ sortOrder, class: iconClass }">
          <component
            :is="
              sortOrder === 1 ? ArrowUpIcon : sortOrder === -1 ? ArrowDownIcon : ArrowsUpDownIcon
            "
            :class="iconClass"
            class="app-icon"
          />
        </template>
        <template #body="{ data }">
          <span
            class="text-xs"
            :class="
              isLate(data.due_date, data.status === 'done')
                ? 'font-bold text-rose-600'
                : 'text-slate-500'
            "
          >
            {{ dateLabel(data.due_date) }}
          </span>
        </template>
      </Column>
      <template #empty>
        <div class="empty-inline">Nenhuma demanda corresponde aos filtros.</div>
      </template>
    </DataTable>
  </div>
</template>
