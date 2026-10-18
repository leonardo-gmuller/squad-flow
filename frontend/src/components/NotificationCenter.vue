<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import {
  BellIcon,
  CheckIcon,
  CheckCircleIcon,
  ArrowPathIcon,
  ArrowRightIcon,
  InboxIcon,
  XMarkIcon,
} from '@heroicons/vue/24/outline'
import SelectButton from 'primevue/selectbutton'
import { useToast } from 'primevue/usetoast'
import AppButton from './AppButton.vue'
import SurfaceButton from './primevue/SurfaceButton'
import Drawer from './primevue/Drawer'
import Message from './primevue/Message'
import { api } from '../api'
import moment from '../config/moment'
import { tooltip } from '../config/tooltip'
import { dateTime } from '../utils/format'
import { useWorkspaceStore } from '../stores/workspace'
import { connected, subscribe } from '../services/realtime'
import type { UserNotification } from '../types'

const workspace = useWorkspaceStore(),
  toast = useToast()
const trigger = ref<InstanceType<typeof AppButton>>()
const visible = ref(false),
  unreadCount = ref(0),
  items = ref<UserNotification[]>([]),
  nextCursor = ref('')
const filter = ref<'all' | 'unread'>('all'),
  loading = ref(false),
  reading = ref<string[]>([]),
  readingAll = ref(false),
  error = ref('')
const filters = [
  { label: 'Todas', value: 'all' },
  { label: 'Não lidas', value: 'unread' },
]
const bellLabel = computed(() =>
  unreadCount.value
    ? `Notificações, ${unreadCount.value} ${unreadCount.value === 1 ? 'não lida' : 'não lidas'}`
    : 'Notificações',
)
let disposed = false,
  pageEpoch = 0,
  countEpoch = 0,
  controller: AbortController | undefined
let refreshTimer: ReturnType<typeof setTimeout> | undefined,
  pollTimer: ReturnType<typeof setInterval> | undefined
let pendingDemand = ''
const relativeTime = (value: string) => moment(value).fromNow()

async function refreshCount() {
  const current = ++countEpoch
  try {
    const result = await api.notificationCount()
    if (!disposed && current === countEpoch)
      unreadCount.value = Math.max(0, Number(result.count) || 0)
  } catch (err) {
    if (!disposed && current === countEpoch) error.value = (err as Error).message
  }
}
async function loadPage(append = false) {
  if (disposed || (append && (loading.value || !nextCursor.value))) return
  controller?.abort()
  controller = new AbortController()
  const current = ++pageEpoch
  loading.value = true
  error.value = ''
  if (!append) {
    items.value = []
    nextCursor.value = ''
  }
  try {
    const page = await api.notifications(
      filter.value,
      append ? nextCursor.value : '',
      controller.signal,
    )
    if (disposed || current !== pageEpoch) return
    items.value = append
      ? [...new Map([...items.value, ...page.items].map((item) => [item.id, item])).values()]
      : page.items
    nextCursor.value = page.next_cursor
    countEpoch++
    unreadCount.value = page.unread_count
  } catch (err) {
    if (!disposed && current === pageEpoch && (err as Error).name !== 'AbortError')
      error.value = (err as Error).message
  } finally {
    if (!disposed && current === pageEpoch) loading.value = false
  }
}
function scheduleRefresh() {
  clearTimeout(refreshTimer)
  refreshTimer = setTimeout(() => {
    if (disposed) return
    if (visible.value && !readingAll.value && !reading.value.length) void loadPage()
    else void refreshCount()
  }, 200)
}
const unsubscribe = subscribe((event) => {
  if (
    event.type === 'workspace.updated' ||
    event.type === 'notifications.changed' ||
    event.type === 'connected'
  )
    scheduleRefresh()
})
async function markRead(item: UserNotification) {
  if (item.read_at || reading.value.includes(item.id) || readingAll.value) return
  reading.value = [...reading.value, item.id]
  error.value = ''
  try {
    const result = await api.readNotification(item.id)
    if (disposed) return
    controller?.abort()
    pageEpoch++
    loading.value = false
    countEpoch++
    items.value =
      filter.value === 'unread'
        ? items.value.filter((value) => value.id !== item.id)
        : items.value.map((value) =>
            value.id === item.id ? { ...value, read_at: result.read_at } : value,
          )
    unreadCount.value = Math.max(0, unreadCount.value - 1)
    await refreshCount()
  } catch (err) {
    if (!disposed) {
      error.value = (err as Error).message
      if (!visible.value)
        toast.add({
          severity: 'error',
          summary: 'Não foi possível marcar como lida',
          detail: error.value,
          life: 5000,
        })
    }
  } finally {
    if (!disposed) reading.value = reading.value.filter((id) => id !== item.id)
  }
}
async function markAllRead() {
  if (readingAll.value || reading.value.length || !unreadCount.value) return
  readingAll.value = true
  error.value = ''
  try {
    const result = await api.readAllNotifications()
    if (disposed) return
    controller?.abort()
    pageEpoch++
    loading.value = false
    countEpoch++
    items.value =
      filter.value === 'unread'
        ? []
        : items.value.map((item) => ({ ...item, read_at: item.read_at || result.read_at }))
    if (filter.value === 'unread') nextCursor.value = ''
    unreadCount.value = 0
    await refreshCount()
  } catch (err) {
    if (!disposed) error.value = (err as Error).message
  } finally {
    if (!disposed) readingAll.value = false
  }
}
function openNotification(item: UserNotification) {
  void markRead(item)
  pendingDemand = item.demand_id
  visible.value = false
}
function openPendingDemand() {
  if (disposed) return
  if (!pendingDemand) {
    trigger.value?.$el?.focus()
    return
  }
  const id = pendingDemand
  pendingDemand = ''
  void workspace.selectDemand(id)
}
function visibilityChanged() {
  if (document.visibilityState === 'visible') scheduleRefresh()
}
watch(visible, (open) => {
  if (open) void loadPage()
  else {
    controller?.abort()
    pageEpoch++
    loading.value = false
  }
})
watch(filter, () => {
  if (visible.value) void loadPage()
})
onMounted(() => {
  void refreshCount()
  document.addEventListener('visibilitychange', visibilityChanged)
  pollTimer = setInterval(() => {
    if (!connected.value && document.visibilityState === 'visible') scheduleRefresh()
  }, 30_000)
})
onUnmounted(() => {
  disposed = true
  pageEpoch++
  countEpoch++
  pendingDemand = ''
  controller?.abort()
  clearTimeout(refreshTimer)
  clearInterval(pollTimer)
  unsubscribe()
  document.removeEventListener('visibilitychange', visibilityChanged)
})
</script>

<template>
  <AppButton
    ref="trigger"
    v-tooltip.bottom="tooltip('Central de notificações')"
    class="notification-trigger"
    variant="ghost"
    :icon="BellIcon"
    :aria-label="bellLabel"
    aria-haspopup="dialog"
    :aria-expanded="visible"
    aria-controls="notification-center"
    :badge="unreadCount ? (unreadCount > 99 ? '99+' : String(unreadCount)) : undefined"
    badge-severity="danger"
    @click="visible = true"
  />
  <Drawer
    id="notification-center"
    v-model:visible="visible"
    position="right"
    header="Notificações"
    role="dialog"
    aria-label="Central de notificações"
    class="notification-drawer"
    :style="{ width: 'min(440px,100vw)' }"
    :close-button-props="{ 'aria-label': 'Fechar notificações' }"
    @after-hide="openPendingDemand"
  >
    <template #closebutton="{ closeCallback }">
      <AppButton
        v-tooltip.left="tooltip('Fechar notificações')"
        variant="ghost"
        :icon="XMarkIcon"
        aria-label="Fechar notificações"
        autofocus
        @click="closeCallback"
      />
    </template>
    <div class="notification-heading">
      <p aria-live="polite">
        {{
          unreadCount
            ? `${unreadCount} ${unreadCount === 1 ? 'notificação não lida' : 'notificações não lidas'}`
            : 'Você está em dia com suas notificações.'
        }}
      </p>
      <AppButton
        label="Marcar todas como lidas"
        variant="ghost"
        size="small"
        :icon="CheckCircleIcon"
        :disabled="!unreadCount || Boolean(reading.length)"
        :loading="readingAll"
        @click="markAllRead"
      />
    </div>
    <div class="notification-toolbar">
      <SelectButton
        v-model="filter"
        :options="filters"
        option-label="label"
        option-value="value"
        :allow-empty="false"
        aria-label="Filtrar notificações"
        :disabled="readingAll"
      />
      <AppButton
        v-tooltip.left="tooltip('Atualizar notificações')"
        :icon="ArrowPathIcon"
        variant="ghost"
        size="small"
        aria-label="Atualizar notificações"
        :loading="loading"
        @click="loadPage()"
      />
    </div>
    <Message
      v-if="error"
      severity="error"
      :closable="false"
      class="mb-4"
    >
      <div class="grid gap-2">
        <span>{{ error }}</span>
        <AppButton
          label="Tentar novamente"
          variant="secondary"
          size="small"
          @click="loadPage()"
        />
      </div>
    </Message>
    <p
      v-if="loading && !items.length"
      class="notification-loading"
      role="status"
    >
      Carregando notificações…
    </p>
    <div
      v-else-if="!items.length && !nextCursor && !error"
      class="notification-empty"
    >
      <InboxIcon aria-hidden="true" />
      <h3>
        {{
          filter === 'unread' ? 'Nenhuma notificação não lida' : 'Nenhuma notificação por enquanto'
        }}
      </h3>
      <p>Suas atribuições, mudanças de etapa e resultados de QA aparecerão aqui.</p>
    </div>
    <ul
      v-if="items.length"
      class="notification-list"
      aria-label="Suas notificações"
      :aria-busy="loading"
    >
      <li
        v-for="item in items"
        :key="item.id"
        class="notification-item"
        :class="{ 'is-unread': !item.read_at }"
      >
        <SurfaceButton
          class="notification-open"
          text
          severity="secondary"
          :aria-label="`Abrir notificação: ${item.title}. ${item.message}`"
          :disabled="readingAll"
          @click="openNotification(item)"
        >
          <span class="notification-item-heading">
            <strong>{{ item.title }}</strong>
            <span
              v-if="!item.read_at"
              class="notification-unread-dot"
              aria-label="Não lida"
            ></span>
          </span>
          <span class="notification-message">{{ item.message }}</span>
          <time
            v-tooltip.top="tooltip(dateTime(item.created_at))"
            :datetime="item.created_at"
          >
            {{ relativeTime(item.created_at) }}
          </time>
          <span class="notification-detail-link">
            Ver demanda
            <ArrowRightIcon
              aria-hidden="true"
              class="app-icon"
            />
          </span>
        </SurfaceButton>
        <AppButton
          v-if="!item.read_at"
          v-tooltip.left="tooltip('Marcar como lida')"
          variant="ghost"
          size="small"
          :icon="CheckIcon"
          class="notification-mark-read"
          :loading="reading.includes(item.id)"
          :disabled="readingAll"
          :aria-label="`Marcar como lida: ${item.title}`"
          @click="markRead(item)"
        />
      </li>
    </ul>
    <AppButton
      v-if="nextCursor"
      label="Carregar mais notificações"
      variant="secondary"
      fluid
      :loading="loading"
      :disabled="readingAll"
      @click="loadPage(true)"
    />
  </Drawer>
</template>

<style>
.notification-trigger.app-button .p-badge {
  position: absolute;
  top: 0;
  right: -2px;
  min-width: 18px;
  height: 18px;
  padding: 0 4px;
  font-size: 10px;
  line-height: 18px;
}
.notification-drawer .p-drawer-content {
  padding-top: 0;
}
.notification-heading {
  padding-bottom: 18px;
  border-bottom: 1px solid var(--sf-border);
}
.notification-heading p {
  margin-bottom: 12px;
  color: var(--sf-muted);
  font-size: 12px;
  line-height: 1.6;
}
.notification-heading .app-button {
  margin-left: -10px;
}
.notification-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 18px 0;
}
.notification-toolbar .p-togglebutton {
  font-size: 12px;
}
.notification-list {
  display: grid;
  gap: 10px;
  list-style: none;
  margin: 0 0 18px;
  padding: 0;
}
.notification-item {
  position: relative;
  border: 1px solid var(--sf-border);
  border-radius: 12px;
  background: var(--sf-soft);
  overflow: hidden;
}
.notification-item.is-unread {
  border-color: color-mix(in srgb, var(--sf-accent) 35%, var(--sf-border));
  background: color-mix(in srgb, var(--sf-accent-soft) 40%, var(--sf-surface));
}
.notification-open.p-button {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 9px;
  width: 100%;
  padding: 17px;
  border: 0;
  border-radius: 11px;
  background: transparent;
  color: var(--sf-text);
  text-align: left;
  white-space: normal;
}
.notification-item.is-unread .notification-open {
  padding-right: 48px;
}
.notification-open.p-button:enabled:hover {
  background: var(--sf-accent-soft);
  color: var(--sf-text);
}
.notification-open:focus-visible {
  outline: 2px solid var(--sf-accent);
  outline-offset: -3px;
}
.notification-item-heading {
  display: flex;
  align-items: center;
  gap: 10px;
}
.notification-item-heading strong {
  font-size: 13px;
  font-weight: 600;
  line-height: 1.5;
  overflow-wrap: anywhere;
}
.notification-unread-dot {
  width: 6px;
  height: 6px;
  flex-shrink: 0;
  border-radius: 50%;
  background: var(--sf-accent);
}
.notification-message {
  color: var(--sf-muted);
  font-size: 12px;
  line-height: 1.7;
  overflow-wrap: anywhere;
}
.notification-open time {
  color: var(--sf-muted);
  font-size: 10px;
  font-weight: 500;
}
.notification-detail-link {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--sf-accent);
  font-size: 11px;
  font-weight: 600;
}
.notification-detail-link .app-icon {
  width: 13px;
  height: 13px;
}
.notification-item .notification-mark-read {
  position: absolute !important;
  right: 8px;
  top: 10px;
}
.notification-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 56px 20px;
  text-align: center;
}
.notification-empty > svg {
  width: 40px;
  height: 40px;
  color: var(--sf-accent);
  opacity: 0.7;
}
.notification-empty h3 {
  font-size: 14px;
  font-weight: 600;
  color: var(--sf-text);
}
.notification-empty p,
.notification-loading {
  color: var(--sf-muted);
  font-size: 12px;
  line-height: 1.7;
}
.notification-loading {
  padding: 24px 0;
  text-align: center;
}
</style>
