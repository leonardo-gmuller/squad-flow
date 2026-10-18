<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import Dialog from 'primevue/dialog'
import Message from './primevue/Message'
import AppButton from './AppButton.vue'
import ReportRenderer from './ReportRenderer.vue'
import { DocumentTextIcon } from '@heroicons/vue/24/outline'
import { api } from '../api'
import { useReportJobsStore } from '../stores/reportJobs'
import type { GeneratedReport } from '../types'
import { dateTime } from '../utils/format'
const jobs = useReportJobsStore(),
  latest = ref<GeneratedReport | null>(null),
  opened = ref<GeneratedReport | null>(null),
  visible = ref(false),
  loading = ref(false),
  error = ref('')
async function load() {
  try {
    latest.value = (await api.reports())[0] || null
  } catch {
    // The optional report preview must not prevent the dashboard from loading.
  }
}
onMounted(load)
watch(
  () =>
    jobs.jobs
      .filter((job) => job.status === 'completed')
      .map((job) => job.id)
      .join(','),
  load,
)
async function open() {
  if (!latest.value) return
  loading.value = true
  error.value = ''
  try {
    opened.value = await api.report(latest.value.id)
    visible.value = true
  } catch (err) {
    error.value = (err as Error).message
  } finally {
    loading.value = false
  }
}
</script>
<template>
  <section
    v-if="latest || jobs.active"
    class="surface-card dashboard-report"
    aria-label="Seu relatório no dashboard"
  >
    <div class="card-heading">
      <div>
        <span class="overline">SUA ANÁLISE COM IA</span>
        <h2>{{ latest?.content.title || 'Preparando seu próximo relatório' }}</h2>
        <p>{{ latest?.content.summary || jobs.active?.progress_message }}</p>
        <small v-if="latest">{{ dateTime(latest.created_at) }}</small>
      </div>
      <AppButton
        v-if="latest"
        variant="secondary"
        label="Abrir relatório no dashboard"
        :icon="DocumentTextIcon"
        :loading="loading"
        @click="open"
      />
    </div>
    <p
      v-if="jobs.active"
      class="mt-3 text-sm"
      role="status"
    >
      {{ jobs.active.progress_message }} · Você pode continuar trabalhando.
    </p>
    <Message
      v-if="error"
      severity="error"
      :closable="false"
    >
      {{ error }}
    </Message>
    <RouterLink
      to="/reports"
      class="subtle-link mt-3 inline-flex"
    >
      Ver análises e solicitações
    </RouterLink>
    <Dialog
      v-model:visible="visible"
      modal
      maximizable
      :header="latest?.content.title"
      :style="{ width: 'min(96vw,1400px)' }"
    >
      <ReportRenderer
        v-if="opened"
        :report="opened"
      />
    </Dialog>
  </section>
</template>
<style scoped>
.dashboard-report {
  padding: 1.5rem;
  margin: 1.5rem 0;
}
.dashboard-report p {
  max-width: 70ch;
}
small {
  color: var(--p-text-muted-color);
}
@media (max-width: 700px) {
  .card-heading {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
