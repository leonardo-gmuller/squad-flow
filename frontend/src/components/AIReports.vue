<script setup lang="ts">
import AppButton from './AppButton.vue'
import Select from './primevue/Select'
import Checkbox from './primevue/Checkbox'
import Message from './primevue/Message'
import { SparklesIcon } from '@heroicons/vue/24/outline'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import Textarea from 'primevue/textarea'
import CalendarDateInput from './CalendarDateInput.vue'
import FormField from './FormField.vue'
import { useToast } from 'primevue/usetoast'
import { api } from '../api'
import { useReportJobsStore } from '../stores/reportJobs'
import ReportJobsPanel from './ReportJobsPanel.vue'
import { useWorkspaceStore } from '../stores/workspace'
import { dateTime, demandCode } from '../utils/format'
import type { GeneratedReport, ReportRequest } from '../types'
import ReportRenderer from './ReportRenderer.vue'

const store = useWorkspaceStore(),
  route = useRoute(),
  toast = useToast(),
  jobs = useReportJobsStore()
const waitingJob = ref(''),
  requestKey = ref('')
const prompt = ref(''),
  demandId = ref(String(route.query.demand_id || '')),
  from = ref(''),
  to = ref('')
const enabled = ref(false),
  configured = ref(false),
  busy = ref(false),
  error = ref(''),
  progress = ref(''),
  history = ref<GeneratedReport[]>([]),
  current = ref<GeneratedReport | null>(null),
  continuePrevious = ref(false)
const invalidPeriod = computed(() => Boolean(from.value && to.value && from.value > to.value))
const completed = computed(() =>
  store.items
    .filter((d) => d.status === 'done')
    .map((d) => ({ label: `${demandCode(d.number)} · ${d.title}`, value: d.id })),
)
let disposed = false
watch([prompt, demandId, from, to], () => {
  if (!busy.value) requestKey.value = ''
})
watch(
  () => jobs.jobs,
  async (rows) => {
    if (!waitingJob.value)
      waitingJob.value =
        rows.find((row) => row.status === 'queued' || row.status === 'processing')?.id || ''
    const job = rows.find((row) => row.id === waitingJob.value)
    if (!job) return
    if (job.status === 'completed' && job.report_id) {
      waitingJob.value = ''
      try {
        current.value = await api.report(job.report_id)
        history.value = await api.reports()
      } catch (err) {
        error.value = (err as Error).message
      }
    } else if (job.status === 'failed') {
      waitingJob.value = ''
      error.value = job.error_message || job.progress_message
    }
  },
)
watch(
  () => route.query.demand_id,
  (value) => {
    demandId.value = String(value || '')
    current.value = null
    continuePrevious.value = false
  },
)
watch([demandId, from, to], () => {
  continuePrevious.value = false
})
onMounted(async () => {
  try {
    const [status, reports] = await Promise.all([api.reportStatus(), api.reports(), jobs.refresh()])
    if (disposed) return
    enabled.value = status.enabled
    history.value = reports
    waitingJob.value = jobs.active?.id || ''
    if (route.query.report_id) await openReport(String(route.query.report_id))
  } catch (err) {
    if (!disposed) error.value = (err as Error).message
  } finally {
    if (!disposed) configured.value = true
  }
})
onUnmounted(() => {
  disposed = true
})
async function openReport(id: string) {
  try {
    error.value = ''
    const result = await api.report(id)
    if (disposed) return
    current.value = result
    continuePrevious.value = false
    history.value = await api.reports()
  } catch (err) {
    if (!disposed) error.value = (err as Error).message
  }
}
async function select(report: GeneratedReport) {
  await openReport(report.id)
}
async function generate() {
  if (invalidPeriod.value) {
    error.value = 'A data final deve ser igual ou posterior à data inicial.'
    return
  }
  busy.value = true
  error.value = ''
  progress.value = 'Registrando solicitação do relatório...'
  const request: ReportRequest = {
    prompt: prompt.value,
    demand_id: demandId.value,
    from: from.value,
    to: to.value,
    previous_report_id: continuePrevious.value ? current.value?.id : undefined,
  }
  try {
    if (!requestKey.value) requestKey.value = crypto.randomUUID()
    const job = await jobs.submit(request, requestKey.value)
    if (disposed) return
    waitingJob.value = job.id
    prompt.value = ''
    continuePrevious.value = false
    requestKey.value = ''
    toast.add({
      severity: 'info',
      summary: 'Solicitação recebida',
      detail: 'Seu relatório será preparado. Você pode continuar usando o sistema.',
      life: 6500,
    })
  } catch (err) {
    if (!disposed) {
      error.value = (err as Error).message
      toast.add({
        severity: 'error',
        summary: 'Não foi possível registrar a solicitação',
        detail: error.value,
        life: 10000,
      })
      try {
        history.value = await api.reports()
      } catch {
        // Preserve the submission error when refreshing the history also fails.
      }
    }
  } finally {
    if (!disposed) {
      busy.value = false
      progress.value = ''
    }
  }
}
</script>
<template>
  <section class="surface-card ai-reports">
    <div class="card-heading">
      <div>
        <span class="overline">ASSISTENTE DE RELATÓRIOS</span>
        <h2>O que você quer analisar?</h2>
        <p>
          Peça uma análise em HTML do workspace ou de uma demanda finalizada. O preparo continua
          mesmo se você sair da tela.
        </p>
      </div>
    </div>
    <Message
      v-if="configured && !enabled"
      severity="info"
      :closable="false"
      role="status"
      class="my-5"
    >
      A geração com IA estará disponível após configurar a chave OpenAI no servidor.
    </Message>
    <form
      class="mt-6 grid gap-5"
      @submit.prevent="generate"
    >
      <div class="grid grid-cols-1 gap-5 sm:grid-cols-2">
        <FormField
          id="report-demand"
          label="Escopo"
          class="col-span-full"
        >
          <Select
            v-model="demandId"
            placeholder="Todas as demandas do workspace"
            aria-label="Escopo"
            input-id="report-demand"
            aria-labelledby="report-demand-label"
            :options="[{ label: 'Todas as demandas do workspace', value: '' }, ...completed]"
            option-label="label"
            option-value="value"
            fluid
            :disabled="busy"
          />
        </FormField>
        <FormField
          id="report-from"
          label="De"
        >
          <CalendarDateInput
            v-model="from"
            input-id="report-from"
            :disabled="busy"
            :invalid="invalidPeriod"
          />
        </FormField>
        <FormField
          id="report-to"
          label="Até"
        >
          <CalendarDateInput
            v-model="to"
            input-id="report-to"
            :min="from"
            :disabled="busy"
            :invalid="invalidPeriod"
          />
        </FormField>
      </div>
      <Message
        v-if="invalidPeriod"
        severity="error"
        size="small"
        variant="simple"
      >
        A data final deve ser igual ou posterior à data inicial.
      </Message>
      <small>
        O período considera a data de conclusão das demandas finalizadas e a data de criação das
        demais (UTC).
      </small>
      <FormField
        id="report-prompt"
        label="Seu pedido"
      >
        <Textarea
          id="report-prompt"
          v-model="prompt"
          rows="3"
          minlength="5"
          maxlength="4000"
          required
          fluid
          :disabled="busy"
          placeholder="Ex.: Analise as devoluções de QA, mostre gargalos por esteira e sugira melhorias com base nas evidências."
        />
      </FormField>
      <div
        v-if="current"
        class="flex items-start gap-3"
      >
        <Checkbox
          v-model="continuePrevious"
          input-id="continue-report"
          binary
          :disabled="busy"
        />
        <label
          for="continue-report"
          class="cursor-pointer text-sm leading-relaxed text-[var(--p-text-muted-color)]"
        >
          Usar o resumo de “{{ current.content.title }}” como contexto da próxima pergunta
        </label>
      </div>
      <div class="flex flex-wrap items-center gap-3">
        <AppButton
          label="Gerar relatório"
          type="submit"
          :loading="busy"
          :disabled="!enabled || !configured || invalidPeriod || !!jobs.active"
          :icon="SparklesIcon"
        />
        <span
          v-if="busy"
          role="status"
          aria-live="polite"
          class="text-sm text-[var(--p-text-muted-color)]"
        >
          {{ progress }}
        </span>
      </div>
      <Message
        v-if="error"
        severity="error"
        :closable="false"
        class="report-generation-error"
      >
        {{ error }}
      </Message>
    </form>
    <FormField
      v-if="history.length"
      id="report-history"
      label="Seus relatórios salvos"
      class="mt-6 border-t border-[var(--p-content-border-color)] pt-5"
    >
      <Select
        aria-label="Seus relatórios salvos"
        input-id="report-history"
        aria-labelledby="report-history-label"
        :model-value="current?.id"
        :options="
          history.map((r) => ({
            label: `${r.content.title} · ${dateTime(r.created_at)}`,
            value: r.id,
          }))
        "
        option-label="label"
        option-value="value"
        placeholder="Abrir um relatório"
        fluid
        :disabled="busy"
        @update:model-value="
          (id) => {
            const report = history.find((r) => r.id === id)
            if (report) select(report)
          }
        "
      />
    </FormField>
    <ReportJobsPanel @open="openReport" />
    <ReportRenderer
      v-if="current"
      :report="current"
    />
  </section>
</template>
