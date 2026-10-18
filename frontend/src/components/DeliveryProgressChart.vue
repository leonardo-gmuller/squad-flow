<script setup lang="ts">
import { computed, useId } from 'vue'
import Chart from 'primevue/chart'
import type { ChartData, ChartOptions } from 'chart.js'
import { useChartTheme } from '../composables/useChartTheme'
import { percent } from '../utils/format'
import { chartFill } from '../utils/chartStyle'

const props = defineProps<{ delivered: number; active: number; total: number }>()
const theme = useChartTheme()
const descriptionId = useId()
const borderColors = computed(() =>
  props.total ? [theme.value.primary, theme.value.track] : [theme.value.track],
)
const data = computed<ChartData<'doughnut'>>(() => ({
  labels: props.total ? ['Concluídas', 'Em aberto'] : ['Sem entregas'],
  datasets: [
    {
      data: props.total ? [props.delivered, props.active] : [1],
      backgroundColor: borderColors.value.map((value) => chartFill(value, 0.28)),
      borderColor: borderColors.value,
      borderWidth: 2,
      hoverBackgroundColor: borderColors.value.map((value) => chartFill(value, 0.4)),
      hoverBorderColor: borderColors.value,
      hoverOffset: props.total ? 3 : 0,
    },
  ],
}))
const options = computed<ChartOptions<'doughnut'>>(() => ({
  responsive: true,
  maintainAspectRatio: false,
  animation: false,
  locale: 'pt-BR',
  cutout: '80%',
  plugins: { legend: { display: false }, tooltip: { enabled: props.total > 0 } },
}))
</script>

<template>
  <div class="progress-ring">
    <Chart
      type="doughnut"
      :data="data"
      :options="options"
      class="delivery-progress-chart"
      :canvas-props="{
        role: 'img',
        'aria-label': 'Progresso das entregas',
        'aria-describedby': descriptionId,
      }"
    />
    <span
      :id="descriptionId"
      class="sr-only"
    >
      {{ delivered }} concluídas, {{ active }} em aberto, {{ percent(delivered, total) }}% concluído
    </span>
    <div
      class="ring-inner"
      aria-hidden="true"
    >
      <strong>{{ percent(delivered, total) }}%</strong>
      <small>concluído</small>
    </div>
  </div>
</template>
