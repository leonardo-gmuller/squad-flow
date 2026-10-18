<script setup lang="ts">
import { useCatalogStore } from '../stores/catalog'
import AppButton from '../components/AppButton.vue'
import Select from '../components/primevue/Select'
import Message from '../components/primevue/Message'
import Dialog from '../components/primevue/Dialog'
import {
  AdjustmentsHorizontalIcon,
  ArchiveBoxIcon,
  ArrowPathIcon,
  ArrowUpRightIcon,
  ChartBarSquareIcon,
  CheckBadgeIcon,
  CheckCircleIcon,
  FunnelIcon,
  InformationCircleIcon,
  ListBulletIcon,
  MagnifyingGlassIcon,
  PlusIcon,
  TagIcon,
  ViewColumnsIcon,
} from '@heroicons/vue/24/outline'
import { tooltip } from '../config/tooltip'
import { computed, onUnmounted, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import InputText from 'primevue/inputtext'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import Textarea from 'primevue/textarea'
import { useToast } from 'primevue/usetoast'
import Avatar from 'primevue/avatar'
import AvatarGroup from 'primevue/avatargroup'
import Tag from 'primevue/tag'
import { demandCode, initials } from '../utils/format'
import { useWorkspaceStore } from '../stores/workspace'
import { useAuthStore } from '../stores/auth'
import { useBoardDrag } from '../composables/useBoardDrag'
import { stages, statusLabel, transitions } from '../config/catalog'
import DemandCard from '../components/DemandCard.vue'
import FormField from '../components/FormField.vue'
import { api, APIError } from '../api'
import type { BoardColumn, BoardFilter, Demand, Status } from '../types'

const catalog = useCatalogStore()
const store = useWorkspaceStore(),
  route = useRoute(),
  router = useRouter()
const auth = useAuthStore(),
  toast = useToast()
const q = ref(String(route.query.q || '')),
  kind = ref(String(route.query.kind || '')),
  owner = ref(String(route.query.owner || ''))
const release = ref(String(route.query.release_id || '')),
  archive = ref(String(route.query.archive || 'active'))
const columns = ref<BoardColumn[]>([]),
  ownerIDs = ref<string[]>([]),
  loading = ref(false),
  error = ref('')
const loadingColumns = ref<Status[]>([])
const movingId = ref(''),
  moveVisible = ref(false),
  moveError = ref(''),
  announcement = ref('')
const moveItem = ref<Demand | null>(null),
  moveTarget = ref<Status>('in_progress'),
  moveReason = ref('')
const moveOptions = computed(() =>
  stages.filter(
    (stage) => moveItem.value && transitions[moveItem.value.status].includes(stage.value),
  ),
)
let reloadPending = false
const drag = useBoardDrag({
  getItem: (id) => columns.value.flatMap((column) => column.items).find((item) => item.id === id),
  canStart: (item) =>
    !loading.value &&
    !store.busy &&
    !movingId.value &&
    !moveVisible.value &&
    transitions[item.status].length > 0,
  onDrop: (item, status) => {
    if (!canDrop(item, status)) {
      announcement.value = 'Essa etapa não está disponível para esta entrega.'
      return
    }
    if (status === 'changes_requested') openMove(item, status)
    else void submitMove(item, status)
  },
})
const dragHelp = computed(() =>
  drag.item.value
    ? `Solte em ${transitions[drag.item.value.status].map(statusLabel).join(' ou ')}. Pressione Escape para cancelar.`
    : movingId.value
      ? 'Salvando movimentação…'
      : announcement.value,
)
function canDrop(item: Demand, status: Status) {
  return transitions[item.status].includes(status)
}
function openMove(item: Demand, target?: Status) {
  if (movingId.value || store.busy || !transitions[item.status].length) return
  moveItem.value = item
  moveTarget.value = target || transitions[item.status][0]
  moveReason.value = ''
  moveError.value = ''
  moveVisible.value = true
}
async function submitMove(item: Demand, target: Status, reason = '') {
  if (movingId.value || store.busy || !canDrop(item, target)) return
  const current = columns.value
    .flatMap((column) => column.items)
    .find((value) => value.id === item.id)
  if (!current || current.status !== item.status) {
    moveError.value = 'Esta entrega foi atualizada. Abra a movimentação novamente.'
    return
  }
  if (target === 'changes_requested' && !reason.trim()) {
    moveError.value = 'Informe o motivo da devolução.'
    return
  }
  moveError.value = ''
  movingId.value = item.id
  try {
    await api.move(item.id, target, auth.user?.name || '', reason.trim())
    if (disposed) return
    moveVisible.value = false
    announcement.value = `${demandCode(item.number)} movida para ${statusLabel(target)}.`
    toast.add({
      severity: 'success',
      summary: 'Entrega movimentada',
      detail: announcement.value,
      life: 3000,
    })
    reloadPending = true
    // A refresh failure must not turn a committed movement into an apparent
    // failure. The board reloads separately and offers its own retry state.
    try {
      await store.refresh()
    } catch {
      if (!disposed)
        toast.add({
          severity: 'warn',
          summary: 'Movimentação salva',
          detail: 'Não foi possível atualizar os indicadores agora.',
          life: 5000,
        })
    }
  } catch (err) {
    if (disposed) return
    moveError.value =
      err instanceof APIError && err.status === 409 && err.message.includes('QA requires')
        ? 'Para concluir, registre ao menos um teste aprovado e nenhum reprovado no ciclo atual de QA.'
        : (err as Error).message
    announcement.value = `Não foi possível mover ${demandCode(item.number)}. ${moveError.value}`
  } finally {
    if (!disposed) movingId.value = ''
  }
}
function scheduleReload() {
  if (disposed) return
  if (drag.item.value || movingId.value || moveVisible.value) {
    reloadPending = true
    return
  }
  if (timer) return
  reloadPending = false
  timer = setTimeout(() => {
    timer = undefined
    if (drag.item.value || movingId.value || moveVisible.value) {
      reloadPending = true
      return
    }
    void reload()
  }, 100)
}
const owners = computed(() => ownerIDs.value.map((id) => ({ id, name: catalog.memberName(id) })))
const showFilters = ref(
  Boolean(
    route.query.kind ||
    route.query.owner ||
    (route.query.archive && route.query.archive !== 'active'),
  ),
)
const activeFilters = computed(
  () =>
    Number(Boolean(kind.value)) + Number(Boolean(owner.value)) + Number(archive.value !== 'active'),
)
const activeVersion = computed(() => store.releaseById(release.value))
const viewQuery = computed(() => ({
  release_id: release.value || undefined,
  q: q.value || undefined,
  kind: kind.value || undefined,
}))
let epoch = 0,
  timer: ReturnType<typeof setTimeout> | undefined,
  controller: AbortController | undefined
let disposed = false
const total = computed(() => columns.value.reduce((sum, column) => sum + column.total, 0))
const releaseOptions = computed(() => [
  { label: 'Todas as versões', value: '' },
  { label: 'Sem versão', value: 'none' },
  ...store.releases
    .filter(
      (r) => archive.value === 'all' || Boolean(r.archived_at) === (archive.value === 'archived'),
    )
    .map((r) => ({ label: r.name + (r.archived_at ? ' · Arquivada' : ''), value: r.id })),
])
const archiveOptions = [
  { label: 'Versões ativas', value: 'active' },
  { label: 'Versões arquivadas', value: 'archived' },
  { label: 'Todas, incluindo arquivadas', value: 'all' },
]
const filter = (): BoardFilter => ({
  q: q.value,
  kind: kind.value,
  owner: owner.value,
  release_id: release.value,
  archive: archive.value,
  limit: 20,
})
const columnFor = (status: Status) => columns.value.find((column) => column.status === status)
async function reload() {
  clearTimeout(timer)
  timer = undefined
  controller?.abort()
  controller = new AbortController()
  const current = ++epoch
  loading.value = true
  loadingColumns.value = []
  error.value = ''
  columns.value = []
  try {
    const board = await api.board(filter(), controller.signal)
    if (current !== epoch || disposed) return
    columns.value = board.columns
    ownerIDs.value = board.owners
  } catch (err) {
    if (current === epoch && !disposed && (err as Error).name !== 'AbortError')
      error.value = (err as Error).message
  } finally {
    if (current === epoch && !disposed) loading.value = false
  }
}
async function loadMore(status: Status) {
  const column = columnFor(status)
  if (!column || loadingColumns.value.includes(status)) return
  const current = epoch
  loadingColumns.value.push(status)
  try {
    const board = await api.board(
      { ...filter(), status, offset: column.items.length },
      controller?.signal,
    )
    if (current !== epoch || disposed) return
    const next = board.columns.find((c) => c.status === status)
    if (next) {
      column.items = [
        ...new Map([...column.items, ...next.items].map((item) => [item.id, item])).values(),
      ]
      column.total = next.total
    }
  } catch (err) {
    if (current === epoch && !disposed && (err as Error).name !== 'AbortError')
      error.value = (err as Error).message
  } finally {
    if (current === epoch && !disposed)
      loadingColumns.value = loadingColumns.value.filter((value) => value !== status)
  }
}
watch([q, kind, owner, release, archive], () => {
  drag.cancel()
  if (!movingId.value) moveVisible.value = false
  // Invalidate in-flight pages as soon as filters change, before the debounce expires.
  epoch++
  controller?.abort()
  loadingColumns.value = []
  loading.value = true
  columns.value = []
  clearTimeout(timer)
  timer = setTimeout(() => {
    timer = undefined
    void router.replace({
      path: '/board',
      query: {
        q: q.value || undefined,
        kind: kind.value || undefined,
        owner: owner.value || undefined,
        release_id: release.value || undefined,
        archive: archive.value,
      },
    })
    void reload()
  }, 250)
})
watch(
  () => route.query,
  (value) => {
    q.value = String(value.q || '')
    kind.value = String(value.kind || '')
    owner.value = String(value.owner || '')
    release.value = String(value.release_id || '')
    archive.value = String(value.archive || 'active')
  },
)
watch(
  () => store.dataRevision,
  () => {
    scheduleReload()
  },
)
watch([drag.item, movingId, moveVisible], () => {
  if (reloadPending) scheduleReload()
})
watch(archive, () => {
  if (
    release.value &&
    release.value !== 'none' &&
    !releaseOptions.value.some((option) => option.value === release.value)
  )
    release.value = ''
})
function clear() {
  q.value = ''
  kind.value = ''
  owner.value = ''
  release.value = ''
  archive.value = 'active'
}
onUnmounted(() => {
  disposed = true
  epoch++
  clearTimeout(timer)
  controller?.abort()
})
if (store.loaded) void reload()
else loading.value = true
</script>

<template>
  <div class="board-workspace">
    <div class="page-heading board-heading">
      <div>
        <div class="eyebrow">WORKSPACE DE DESENVOLVIMENTO</div>
        <h1>Quadro Kanban<span class="title-period">.</span></h1>
        <p>
          {{
            activeVersion
              ? `Entregas da versão ${activeVersion.name}`
              : 'Cada entrega no seu lugar. Do planejamento à homologação.'
          }}
        </p>
      </div>
      <div class="heading-actions">
        <Tag
          v-if="archive !== 'active'"
          :value="archive === 'archived' ? 'Arquivo de versões' : 'Incluindo arquivadas'"
          severity="secondary"
          rounded
        />
        <AppButton
          label="Nova entrega"
          data-testid="new-demand"
          :icon="PlusIcon"
          @click="
            store.openCreate('', store.activeReleases.some((r) => r.id === release) ? release : '')
          "
        />
      </div>
    </div>
    <div class="board-navigation">
      <nav
        class="board-view-tabs"
        aria-label="Visualizações do workspace"
      >
        <RouterLink
          :to="{ path: '/board', query: route.query }"
          class="board-view-tab is-current"
          aria-current="page"
        >
          <ViewColumnsIcon style="width: 17px; height: 17px; flex-shrink: 0" />
          Kanban
        </RouterLink>
        <RouterLink
          :to="{ path: '/tickets', query: viewQuery }"
          class="board-view-tab"
        >
          <ListBulletIcon style="width: 18px; height: 18px; flex-shrink: 0" />
          Demandas
        </RouterLink>
        <RouterLink
          :to="{ path: '/releases', query: { release_id: activeVersion?.id } }"
          class="board-view-tab"
        >
          <TagIcon style="width: 17px; height: 17px; flex-shrink: 0" />
          Versões
        </RouterLink>
        <RouterLink
          to="/reports"
          class="board-view-tab"
        >
          <ChartBarSquareIcon style="width: 17px; height: 17px; flex-shrink: 0" />
          Relatórios
        </RouterLink>
      </nav>
      <div
        v-if="owners.length"
        class="board-team"
        aria-label="Responsáveis do workspace"
      >
        <AvatarGroup>
          <Avatar
            v-for="(person, index) in owners.slice(0, 4)"
            :key="person.id"
            v-tooltip.bottom="tooltip(person.name)"
            :label="initials(person.name)"
            :aria-label="person.name"
            shape="circle"
            :class="`person-tone-${index % 4}`"
          />
          <Avatar
            v-if="owners.length > 4"
            v-tooltip.bottom="
              tooltip(
                owners
                  .slice(4)
                  .map((person) => person.name)
                  .join(', '),
              )
            "
            :label="`+${owners.length - 4}`"
            :aria-label="`${owners.length - 4} outros responsáveis`"
            shape="circle"
            class="person-overflow"
          />
        </AvatarGroup>
        <span>{{ owners.length }} {{ owners.length === 1 ? 'responsável' : 'responsáveis' }}</span>
      </div>
    </div>
    <div class="board-filterbar">
      <IconField class="board-search">
        <InputIcon><MagnifyingGlassIcon class="app-icon" /></InputIcon>
        <InputText
          v-model="q"
          maxlength="300"
          placeholder="Buscar no quadro..."
          aria-label="Buscar demandas no quadro"
          fluid
        />
      </IconField>
      <Select
        v-model="release"
        class="board-version-select"
        aria-label="Filtrar por versão"
        placeholder="Todas as versões"
        :options="releaseOptions"
        option-label="label"
        option-value="value"
        filter
      />
      <AppButton
        variant="ghost"
        label="Filtros"
        :badge="activeFilters ? String(activeFilters) : undefined"
        :aria-expanded="showFilters"
        aria-controls="board-advanced-filters"
        :icon="AdjustmentsHorizontalIcon"
        @click="showFilters = !showFilters"
      />
      <span
        class="board-filter-count"
        role="status"
      >
        {{ loading ? 'Carregando…' : `${total} entregas` }}
      </span>
    </div>
    <Transition name="board-filters">
      <div
        v-show="showFilters"
        class="board-filter-collapse"
        :inert="!showFilters"
        :aria-hidden="!showFilters"
      >
        <div class="board-filter-collapse-inner">
          <div
            id="board-advanced-filters"
            class="board-advanced-filters"
          >
            <Select
              v-model="archive"
              aria-label="Visibilidade das versões"
              :options="archiveOptions"
              option-label="label"
              option-value="value"
            />
            <Select
              v-model="kind"
              aria-label="Filtrar por esteira"
              placeholder="Todas as esteiras"
              :options="[{ label: 'Todas as esteiras', value: '' }, ...catalog.kindOptions]"
              option-label="label"
              option-value="value"
            />
            <Select
              v-model="owner"
              aria-label="Filtrar por responsável"
              placeholder="Todos os responsáveis"
              :options="[
                { label: 'Todos os responsáveis', value: '' },
                ...catalog.memberOptions
                  .filter((o) => o.id)
                  .map((o) => ({ label: o.label, value: o.id })),
              ]"
              option-label="label"
              option-value="value"
              filter
            />
            <AppButton
              variant="ghost"
              label="Limpar filtros"
              :icon="FunnelIcon"
              @click="clear"
            />
            <RouterLink
              to="/releases"
              class="board-manage-link"
            >
              Gerenciar versões
              <ArrowUpRightIcon class="app-icon" />
            </RouterLink>
          </div>
        </div>
      </div>
    </Transition>
    <Message
      v-if="error"
      severity="error"
      :closable="false"
      class="mb-4"
    >
      <div class="flex flex-wrap items-center gap-3">
        <span>{{ error }}</span>
        <AppButton
          variant="secondary"
          label="Tentar novamente"
          size="small"
          @click="reload"
        />
      </div>
    </Message>
    <Message
      v-if="moveError && !moveVisible"
      severity="error"
      :closable="false"
      class="mb-4"
      role="alert"
    >
      {{ moveError }}
    </Message>
    <span
      class="sr-only"
      role="status"
      aria-live="polite"
    >
      {{ dragHelp }}
    </span>
    <div
      class="kanban-board"
      :aria-busy="loading || Boolean(movingId)"
      @pointerdown="drag.start"
      @click.capture="drag.suppressClick"
      @dragstart.prevent
    >
      <section
        v-for="stage in stages"
        :key="stage.value"
        class="kanban-column"
        :aria-label="stage.label"
        :data-testid="`board-${stage.value}`"
        :data-board-status="stage.value"
        :class="{
          'is-drop-allowed': drag.item.value && canDrop(drag.item.value, stage.value),
          'is-drop-blocked':
            drag.item.value &&
            stage.value !== drag.item.value.status &&
            !canDrop(drag.item.value, stage.value),
          'is-drop-over':
            drag.over.value === stage.value &&
            drag.item.value &&
            canDrop(drag.item.value, stage.value),
        }"
      >
        <header class="kanban-column-head">
          <div class="kanban-stage-title">
            <span
              class="stage-bubble"
              :class="stage.tone"
            ></span>
            <h2>{{ stage.label }}</h2>
            <span class="column-count">{{ columnFor(stage.value)?.total || 0 }}</span>
          </div>
          <AppButton
            v-if="stage.value === 'backlog'"
            v-tooltip.top="tooltip('Adicionar entrega ao backlog')"
            variant="ghost"
            size="small"
            aria-label="Adicionar entrega ao backlog"
            :icon="PlusIcon"
            @click="
              store.openCreate(
                '',
                store.activeReleases.some((r) => r.id === release) ? release : '',
              )
            "
          />
          <component
            :is="
              stage.value === 'done'
                ? CheckCircleIcon
                : stage.value === 'in_qa'
                  ? CheckBadgeIcon
                  : ArrowPathIcon
            "
            v-else
            class="app-icon column-stage-icon"
            aria-hidden="true"
          />
        </header>
        <p class="kanban-subtitle">{{ stage.subtitle }}</p>
        <div class="kanban-items">
          <div
            v-if="loading"
            class="kanban-empty"
            role="status"
          >
            Carregando entregas…
          </div>
          <template v-else>
            <DemandCard
              v-for="item in columnFor(stage.value)?.items || []"
              :key="item.id"
              :item="item"
              :movable="transitions[item.status].length > 0"
              :moving="movingId === item.id"
              :class="{ 'is-drag-source': drag.item.value?.id === item.id }"
              @request-move="openMove(item)"
            />
            <div
              v-if="!columnFor(stage.value)?.total && !error"
              class="kanban-empty"
            >
              <ArchiveBoxIcon
                aria-hidden="true"
                class="app-icon"
              />
              <span>Nenhuma entrega nesta etapa.</span>
            </div>
            <AppButton
              v-if="
                (columnFor(stage.value)?.items.length || 0) < (columnFor(stage.value)?.total || 0)
              "
              variant="secondary"
              :label="`Carregar mais · ${(columnFor(stage.value)?.total || 0) - (columnFor(stage.value)?.items.length || 0)} restantes`"
              :aria-label="`Carregar mais em ${stage.label}`"
              fluid
              :loading="loadingColumns.includes(stage.value)"
              :disabled="loadingColumns.includes(stage.value)"
              @click="loadMore(stage.value)"
            />
          </template>
        </div>
      </section>
    </div>
    <p
      id="board-drag-help"
      class="board-hint"
    >
      <InformationCircleIcon
        aria-hidden="true"
        class="app-icon"
      />
      <span>
        {{
          archive === 'active'
            ? 'O quadro mostra versões ativas e demandas sem versão.'
            : 'Você está consultando o histórico das versões.'
        }}
        Arraste os cards para as etapas disponíveis. No celular, use o ícone de mover; clique nele
        ou pressione Enter para escolher a etapa.
      </span>
    </p>
    <Dialog
      v-model:visible="moveVisible"
      modal
      header="Mover entrega"
      :style="{ width: 'min(480px,calc(100vw - 24px))' }"
      :closable="!movingId"
      :close-on-escape="!movingId"
    >
      <form
        v-if="moveItem"
        id="board-move-form"
        class="grid gap-5"
        @submit.prevent="submitMove(moveItem, moveTarget, moveReason)"
      >
        <p class="text-sm">
          <strong>{{ demandCode(moveItem.number) }}</strong>
          · {{ moveItem.title }}
        </p>
        <FormField
          id="board-move-target"
          label="Mover para"
          required
        >
          <Select
            v-model="moveTarget"
            input-id="board-move-target"
            aria-labelledby="board-move-target-label"
            :options="moveOptions"
            option-label="label"
            option-value="value"
            :disabled="Boolean(movingId)"
            fluid
          />
        </FormField>
        <FormField
          v-if="moveTarget === 'changes_requested'"
          id="board-move-reason"
          label="Motivo da devolução"
          required
        >
          <Textarea
            id="board-move-reason"
            v-model="moveReason"
            required
            maxlength="2000"
            rows="3"
            :disabled="Boolean(movingId)"
            fluid
            placeholder="Descreva o que precisa ser corrigido"
          />
        </FormField>
        <Message
          v-if="moveTarget === 'done'"
          severity="info"
          :closable="false"
        >
          A conclusão exige ao menos um teste aprovado e nenhum reprovado no ciclo atual de QA.
        </Message>
        <Message
          v-if="moveError"
          severity="error"
          :closable="false"
        >
          {{ moveError }}
        </Message>
        <p class="text-xs text-[var(--p-text-muted-color)]">
          Movimentação registrada por {{ auth.user?.name }}.
        </p>
      </form>
      <template #footer>
        <AppButton
          variant="ghost"
          label="Cancelar"
          :disabled="Boolean(movingId)"
          @click="moveVisible = false"
        />
        <AppButton
          label="Mover entrega"
          type="submit"
          form="board-move-form"
          :loading="Boolean(movingId)"
        />
      </template>
    </Dialog>
    <Teleport to="body">
      <div
        v-if="drag.item.value"
        class="kanban-drag-preview"
        :style="{
          left: `${drag.preview.value.x}px`,
          top: `${drag.preview.value.y}px`,
          width: `${drag.preview.value.width}px`,
        }"
        aria-hidden="true"
      >
        <DemandCard
          :item="drag.item.value"
          compact
        />
      </div>
    </Teleport>
  </div>
</template>
