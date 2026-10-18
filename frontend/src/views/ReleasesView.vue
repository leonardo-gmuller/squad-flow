<script setup lang="ts">
import AppButton from '../components/AppButton.vue'
import Select from '../components/primevue/Select'
import DataTable from '../components/primevue/DataTable'
import Dialog from '../components/primevue/Dialog'
import Message from '../components/primevue/Message'
import {
  ArchiveBoxIcon,
  ArrowDownIcon,
  ArrowPathIcon,
  ArrowUpIcon,
  ArrowUpRightIcon,
  ArrowsUpDownIcon,
  PencilSquareIcon,
  PlusIcon,
} from '@heroicons/vue/24/outline'
import { tooltip } from '../config/tooltip'
import { computed, reactive, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import Column from 'primevue/column'
import Tag from 'primevue/tag'
import ProgressBar from 'primevue/progressbar'
import { useToast } from 'primevue/usetoast'
import FormField from '../components/FormField.vue'
import CalendarDateInput from '../components/CalendarDateInput.vue'
import { useWorkspaceStore } from '../stores/workspace'
import { dateLabel, demandCode, percent } from '../utils/format'
import { statusLabel } from '../config/catalog'
import type { Release } from '../types'

const store = useWorkspaceStore(),
  route = useRoute(),
  router = useRouter(),
  toast = useToast()
const query = ref(''),
  archive = ref('active'),
  error = ref('')
const editing = ref(false),
  editId = ref(''),
  pendingArchive = ref<Release | null>(null)
const form = reactive({ name: '', description: '', date: '' })
const filtered = computed(() =>
  store.releases.filter(
    (r) =>
      (archive.value === 'all' || Boolean(r.archived_at) === (archive.value === 'archived')) &&
      `${r.name} ${r.description}`.toLowerCase().includes(query.value.toLowerCase()),
  ),
)
const selected = computed(() => store.releaseById(String(route.query.release_id || '')))
const items = computed(() =>
  selected.value ? store.items.filter((item) => item.release_id === selected.value!.id) : [],
)
const archiveOptions = [
  { label: 'Versões ativas', value: 'active' },
  { label: 'Arquivadas', value: 'archived' },
  { label: 'Todas as versões', value: 'all' },
]
function confirmArchive(release: Release) {
  error.value = ''
  pendingArchive.value = release
}
function openForm(release?: Release) {
  editId.value = release?.id || ''
  Object.assign(form, {
    name: release?.name || '',
    description: release?.description || '',
    date: release?.release_date?.slice(0, 10) || '',
  })
  error.value = ''
  editing.value = true
}
async function save() {
  error.value = ''
  try {
    const input = {
      name: form.name.trim(),
      description: form.description,
      release_date: form.date ? `${form.date}T12:00:00Z` : null,
    }
    if (editId.value) await store.editRelease(editId.value, input)
    else {
      const release = await store.createRelease(input)
      archive.value = 'active'
      await router.replace({ path: '/releases', query: { release_id: release.id } })
    }
    editing.value = false
    toast.add({ severity: 'success', summary: 'Versão salva', life: 3000 })
  } catch (err) {
    error.value = (err as Error).message
  }
}
async function setArchived(release: Release, value: boolean) {
  error.value = ''
  try {
    await store.archiveRelease(release.id, value)
    pendingArchive.value = null
    toast.add({
      severity: 'success',
      summary: value ? 'Versão arquivada' : 'Versão restaurada',
      life: 3500,
    })
  } catch (err) {
    error.value = (err as Error).message
  }
}
function boardLink(release: Release) {
  return {
    path: '/board',
    query: { release_id: release.id, archive: release.archived_at ? 'archived' : 'active' },
  }
}
</script>

<template>
  <div class="page-heading">
    <div>
      <div class="eyebrow">
        <span class="eyebrow-line"></span>
        PLANEJAMENTO / ENTREGAS
      </div>
      <h1>Versões<span class="title-period">.</span></h1>
      <p>Saiba o que foi entregue em cada versão e mantenha o Kanban focado no trabalho atual.</p>
    </div>
    <div class="heading-actions">
      <AppButton
        label="Nova versão"
        :icon="PlusIcon"
        @click="openForm()"
      />
    </div>
  </div>
  <div class="report-metrics release-summary">
    <div class="report-metric">
      <span>Versões ativas</span>
      <strong>{{ store.activeReleases.length }}</strong>
      <p>Disponíveis para novas demandas</p>
    </div>
    <div class="report-metric">
      <span>Versões arquivadas</span>
      <strong>{{ store.releases.length - store.activeReleases.length }}</strong>
      <p>Histórico preservado e consultável</p>
    </div>
  </div>
  <Message
    v-if="error && !editing && !pendingArchive"
    severity="error"
    :closable="false"
    class="mb-5"
  >
    {{ error }}
  </Message>
  <section class="surface-card list-card">
    <div class="filter-toolbar">
      <InputText
        v-model="query"
        aria-label="Buscar versões"
        placeholder="Buscar versão ou descrição..."
        class="min-w-0 flex-1"
      />
      <Select
        v-model="archive"
        aria-label="Situação das versões"
        :options="archiveOptions"
        option-label="label"
        option-value="value"
      />
    </div>
    <DataTable
      :value="filtered"
      data-key="id"
      paginator
      :rows="10"
      :rows-per-page-options="[10, 20, 50]"
      scrollable
      :table-props="{ 'aria-label': 'Versões de entrega' }"
    >
      <Column
        field="name"
        header="Versão"
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
          <AppButton
            variant="link"
            :label="data.name"
            @click="router.replace({ path: '/releases', query: { release_id: data.id } })"
          />
          <Tag
            v-if="data.archived_at"
            value="Arquivada"
            severity="secondary"
          />
        </template>
      </Column>
      <Column header="Data de entrega">
        <template #body="{ data }">{{ dateLabel(data.release_date?.slice(0, 10)) }}</template>
      </Column>
      <Column header="Entregas">
        <template #body="{ data }">{{ data.done }} / {{ data.total }} concluídas</template>
      </Column>
      <Column
        header="Progresso"
        class="min-w-36"
      >
        <template #body="{ data }">
          <ProgressBar
            :value="percent(data.done, data.total)"
            :show-value="false"
            class="package-progress"
            :aria-label="`Progresso da versão ${data.name}`"
          />
        </template>
      </Column>
      <Column header="Ações">
        <template #body="{ data }">
          <div class="flex flex-wrap gap-1">
            <AppButton
              v-tooltip.top="tooltip(`Editar versão ${data.name}`)"
              variant="ghost"
              size="small"
              :aria-label="`Editar versão ${data.name}`"
              :icon="PencilSquareIcon"
              @click="openForm(data)"
            />
            <AppButton
              v-if="!data.archived_at"
              v-tooltip.top="tooltip(`Arquivar versão ${data.name}`)"
              variant="ghost"
              size="small"
              :aria-label="`Arquivar versão ${data.name}`"
              :icon="ArchiveBoxIcon"
              @click="confirmArchive(data)"
            />
            <AppButton
              v-else
              variant="ghost"
              label="Restaurar"
              :aria-label="`Restaurar versão ${data.name}`"
              :loading="store.busy"
              :icon="ArrowPathIcon"
              @click="setArchived(data, false)"
            />
          </div>
        </template>
      </Column>
      <template #empty>
        <div class="empty-inline">
          Nenhuma versão encontrada. Crie uma versão para organizar as próximas entregas.
        </div>
      </template>
    </DataTable>
  </section>
  <section
    v-if="selected"
    class="surface-card releases-detail mt-6"
  >
    <div class="card-heading flex-wrap">
      <div>
        <span class="overline">CONTEÚDO DA VERSÃO</span>
        <h2>{{ selected.name }}</h2>
        <p>{{ selected.description || 'Entregas vinculadas a esta versão.' }}</p>
      </div>
      <Tag
        :value="selected.archived_at ? 'Arquivada' : 'Ativa'"
        :severity="selected.archived_at ? 'secondary' : 'success'"
      />
    </div>
    <div class="my-5 flex flex-wrap items-center gap-3">
      <span class="text-sm">
        {{ selected.done }} {{ selected.done === 1 ? 'concluída' : 'concluídas' }} ·
        {{ selected.total - selected.done }}
        {{ selected.total - selected.done === 1 ? 'aberta' : 'abertas' }} · {{ selected.in_qa }} em
        QA
      </span>
      <RouterLink
        :to="boardLink(selected)"
        class="subtle-link"
      >
        Abrir no Kanban
        <ArrowUpRightIcon class="app-icon" />
      </RouterLink>
      <AppButton
        v-if="!selected.archived_at"
        label="Adicionar demanda"
        size="small"
        :icon="PlusIcon"
        @click="store.openCreate('', selected.id)"
      />
    </div>
    <Message
      v-if="selected.archived_at"
      severity="info"
      :closable="false"
      class="mb-4"
    >
      Esta versão está arquivada. Suas demandas continuam disponíveis para consulta; restaure a
      versão para vincular novas entregas.
    </Message>
    <DataTable
      :value="items"
      data-key="id"
      paginator
      :rows="10"
      :rows-per-page-options="[10, 20, 50]"
      scrollable
      :table-props="{ 'aria-label': `Demandas da versão ${selected.name}` }"
      :pt="{ bodyRow: { class: 'cursor-pointer' } }"
      @row-click="(event) => store.selectDemand(event.data.id)"
    >
      <Column header="Demanda">
        <template #body="{ data }">
          <AppButton
            variant="link"
            :label="`${demandCode(data.number)} · ${data.title}`"
            @click.stop="store.selectDemand(data.id)"
          />
        </template>
      </Column>
      <Column
        field="assignee"
        header="Responsável"
      />
      <Column header="Status">
        <template #body="{ data }">
          <Tag
            :value="statusLabel(data.status)"
            :severity="data.status === 'done' ? 'success' : 'secondary'"
          />
        </template>
      </Column>
      <Column header="Concluída em">
        <template #body="{ data }">{{ dateLabel(data.completed_at) }}</template>
      </Column>
      <template #empty>
        <div class="empty-inline">Nenhuma demanda vinculada a esta versão.</div>
      </template>
    </DataTable>
  </section>
  <Dialog
    v-model:visible="editing"
    modal
    :header="editId ? 'Editar versão' : 'Nova versão'"
    class="w-[min(560px,96vw)]"
  >
    <form
      id="release-form"
      class="grid gap-5"
      @submit.prevent="save"
    >
      <FormField
        id="release-name"
        label="Nome da versão"
        required
      >
        <InputText
          id="release-name"
          v-model="form.name"
          required
          maxlength="80"
          placeholder="Ex.: 2.6.0 ou 2026.10"
          fluid
        />
      </FormField>
      <FormField
        id="release-description"
        label="Descrição"
      >
        <Textarea
          id="release-description"
          v-model="form.description"
          rows="3"
          maxlength="20000"
          fluid
        />
      </FormField>
      <FormField
        id="release-date"
        label="Data de entrega (opcional)"
      >
        <CalendarDateInput
          v-model="form.date"
          input-id="release-date"
        />
      </FormField>
      <Message
        v-if="error"
        severity="error"
        :closable="false"
      >
        {{ error }}
      </Message>
    </form>
    <template #footer>
      <AppButton
        variant="ghost"
        label="Cancelar"
        @click="editing = false"
      />
      <AppButton
        label="Salvar versão"
        type="submit"
        form="release-form"
        :loading="store.busy"
      />
    </template>
  </Dialog>
  <Dialog
    :visible="Boolean(pendingArchive)"
    modal
    header="Arquivar versão"
    class="w-[min(560px,96vw)]"
    @update:visible="
      (value) => {
        if (!value) pendingArchive = null
      }
    "
  >
    <template v-if="pendingArchive">
      <p class="text-sm leading-relaxed">
        Arquivar
        <strong>{{ pendingArchive.name }}</strong>
        retira suas {{ pendingArchive.total }} demandas da visão ativa do Kanban.
      </p>
      <Message
        v-if="pendingArchive.total > pendingArchive.done"
        severity="warn"
        :closable="false"
        class="my-4"
      >
        Esta versão ainda tem {{ pendingArchive.total - pendingArchive.done }}
        {{
          pendingArchive.total - pendingArchive.done === 1 ? 'demanda aberta' : 'demandas abertas'
        }}. Elas manterão o status atual e poderão ser consultadas no filtro de versões arquivadas.
      </Message>
      <p class="mt-4 text-sm leading-relaxed">
        As {{ pendingArchive.done }} entregas concluídas, testes, métricas e históricos serão
        preservados. Você pode restaurar a versão a qualquer momento.
      </p>
      <Message
        v-if="error"
        severity="error"
        :closable="false"
        class="mt-4"
      >
        {{ error }}
      </Message>
    </template>
    <template #footer>
      <AppButton
        variant="ghost"
        label="Cancelar"
        @click="pendingArchive = null"
      />
      <AppButton
        label="Arquivar versão"
        :loading="store.busy"
        :icon="ArchiveBoxIcon"
        @click="pendingArchive && setArchived(pendingArchive, true)"
      />
    </template>
  </Dialog>
</template>

<style scoped>
.release-summary {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}
.releases-detail {
  padding: 24px;
  min-width: 0;
}
.releases-detail h2,
.releases-detail p {
  overflow-wrap: anywhere;
}
</style>
