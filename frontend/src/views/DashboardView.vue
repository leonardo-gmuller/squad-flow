<script setup lang="ts">
import AppButton from '../components/AppButton.vue'
import DashboardReport from '../components/DashboardReport.vue'
import DataTable from '../components/primevue/DataTable'
import {
  ArrowRightIcon,
  ArrowTrendingUpIcon,
  ArrowUpRightIcon,
  CheckCircleIcon,
  ClipboardDocumentCheckIcon,
  CodeBracketIcon,
  DocumentDuplicateIcon,
  PlusIcon,
  SparklesIcon,
  Square3Stack3DIcon,
} from '@heroicons/vue/24/outline'
import { tooltip } from '../config/tooltip'
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import Column from 'primevue/column'
import StatCard from '../components/StatCard.vue'
import KindDistributionChart from '../components/KindDistributionChart.vue'
import DeliveryProgressChart from '../components/DeliveryProgressChart.vue'
import { useWorkspaceStore } from '../stores/workspace'
import { kindLabel, kindStyle, statusLabel } from '../config/catalog'
import { dateLabel, dateTimestamp, demandCode, initials } from '../utils/format'
const store = useWorkspaceStore()
const recent = computed(() =>
  [...store.items]
    .sort((a, b) => dateTimestamp(b.updated_at) - dateTimestamp(a.updated_at))
    .slice(0, 7),
)
</script>
<template>
  <div class="page-heading">
    <div>
      <div class="eyebrow">
        <span class="eyebrow-line"></span>
        DESENVOLVIMENTO DE SOFTWARE
      </div>
      <h1>Visão geral<span class="title-period">.</span></h1>
      <p>Acompanhe o que está em desenvolvimento, em teste e pronto para entrega.</p>
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
  <div class="welcome-banner">
    <div class="welcome-copy">
      <div class="banner-kicker">
        <SparklesIcon style="width: 14px; height: 14px; flex-shrink: 0" />
        WORKSPACE / DESENVOLVIMENTO
      </div>
      <h2>
        Todo o trabalho da squad.
        <br />
        <span>Uma única visão.</span>
      </h2>
      <p>
        Da demanda à homologação, com contexto, responsáveis e rastreabilidade para tomar decisões
        melhores.
      </p>
      <RouterLink
        to="/board"
        class="welcome-link"
      >
        Explorar Kanban
        <ArrowRightIcon style="width: 16px; height: 16px; flex-shrink: 0" />
      </RouterLink>
    </div>
    <div
      class="banner-art"
      aria-hidden="true"
    >
      <div class="art-orbit orbit-one"></div>
      <div class="art-orbit orbit-two"></div>
      <div class="art-card art-back">
        <div class="art-line sm"></div>
        <div class="art-line"></div>
        <div class="art-progress"></div>
      </div>
      <div class="art-card art-front">
        <span class="art-overline">SPRINT / ENTREGA</span>
        <span class="art-check">
          <CheckCircleIcon style="width: 20px; height: 20px; flex-shrink: 0" />
        </span>
        <div class="art-line"></div>
        <div class="art-line short"></div>
        <div class="art-bar"><i></i></div>
      </div>
    </div>
  </div>
  <section class="stat-grid">
    <StatCard
      label="Demandas registradas"
      :value="store.metrics.total"
      helper="Todos os tipos de entrega"
      :icon="DocumentDuplicateIcon"
      tone="violet"
    />
    <StatCard
      label="Em aberto"
      :value="store.metrics.active"
      helper="Aguardando conclusão"
      :icon="CodeBracketIcon"
      tone="blue"
    />
    <StatCard
      label="Em homologação"
      :value="store.metrics.in_qa"
      helper="Na fila da qualidade"
      :icon="ClipboardDocumentCheckIcon"
      tone="amber"
    />
    <StatCard
      label="Entregues"
      :value="store.metrics.delivered"
      helper="Concluídas pelo fluxo"
      :icon="CheckCircleIcon"
      tone="green"
    />
  </section>
  <div class="dashboard-grid">
    <section class="surface-card analytics-card">
      <div class="card-heading">
        <div>
          <span class="overline">VOLUME DE TRABALHO</span>
          <h2>Distribuição por esteira</h2>
          <p>Entregas registradas, independentemente da OS de origem.</p>
        </div>
        <RouterLink
          to="/reports"
          class="subtle-link"
        >
          Ver relatório
          <ArrowUpRightIcon style="width: 15px; height: 15px; flex-shrink: 0" />
        </RouterLink>
      </div>
      <KindDistributionChart
        :counts="store.metrics.by_kind"
        title="Distribuição por esteira"
      />
      <div class="card-footnote">
        <Square3Stack3DIcon style="width: 15px; height: 15px; flex-shrink: 0" />
        As OS agrupadoras não são contadas como entregas.
      </div>
    </section>
    <section class="surface-card delivery-card">
      <div class="card-heading">
        <div>
          <span class="overline">EVOLUÇÃO DA SQUAD</span>
          <h2>Progresso das entregas</h2>
          <p>Visão consolidada do fluxo atual.</p>
        </div>
        <span class="neutral-chip">Geral</span>
      </div>
      <div class="progress-display">
        <DeliveryProgressChart
          :delivered="store.metrics.delivered"
          :active="store.metrics.active"
          :total="store.metrics.total"
        />
        <div class="progress-legend">
          <div>
            <span class="legend-marker green"></span>
            Concluídas
            <strong>{{ store.metrics.delivered }}</strong>
          </div>
          <div>
            <span class="legend-marker slate"></span>
            Em aberto
            <strong>{{ store.metrics.active }}</strong>
          </div>
          <div>
            <span class="legend-marker purple"></span>
            Devoluções
            <strong>{{ store.metrics.returned }}</strong>
          </div>
        </div>
      </div>
      <div class="first-pass-note">
        <ArrowTrendingUpIcon style="width: 17px; height: 17px; flex-shrink: 0" />
        <span>
          <strong>{{ store.metrics.first_pass_percent.toFixed(0) }}%</strong>
          de aprovação sem devolução
        </span>
      </div>
    </section>
  </div>
  <DashboardReport />
  <section class="surface-card recent-card">
    <div class="card-heading">
      <div>
        <span class="overline">ACOMPANHAMENTO</span>
        <h2>Atividade recente</h2>
        <p>As últimas entregas movimentadas pela squad.</p>
      </div>
      <RouterLink
        to="/tickets"
        class="subtle-link"
      >
        Todas as demandas
        <ArrowRightIcon style="width: 16px; height: 16px; flex-shrink: 0" />
      </RouterLink>
    </div>
    <DataTable
      :value="recent"
      data-key="id"
      size="small"
      scrollable
      :pt="{ bodyRow: { class: 'cursor-pointer' } }"
      @row-click="(event) => store.selectDemand(event.data.id)"
    >
      <Column
        field="title"
        header="Demanda"
        class="min-w-64"
      >
        <template #body="{ data }">
          <div class="ticket-row-name">
            <strong>{{ data.title }}</strong>
            <small>{{ demandCode(data.number) }} · {{ data.reference || 'Sem referência' }}</small>
          </div>
        </template>
      </Column>
      <Column
        field="kind"
        header="Esteira"
      >
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
        field="assignee"
        header="Responsável"
        class="min-w-40"
      >
        <template #body="{ data }">
          <div class="flex items-center gap-2">
            <span class="avatar-mini">{{ initials(data.assignee) }}</span>
            <span>{{ data.assignee || 'Não atribuído' }}</span>
          </div>
        </template>
      </Column>
      <Column
        field="status"
        header="Status"
      >
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
        field="updated_at"
        header="Atualização"
        class="whitespace-nowrap"
      >
        <template #body="{ data }">{{ dateLabel(data.updated_at) }}</template>
      </Column>
      <Column>
        <template #body="{ data }">
          <AppButton
            v-tooltip.top="tooltip('Abrir detalhes da entrega')"
            variant="ghost"
            size="small"
            :aria-label="`Abrir ${data.title}`"
            :icon="ArrowUpRightIcon"
            @click.stop="store.selectDemand(data.id)"
          />
        </template>
      </Column>
      <template #empty>
        <div class="empty-inline">Nenhuma demanda ainda. Crie a primeira entrega para iniciar.</div>
      </template>
    </DataTable>
  </section>
</template>
