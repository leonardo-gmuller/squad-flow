<script setup lang="ts">
import AppButton from './AppButton.vue'
import DataTable from './primevue/DataTable'
import { ArrowDownTrayIcon, PresentationChartLineIcon } from '@heroicons/vue/24/outline'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import Message from './primevue/Message'
import { ref } from 'vue'
import { useToast } from 'primevue/usetoast'
import { api } from '../api'
import ReportChart from './ReportChart.vue'
import { useWorkspaceStore } from '../stores/workspace'
import { dateTime, reportCellLabel } from '../utils/format'
import type { GeneratedReport } from '../types'
const props = defineProps<{ report: GeneratedReport }>(),
  store = useWorkspaceStore()
const presenting = ref(false),
  presentationHTML = ref(''),
  presentationReport = ref<GeneratedReport | null>(null),
  loadingHTML = ref(false),
  documentError = ref(''),
  toast = useToast()
async function htmlDocument(id: string) {
  documentError.value = ''
  loadingHTML.value = true
  try {
    return await api.reportHTML(id)
  } catch (err) {
    documentError.value = (err as Error).message
    toast.add({
      severity: 'error',
      summary: 'Não foi possível abrir o documento',
      detail: documentError.value,
      life: 8000,
    })
    return ''
  } finally {
    loadingHTML.value = false
  }
}
async function downloadHTML(id = props.report.id) {
  const html = await htmlDocument(id)
  if (!html) return
  const url = URL.createObjectURL(new Blob([html], { type: 'text/html;charset=utf-8' }))
  const link = document.createElement('a')
  link.href = url
  link.download = `relatorio-${id}.html`
  link.click()
  setTimeout(() => URL.revokeObjectURL(url), 1000)
}
async function present() {
  const selected = props.report
  presentationHTML.value = selected.content.html || (await htmlDocument(selected.id))
  if (presentationHTML.value) {
    presentationReport.value = selected
    presenting.value = true
  }
}

function download() {
  const url = URL.createObjectURL(
    new Blob([JSON.stringify(props.report, null, 2)], { type: 'application/json' }),
  )
  const a = document.createElement('a')
  a.href = url
  a.download = `relatorio-${props.report.id}.json`
  a.click()
  setTimeout(() => URL.revokeObjectURL(url), 1000)
}
</script>
<template>
  <article
    class="generated-report"
    aria-label="Relatório gerado"
  >
    <div class="card-heading">
      <div>
        <span class="overline">RELATÓRIO GERADO · {{ dateTime(report.created_at) }}</span>
        <h2>{{ report.content.title }}</h2>
      </div>
      <div class="flex flex-wrap gap-2">
        <AppButton
          variant="secondary"
          label="Apresentar"
          size="small"
          :loading="loadingHTML"
          :icon="PresentationChartLineIcon"
          @click="present"
        />
        <AppButton
          variant="secondary"
          label="Baixar HTML"
          size="small"
          :loading="loadingHTML"
          :icon="ArrowDownTrayIcon"
          @click="downloadHTML()"
        />
        <AppButton
          variant="ghost"
          label="Baixar JSON"
          size="small"
          :icon="ArrowDownTrayIcon"
          @click="download"
        />
      </div>
    </div>
    <p class="report-summary">{{ report.content.summary }}</p>
    <Message
      v-if="documentError"
      severity="error"
      :closable="false"
    >
      {{ documentError }}
    </Message>
    <iframe
      v-if="report.content.html"
      :srcdoc="report.content.html"
      sandbox=""
      referrerpolicy="no-referrer"
      :title="`Relatório HTML: ${report.content.title}`"
      class="report-html-frame"
      data-testid="report-html"
    />
    <template v-else>
      <div
        v-if="report.content.metrics.length"
        class="report-metrics ai-metrics"
      >
        <div
          v-for="(metric, index) in report.content.metrics"
          :key="index"
          class="report-metric"
        >
          <span>{{ metric.label }}</span>
          <strong>
            {{ metric.value.toLocaleString('pt-BR', { maximumFractionDigits: 2 }) }}
            <small>{{ metric.unit }}</small>
          </strong>
        </div>
      </div>
      <section
        v-for="(section, index) in report.content.sections"
        :key="index"
        class="ai-section"
      >
        <h3>{{ section.title }}</h3>
        <p>{{ section.body }}</p>
      </section>
      <div class="reports-grid">
        <ReportChart
          v-for="(chart, index) in report.content.charts"
          :key="index"
          :chart="chart"
        />
      </div>
      <section
        v-for="(table, index) in report.content.tables"
        :key="index"
        class="ai-section"
      >
        <h3>{{ table.title }}</h3>
        <DataTable
          :value="table.rows"
          size="small"
          scrollable
          striped-rows
          :table-props="{ 'aria-label': table.title }"
        >
          <Column
            v-for="(column, col) in table.columns"
            :key="col"
            :field="String(col)"
            :header="column"
          >
            <template #body="{ data }">{{ reportCellLabel(data[col]) }}</template>
          </Column>
          <template #empty>Nenhum dado neste relatório.</template>
        </DataTable>
      </section>
      <section
        v-if="report.content.recommendations.length"
        class="ai-section"
      >
        <h3>Recomendações</h3>
        <ul>
          <li
            v-for="(item, index) in report.content.recommendations"
            :key="index"
          >
            {{ item }}
          </li>
        </ul>
      </section>
      <section
        v-if="report.content.missing_data.length"
        class="ai-section missing-data"
      >
        <h3>Limitações e dados ausentes</h3>
        <ul>
          <li
            v-for="(item, index) in report.content.missing_data"
            :key="index"
          >
            {{ item }}
          </li>
        </ul>
      </section>
    </template>
    <section
      v-if="report.content.source_ids?.length"
      class="ai-section"
    >
      <h3>Demandas consultadas</h3>
      <div class="source-links">
        <AppButton
          v-for="id in report.content.source_ids"
          :key="id"
          variant="link"
          :label="store.items.find((d) => d.id === id)?.title || id"
          size="small"
          @click="store.selectDemand(id)"
        />
      </div>
    </section>
    <p class="card-footnote">
      Análise gerada por IA. Confira as fontes e as evidências antes de tomar decisões.
    </p>
    <Dialog
      v-model:visible="presenting"
      modal
      maximizable
      :header="presentationReport?.content.title"
      :style="{ width: 'min(96vw,1500px)' }"
      :content-style="{ padding: '0', overflow: 'hidden' }"
    >
      <iframe
        v-if="presenting"
        :srcdoc="presentationHTML"
        sandbox=""
        referrerpolicy="no-referrer"
        :title="`Apresentação: ${presentationReport?.content.title}`"
        class="report-presentation-frame"
      />
      <template #footer>
        <AppButton
          variant="secondary"
          label="Baixar HTML"
          :loading="loadingHTML"
          :icon="ArrowDownTrayIcon"
          @click="downloadHTML(presentationReport?.id)"
        />
        <AppButton
          label="Fechar apresentação"
          @click="presenting = false"
        />
      </template>
    </Dialog>
  </article>
</template>

<style scoped>
.report-html-frame {
  display: block;
  width: 100%;
  height: clamp(480px, 78vh, 1000px);
  border: 1px solid var(--p-content-border-color);
  border-radius: 16px;
  background: #f6f7fb;
  margin: 1.5rem 0;
}
.report-presentation-frame {
  display: block;
  width: 100%;
  height: 76vh;
  border: 0;
  background: #f6f7fb;
}
.generated-report {
  min-width: 0;
}
@media (max-width: 600px) {
  .generated-report > .card-heading {
    flex-direction: column;
    align-items: flex-start;
  }
  .report-html-frame {
    height: 75svh;
    min-height: 450px;
  }
}
</style>
