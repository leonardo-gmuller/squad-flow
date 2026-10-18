<script setup lang="ts">
import { ref } from 'vue'
import Tag from 'primevue/tag'
import ProgressBar from 'primevue/progressbar'
import Message from './primevue/Message'
import AppButton from './AppButton.vue'
import { ArrowPathIcon, DocumentTextIcon } from '@heroicons/vue/24/outline'
import { useReportJobsStore } from '../stores/reportJobs'
import { dateTime } from '../utils/format'
import { api } from '../api'
import type { ReportJobEvent } from '../types'
const jobs = useReportJobsStore()
const emit = defineEmits<{ open: [id: string] }>()
const error = ref(''),
  retrying = ref(''),
  expanded = ref(''),
  events = ref<ReportJobEvent[]>([])
const labels = {
  queued: 'Na fila',
  processing: 'Em preparação',
  completed: 'Pronto',
  failed: 'Falha',
}
async function retry(id: string) {
  retrying.value = id
  error.value = ''
  try {
    await jobs.retry(id)
  } catch (err) {
    error.value = (err as Error).message
  } finally {
    retrying.value = ''
  }
}
async function history(id: string) {
  if (expanded.value === id) {
    expanded.value = ''
    return
  }
  error.value = ''
  try {
    events.value = (await api.reportJob(id)).events || []
    expanded.value = id
  } catch (err) {
    error.value = (err as Error).message
  }
}
</script>
<template>
  <section
    v-if="jobs.jobs.length || jobs.error"
    class="report-queue-panel mt-6"
    aria-label="Solicitações de relatórios"
  >
    <div class="flex flex-wrap items-center justify-between gap-3">
      <h3>Suas solicitações</h3>
      <small v-if="jobs.active">Você pode sair desta tela; o preparo continua.</small>
    </div>
    <Message
      v-if="jobs.error"
      severity="warn"
      :closable="false"
      class="mt-3"
    >
      {{ jobs.error }}
      <AppButton
        variant="link"
        size="small"
        label="Atualizar andamento"
        @click="jobs.refresh"
      />
    </Message>
    <Message
      v-if="error"
      severity="error"
      :closable="false"
      class="mt-3"
    >
      {{ error }}
    </Message>
    <ol class="mt-3 grid gap-3">
      <li
        v-for="job in jobs.jobs.slice(0, 5)"
        :key="job.id"
        class="report-job-row"
        :data-testid="`report-job-${job.id}`"
      >
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div class="min-w-0 flex-1">
            <p class="font-medium break-words">{{ job.request.prompt }}</p>
            <small>
              {{ dateTime(job.created_at) }} ·
              {{
                job.attempts
                  ? `Tentativa ${job.attempts} de ${job.max_attempts}`
                  : 'Solicitação recebida'
              }}
            </small>
          </div>
          <Tag
            :value="labels[job.status]"
            :severity="
              job.status === 'completed'
                ? 'success'
                : job.status === 'failed'
                  ? 'danger'
                  : job.status === 'processing'
                    ? 'info'
                    : 'secondary'
            "
          />
        </div>
        <p
          v-if="job.status === 'queued' || job.status === 'processing'"
          role="status"
          aria-live="polite"
          class="mt-3 text-sm"
        >
          {{ job.progress_message }}
        </p>
        <ProgressBar
          v-if="job.status === 'processing'"
          mode="indeterminate"
          class="mt-3 h-1"
          aria-label="Preparando relatório"
        />
        <p
          v-if="job.status === 'failed'"
          class="mt-3 text-sm"
        >
          {{ job.error_message || job.progress_message }}
        </p>
        <div class="mt-3 flex flex-wrap gap-2">
          <AppButton
            v-if="job.status === 'completed' && job.report_id"
            size="small"
            variant="secondary"
            label="Abrir relatório"
            :icon="DocumentTextIcon"
            @click="emit('open', job.report_id)"
          />
          <AppButton
            v-if="job.status === 'failed'"
            size="small"
            variant="secondary"
            label="Tentar novamente"
            :icon="ArrowPathIcon"
            :loading="retrying === job.id"
            :disabled="!!jobs.active"
            @click="retry(job.id)"
          />
          <AppButton
            variant="link"
            size="small"
            :label="expanded === job.id ? 'Ocultar etapas' : 'Ver etapas'"
            @click="history(job.id)"
          />
        </div>
        <ol
          v-if="expanded === job.id"
          class="report-job-timeline mt-3"
          aria-label="Etapas da solicitação"
        >
          <li
            v-for="event in events"
            :key="event.id"
          >
            <strong>{{ event.agent || 'Solicitação' }}</strong>
            · {{ event.message }}
            <small>{{ dateTime(event.created_at) }}</small>
          </li>
        </ol>
      </li>
    </ol>
  </section>
</template>
<style scoped>
.report-queue-panel {
  border-top: 1px solid var(--p-content-border-color);
  padding-top: 1.25rem;
}
.report-job-row {
  border: 1px solid var(--p-content-border-color);
  border-radius: 16px;
  padding: 1rem;
  background: color-mix(in srgb, var(--p-content-background) 65%, transparent);
}
small {
  color: var(--p-text-muted-color);
}
.report-job-timeline {
  border-left: 2px solid var(--p-primary-color);
  padding-left: 1rem;
  display: grid;
  gap: 0.65rem;
  font-size: 0.85rem;
}
.report-job-timeline small {
  display: block;
}
</style>
