<script setup lang="ts">
import AppButton from '../components/AppButton.vue'
import SurfaceButton from '../components/primevue/SurfaceButton'
import {
  CheckCircleIcon,
  ChevronRightIcon,
  ListBulletIcon,
  MagnifyingGlassIcon,
  PlusIcon,
  Square3Stack3DIcon,
  ViewColumnsIcon,
} from '@heroicons/vue/24/outline'
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import InputText from 'primevue/inputtext'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import ProgressBar from 'primevue/progressbar'
import { useWorkspaceStore } from '../stores/workspace'
import { demandCode, percent } from '../utils/format'
import { kindLabel, kindStyle, statusLabel } from '../config/catalog'
const store = useWorkspaceStore(),
  route = useRoute(),
  router = useRouter(),
  query = ref('')
const selection = computed({
  get: () => String(route.query.os || '') || store.packages[0]?.id || '',
  set: (v: string) => {
    router.replace({ path: '/packages', query: { os: v } })
  },
})
const list = computed(() =>
  store.packages.filter((p) =>
    `${p.title} ${p.reference} ${p.initiative}`.toLowerCase().includes(query.value.toLowerCase()),
  ),
)
const selected = computed(() => store.packages.find((p) => p.id === selection.value))
const tasks = computed(() => (selected.value ? store.itemsInPackage(selected.value.id) : []))
function createItem() {
  if (selected.value) store.openCreate(selected.value.id)
}
</script>
<template>
  <div class="page-heading">
    <div>
      <div class="eyebrow">
        <span class="eyebrow-line"></span>
        PLANEJAMENTO / ORGANIZAÇÃO
      </div>
      <h1>OS e iniciativas<span class="title-period">.</span></h1>
      <p>Organize entregas individuais ou compartilhadas, sem perder a visão do todo.</p>
    </div>
    <div class="heading-actions">
      <AppButton
        variant="secondary"
        label="Nova entrega"
        data-testid="new-demand"
        :icon="ListBulletIcon"
        @click="store.openCreate()"
      />
      <AppButton
        label="Nova OS"
        data-testid="new-package"
        :icon="PlusIcon"
        @click="store.createPackageVisible = true"
      />
    </div>
  </div>
  <div class="package-intro">
    <span class="package-intro-icon">
      <Square3Stack3DIcon style="width: 23px; height: 23px; flex-shrink: 0" />
    </span>
    <div>
      <strong>Uma OS pode reunir várias entregas</strong>
      <p>
        Distribua os itens entre desenvolvedores e QA. Cada entrega mantém seu próprio status, prazo
        e histórico.
      </p>
    </div>
    <span class="package-intro-counter">{{ store.packages.length }} OS cadastradas</span>
  </div>
  <div class="packages-grid">
    <section class="package-side surface-card">
      <div class="package-side-title">
        <div>
          <span class="overline">INICIATIVAS & AGRUPAMENTOS</span>
          <h2>Ordens de serviço</h2>
        </div>
        <span class="light-count">{{ list.length }}</span>
      </div>
      <IconField class="mx-4 mb-4">
        <InputIcon><MagnifyingGlassIcon class="app-icon" /></InputIcon>
        <InputText
          v-model="query"
          placeholder="Buscar OS ou iniciativa"
          fluid
          aria-label="Buscar OS"
        />
      </IconField>
      <div class="packages-list">
        <SurfaceButton
          v-for="pkg in list"
          :key="pkg.id"
          severity="secondary"
          outlined
          class="package-list-item"
          :class="{ selected: selection === pkg.id }"
          :aria-label="`${pkg.reference} ${pkg.title}`"
          @click="selection = pkg.id"
        >
          <div class="flex items-center justify-between">
            <span class="package-ref">{{ pkg.reference || 'Sem número externo' }}</span>
            <ChevronRightIcon style="width: 15px; height: 15px; flex-shrink: 0" />
          </div>
          <strong>{{ pkg.title }}</strong>
          <span class="package-initiative">{{ pkg.initiative || 'Melhoria contínua' }}</span>
          <ProgressBar
            :value="percent(pkg.done, pkg.total)"
            :show-value="false"
            class="package-progress"
            :aria-label="`Progresso da OS ${pkg.reference || pkg.title}`"
          />
          <div class="package-list-footer">
            <span>{{ pkg.done }}/{{ pkg.total }} entregas</span>
            <span>{{ percent(pkg.done, pkg.total) }}%</span>
          </div>
        </SurfaceButton>
        <div
          v-if="!list.length"
          class="empty-inline"
        >
          Nenhuma OS encontrada.
        </div>
      </div>
    </section>
    <section
      v-if="selected"
      class="package-main surface-card"
    >
      <div class="package-detail-heading">
        <span class="overline">DETALHAMENTO DA OS</span>
        <span class="package-ref big">{{ selected.reference || 'Sem referência externa' }}</span>
      </div>
      <h2 class="package-detail-title">{{ selected.title }}</h2>
      <p class="package-detail-description">
        {{ selected.description || 'Sem descrição cadastrada para esta OS.' }}
      </p>
      <div class="package-mini-tags">
        <span
          class="kind-pill"
          :style="kindStyle(selected.kind)"
        >
          {{ kindLabel(selected.kind) }}
        </span>
        <span class="neutral-chip">{{ selected.initiative || 'Sem iniciativa específica' }}</span>
      </div>
      <div class="package-metrics">
        <div>
          <span>Total de entregas</span>
          <strong>{{ selected.total }}</strong>
        </div>
        <div>
          <span>Em QA</span>
          <strong>{{ selected.in_qa }}</strong>
        </div>
        <div>
          <span>Concluídas</span>
          <strong>{{ selected.done }}</strong>
        </div>
        <div>
          <span>Progresso</span>
          <strong>{{ percent(selected.done, selected.total) }}%</strong>
        </div>
      </div>
      <div class="package-items-header">
        <div>
          <span class="overline">DIVISÃO DO TRABALHO</span>
          <h3>
            Entregas vinculadas
            <span class="light-count">{{ tasks.length }}</span>
          </h3>
        </div>
        <AppButton
          label="Adicionar item"
          data-testid="new-package-item"
          size="small"
          :icon="PlusIcon"
          @click="createItem"
        />
      </div>
      <div
        v-if="!tasks.length"
        class="empty-inline tall"
      >
        Esta OS ainda não tem entregas. Adicione um item e atribua um desenvolvedor.
      </div>
      <SurfaceButton
        v-for="task in tasks"
        :key="task.id"
        severity="secondary"
        outlined
        class="package-task"
        @click="store.selectDemand(task.id)"
      >
        <span class="package-task-icon">
          <CheckCircleIcon
            v-if="task.status === 'done'"
            style="width: 18px; height: 18px; flex-shrink: 0"
          />
          <Square3Stack3DIcon
            v-else
            style="width: 18px; height: 18px; flex-shrink: 0"
          />
        </span>
        <span class="flex-1 min-w-0">
          <span class="package-task-id">{{ demandCode(task.number) }}</span>
          <strong>{{ task.title }}</strong>
          <small>
            {{ task.assignee || 'Não atribuído' }} · {{ task.reviewer || 'QA não atribuído' }}
          </small>
        </span>
        <span
          class="status-pill"
          :class="task.status"
        >
          <span></span>
          {{ statusLabel(task.status) }}
        </span>
        <ChevronRightIcon
          style="width: 15px; height: 15px; flex-shrink: 0"
          class="text-slate-400"
        />
      </SurfaceButton>
    </section>
    <section
      v-else
      class="package-main surface-card flex flex-col items-center justify-center text-center"
    >
      <ViewColumnsIcon
        style="width: 42px; height: 42px; flex-shrink: 0"
        class="mb-3 text-slate-300"
      />
      <h2>Comece pela primeira OS</h2>
      <p class="my-3 text-sm text-slate-500">
        Cadastre a primeira ordem de serviço e distribua os itens pelo time.
      </p>
      <AppButton
        label="Criar OS"
        :icon="PlusIcon"
        @click="store.createPackageVisible = true"
      />
    </section>
  </div>
</template>
