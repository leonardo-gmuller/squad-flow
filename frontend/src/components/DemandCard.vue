<script setup lang="ts">
import AppButton from './AppButton.vue'
import SurfaceButton from './primevue/SurfaceButton'
import {
  ArrowsPointingOutIcon,
  ArrowUpRightIcon,
  ArrowUturnLeftIcon,
  CalendarDaysIcon,
  CheckBadgeIcon,
  ClipboardDocumentCheckIcon,
  UserIcon,
} from '@heroicons/vue/24/outline'
import { tooltip } from '../config/tooltip'
import { computed } from 'vue'
import Avatar from 'primevue/avatar'
import AvatarGroup from 'primevue/avatargroup'
import Tag from 'primevue/tag'
import type { Demand } from '../types'
import { useWorkspaceStore } from '../stores/workspace'
import { dateLabel, demandCode, initials, isLate } from '../utils/format'
import { chartDateLabel } from '../utils/chartDate'
import { kindLabel, kindStyle, priorityLabel } from '../config/catalog'
const props = defineProps<{
  item: Demand
  compact?: boolean
  movable?: boolean
  moving?: boolean
}>()
defineEmits<{ requestMove: [] }>()
const store = useWorkspaceStore()
const people = computed(() => {
  const item = props.item
  const assigned = [
    {
      id: item.assignee_id || `legacy:${item.assignee}`,
      name: item.assignee,
      role: 'Desenvolvedor',
    },
    { id: item.reviewer_id || `legacy:${item.reviewer}`, name: item.reviewer, role: 'QA' },
  ].filter((person) => person.name)
  return [
    ...new Map(
      assigned.map((person) => [
        person.id,
        {
          ...person,
          role: assigned
            .filter((other) => other.id === person.id)
            .map((other) => other.role)
            .join(' / '),
        },
      ]),
    ).values(),
  ]
})
const version = computed(
  () => store.releaseById(props.item.release_id)?.name || props.item.release_name,
)
function open() {
  store.selectDemand(props.item.id).catch(() => {})
}
</script>
<template>
  <div
    class="work-card-container"
    :class="{ 'has-move-handle': movable, 'is-moving': moving }"
    :data-demand-id="item.id"
    :aria-busy="moving || undefined"
  >
    <SurfaceButton
      severity="secondary"
      outlined
      class="work-card"
      :class="{ 'work-card-compact': compact }"
      :disabled="moving"
      :aria-label="`Abrir ${demandCode(item.number)} ${item.title}`"
      @click="open"
    >
      <div class="work-card-topline">
        <div class="work-card-tags">
          <Tag
            :value="kindLabel(item.kind)"
            rounded
            class="card-kind-tag"
            :style="kindStyle(item.kind)"
          />
          <Tag
            v-if="item.release_id"
            v-tooltip.top="tooltip(version)"
            :value="version"
            rounded
            class="card-version-tag"
          />
        </div>
        <ArrowUpRightIcon
          style="width: 16px; height: 16px; flex-shrink: 0"
          class="card-open-icon"
          aria-hidden="true"
        />
      </div>
      <div class="work-card-identity">
        <span class="work-card-code">{{ demandCode(item.number) }}</span>
        <span
          class="priority-badge"
          :class="item.priority"
        >
          <span></span>
          {{ priorityLabel(item.priority) }}
        </span>
      </div>
      <h3 class="work-card-title">{{ item.title }}</h3>
      <p
        v-if="!compact"
        class="work-card-description"
      >
        {{ item.description || 'Abra a demanda para consultar os detalhes da entrega.' }}
      </p>
      <div
        v-if="item.event_codes.length"
        class="work-card-events"
      >
        <span
          v-for="code in item.event_codes.slice(0, 2)"
          :key="code"
        >
          {{ code }}
        </span>
        <span v-if="item.event_codes.length > 2">+{{ item.event_codes.length - 2 }}</span>
      </div>
      <div
        v-if="item.status === 'done' || item.return_count || item.qa_cycle"
        class="work-card-workflow"
      >
        <span
          v-if="item.status === 'done'"
          class="card-delivered"
        >
          <CheckBadgeIcon style="width: 14px; height: 14px; flex-shrink: 0" />
          Entregue
        </span>
        <span
          v-else-if="item.qa_cycle"
          v-tooltip.top="tooltip(`Ciclo de QA ${item.qa_cycle}`)"
        >
          <ClipboardDocumentCheckIcon style="width: 14px; height: 14px; flex-shrink: 0" />
          Ciclo {{ item.qa_cycle }}
        </span>
        <span
          v-if="item.return_count"
          class="card-returned"
        >
          <ArrowUturnLeftIcon style="width: 13px; height: 13px; flex-shrink: 0" />
          {{ item.return_count }} {{ item.return_count === 1 ? 'devolução' : 'devoluções' }}
        </span>
      </div>
      <div class="work-card-foot">
        <AvatarGroup
          v-if="people.length"
          class="card-people"
        >
          <Avatar
            v-for="(person, index) in people"
            :key="person.id"
            v-tooltip.top="tooltip(`${person.name} · ${person.role}`)"
            :label="initials(person.name)"
            :aria-label="person.name"
            shape="circle"
            :class="`person-tone-${index}`"
          />
        </AvatarGroup>
        <Avatar
          v-else
          v-tooltip.top="tooltip('Sem responsável')"
          shape="circle"
          class="card-unassigned"
          aria-label="Sem responsável"
        >
          <template #icon><UserIcon class="app-icon" /></template>
        </Avatar>
        <span class="owner-name">{{ item.assignee || 'Sem responsável' }}</span>
        <time
          v-if="item.due_date"
          v-tooltip.top="tooltip(`Prazo: ${dateLabel(item.due_date)}`)"
          :datetime="item.due_date"
          class="due-marker"
          :class="{ 'is-late': isLate(item.due_date, item.status === 'done') }"
        >
          <CalendarDaysIcon
            aria-hidden="true"
            class="app-icon"
          />
          {{ chartDateLabel(item.due_date) }}
        </time>
      </div>
    </SurfaceButton>
    <AppButton
      v-if="movable"
      v-tooltip.top="tooltip('Arraste para outra etapa ou clique para mover')"
      variant="ghost"
      class="kanban-move-handle"
      size="small"
      :loading="moving"
      :disabled="moving"
      :aria-label="`Mover ${demandCode(item.number)} ${item.title}`"
      aria-describedby="board-drag-help"
      @click.stop="$emit('requestMove')"
    >
      <template #icon="slotProps">
        <ArrowsPointingOutIcon
          :class="slotProps.class"
          class="app-icon"
        />
      </template>
    </AppButton>
  </div>
</template>
