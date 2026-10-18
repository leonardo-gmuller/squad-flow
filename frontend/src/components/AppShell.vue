<script setup lang="ts">
import AppButton from './AppButton.vue'
import AppLogo from './AppLogo.vue'
import NotificationCenter from './NotificationCenter.vue'
import Message from './primevue/Message'
import {
  ArrowRightOnRectangleIcon,
  Bars3Icon,
  ChartBarIcon,
  ChevronDoubleLeftIcon,
  ChevronDoubleRightIcon,
  ChevronRightIcon,
  ClipboardDocumentCheckIcon,
  CodeBracketIcon,
  HomeIcon,
  MagnifyingGlassIcon,
  MoonIcon,
  Squares2X2Icon,
  SunIcon,
  TagIcon,
  UserCircleIcon,
  UsersIcon,
  ViewColumnsIcon,
  XMarkIcon,
} from '@heroicons/vue/24/outline'
import { tooltip } from '../config/tooltip'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { useToast } from 'primevue/usetoast'
import InputText from 'primevue/inputtext'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import { useWorkspaceStore } from '../stores/workspace'
import { useAuthStore } from '../stores/auth'
import { useReportJobsStore } from '../stores/reportJobs'
import { startRealtime, stopRealtime, subscribe } from '../services/realtime'
import { initials } from '../utils/format'
import CreateDemandDialog from './CreateDemandDialog.vue'
import CreatePackageDialog from './CreatePackageDialog.vue'
import TaskDrawer from './TaskDrawer.vue'

const store = useWorkspaceStore(),
  auth = useAuthStore(),
  route = useRoute(),
  router = useRouter(),
  toast = useToast()
const reportJobs = useReportJobsStore()
watch(
  () => auth.user?.id,
  (userID) => {
    if (userID) reportJobs.start(userID)
    else reportJobs.stop()
  },
  { immediate: true },
)
watch(
  () => reportJobs.jobs.map((job) => ({ id: job.id, status: job.status })),
  (jobs, previous) => {
    for (const job of jobs) {
      const old = previous?.find((row) => row.id === job.id)
      if (!old || old.status === job.status) continue
      if (job.status === 'completed')
        toast.add({
          severity: 'success',
          summary: 'Seu relatório está pronto',
          detail: 'Abra em Relatórios para visualizar, apresentar ou baixar o HTML.',
          life: 9000,
        })
      if (job.status === 'failed')
        toast.add({
          severity: 'error',
          summary: 'Não foi possível concluir o relatório',
          detail: 'Confira o histórico em Relatórios para tentar novamente.',
          life: 9000,
        })
    }
  },
)
const mobileMenu = ref(false),
  search = ref(''),
  darkMode = ref(localStorage.getItem('squadflow:dark') !== 'false')
const sidebarCollapsed = ref(localStorage.getItem('squadflow:sidebar-collapsed') === 'true')
const nav = computed(() => [
  { label: 'Visão geral', to: '/dashboard', icon: HomeIcon },
  { label: 'Demandas', to: '/tickets', icon: Squares2X2Icon },
  { label: 'Kanban', to: '/board', icon: ViewColumnsIcon },
  { label: 'Versões', to: '/releases', icon: TagIcon },
  { label: 'OS e iniciativas', to: '/packages', icon: CodeBracketIcon },
  { label: 'Central de QA', to: '/qa', icon: ClipboardDocumentCheckIcon },
  { label: 'Relatórios', to: '/reports', icon: ChartBarIcon },
  { label: 'Minha conta', to: '/account', icon: UserCircleIcon },
  ...(auth.user?.role === 'admin'
    ? [
        { label: 'Esteiras', to: '/workstreams', icon: Squares2X2Icon },
        { label: 'Usuários', to: '/users', icon: UsersIcon },
      ]
    : []),
])
const pageTitle = computed(() => String(route.meta.title || 'SquadFlow'))
watch(
  () => route.fullPath,
  () => {
    mobileMenu.value = false
  },
)
watch(
  () => route.path,
  (path, previous) => {
    store.setBoardMode(path === '/board')
    if (previous === '/board' && path !== '/board') void store.load().catch(() => {})
  },
)
watch(
  darkMode,
  (value) => {
    document.documentElement.classList.toggle('app-dark', value)
    localStorage.setItem('squadflow:dark', String(value))
  },
  { immediate: true },
)
watch(sidebarCollapsed, (value) =>
  localStorage.setItem('squadflow:sidebar-collapsed', String(value)),
)
function keyboard(event: KeyboardEvent) {
  if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'k') {
    event.preventDefault()
    document.getElementById('global-search')?.focus()
  }
}
let refreshTimer: ReturnType<typeof setTimeout> | undefined
const unsubscribe = subscribe((event) => {
  if (event.type === 'workspace.updated') {
    clearTimeout(refreshTimer)
    refreshTimer = setTimeout(() => {
      void store.refresh().catch(() => {})
    }, 200)
  }
})
onMounted(async () => {
  store.setBoardMode(route.path === '/board')
  window.addEventListener('keydown', keyboard)
  startRealtime()
  try {
    await store.load()
  } catch (error) {
    toast.add({
      severity: 'error',
      summary: 'Não foi possível carregar o workspace',
      detail: (error as Error).message,
      life: 6500,
    })
  }
})
onUnmounted(() => {
  window.removeEventListener('keydown', keyboard)
  unsubscribe()
  clearTimeout(refreshTimer)
  stopRealtime()
  reportJobs.stop()
  store.reset()
})
function globalSearch() {
  if (!search.value.trim()) return
  void router.push({ path: '/tickets', query: { q: search.value.trim() } })
  search.value = ''
}
async function logout() {
  try {
    await auth.logout()
    await router.replace('/login')
  } catch (error) {
    toast.add({
      severity: 'error',
      summary: 'Não foi possível sair',
      detail: (error as Error).message,
      life: 4000,
    })
  }
}
</script>
<template>
  <div
    class="shell"
    :class="{ 'is-board': route.path === '/board', 'sidebar-expanded': !sidebarCollapsed }"
  >
    <Transition name="backdrop">
      <div
        v-if="mobileMenu"
        class="mobile-backdrop"
        @click="mobileMenu = false"
      ></div>
    </Transition>
    <aside
      id="workspace-sidebar"
      class="shell-sidebar"
      :class="{ 'is-open': mobileMenu }"
    >
      <div class="brand-section">
        <RouterLink
          to="/dashboard"
          aria-label="SquadFlow - Início"
        >
          <AppLogo />
        </RouterLink>
        <AppButton
          v-tooltip.right="tooltip('Fechar navegação')"
          variant="ghost"
          class="mobile-nav-close"
          aria-label="Fechar navegação"
          :icon="XMarkIcon"
          @click="mobileMenu = false"
        />
      </div>
      <div class="sidebar-controls">
        <AppButton
          v-tooltip.right="
            tooltip(sidebarCollapsed ? 'Mostrar nomes das telas' : 'Mostrar somente ícones')
          "
          variant="ghost"
          class="sidebar-toggle"
          size="small"
          label="Recolher menu"
          :aria-label="sidebarCollapsed ? 'Expandir menu lateral' : 'Recolher menu lateral'"
          :aria-expanded="!sidebarCollapsed"
          aria-controls="sidebar-navigation"
          :icon="sidebarCollapsed ? ChevronDoubleRightIcon : ChevronDoubleLeftIcon"
          @click="sidebarCollapsed = !sidebarCollapsed"
        />
      </div>
      <nav
        id="sidebar-navigation"
        class="sidebar-links sidebar-main-nav"
        aria-label="Navegação principal"
      >
        <RouterLink
          v-for="item in nav"
          :key="item.to"
          v-tooltip.right="tooltip(item.label, !sidebarCollapsed)"
          :to="item.to"
          :data-testid="item.to === '/packages' ? 'nav-packages' : undefined"
          class="sidebar-link"
          :aria-label="item.label"
        >
          <component
            :is="item.icon"
            style="width: 22px; height: 22px; flex-shrink: 0"
          />
          <span>{{ item.label }}</span>
        </RouterLink>
      </nav>
      <div class="flex-1"></div>
      <div class="sidebar-profile">
        <span
          v-tooltip.right="tooltip(auth.user?.name, !sidebarCollapsed)"
          class="profile-avatar"
          :aria-label="auth.user?.name"
        >
          {{ initials(auth.user?.name) }}
        </span>
        <span class="min-w-0 flex-1">
          <strong>{{ auth.user?.name }}</strong>
          <small>
            {{ auth.user?.role === 'admin' ? 'Administrador' : 'Workspace compartilhado' }}
          </small>
        </span>
        <div class="sidebar-logout">
          <AppButton
            v-tooltip.right="tooltip('Sair')"
            variant="ghost"
            class="logout-button"
            size="small"
            aria-label="Sair"
            :icon="ArrowRightOnRectangleIcon"
            @click="logout"
          />
        </div>
      </div>
    </aside>
    <div class="shell-main">
      <header class="shell-topbar">
        <div class="flex min-w-0 items-center gap-3">
          <AppButton
            v-tooltip.bottom="tooltip('Abrir navegação')"
            variant="ghost"
            class="menu-trigger"
            aria-label="Abrir navegação"
            :aria-expanded="mobileMenu"
            aria-controls="workspace-sidebar"
            :icon="Bars3Icon"
            @click="mobileMenu = true"
          />
          <div class="breadcrumb">
            <span>Workspace</span>
            <ChevronRightIcon style="width: 14px; height: 14px; flex-shrink: 0" />
            <strong>{{ pageTitle }}</strong>
          </div>
        </div>
        <div class="topbar-actions">
          <form
            class="hidden w-52 lg:block xl:w-64"
            @submit.prevent="globalSearch"
          >
            <IconField>
              <InputIcon><MagnifyingGlassIcon class="app-icon" /></InputIcon>
              <InputText
                id="global-search"
                v-model="search"
                v-tooltip.bottom.focus="
                  tooltip('Busque demandas por título, evento ou referência e pressione Enter.')
                "
                size="small"
                placeholder="Buscar demandas..."
                aria-label="Busca global"
                fluid
              />
            </IconField>
          </form>
          <AppButton
            v-tooltip.bottom="tooltip(darkMode ? 'Usar tema claro' : 'Usar tema escuro')"
            variant="ghost"
            :aria-label="darkMode ? 'Tema claro' : 'Tema escuro'"
            :icon="darkMode ? SunIcon : MoonIcon"
            @click="darkMode = !darkMode"
          />
          <NotificationCenter
            v-if="auth.user?.role === 'member'"
            :key="auth.user.id"
          />
        </div>
      </header>
      <main class="page-container">
        <Message
          v-if="store.error"
          severity="error"
          :closable="false"
          class="mb-5"
        >
          <div class="flex flex-wrap items-center gap-3">
            <span>
              <strong>Erro ao carregar dados:</strong>
              {{ store.error }}
            </span>
            <AppButton
              variant="secondary"
              label="Tentar novamente"
              size="small"
              @click="store.refresh().catch(() => {})"
            />
          </div>
        </Message>
        <RouterView v-slot="{ Component, route: viewRoute }">
          <Transition
            name="page"
            mode="out-in"
          >
            <div
              :key="viewRoute.path"
              class="route-page"
            >
              <component :is="Component" />
            </div>
          </Transition>
        </RouterView>
        <footer class="page-footer">
          <span>SquadFlow · Gestão do desenvolvimento de software</span>
        </footer>
      </main>
    </div>
    <CreateDemandDialog />
    <CreatePackageDialog />
    <TaskDrawer />
  </div>
</template>
