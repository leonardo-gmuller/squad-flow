<script setup lang="ts">
import { computed, useId } from 'vue'
import Chart from 'primevue/chart'
import type { ChartData, ChartOptions } from 'chart.js'
import { useChartTheme } from '../composables/useChartTheme'
import { chartDateLabel } from '../utils/chartDate'
import { chartFill } from '../utils/chartStyle'

const props = withDefaults(
  defineProps<{
    chart: { title: string; kind: 'bar' | 'line'; points: { label: string; value: number }[] }
    horizontal?: boolean
    colors?: string[]
    showTitle?: boolean
    localizeDates?: boolean
  }>(),
  { horizontal: false, colors: undefined, showTitle: true, localizeDates: true },
)
const theme = useChartTheme()
const tableId = useId()
const label = (value: string, full = false) =>
  props.localizeDates ? chartDateLabel(value, full) : value
const number = (value: number) => value.toLocaleString('pt-BR', { maximumFractionDigits: 2 })
const borderColors = computed(() =>
  props.chart.kind === 'bar' ? (props.colors ?? [theme.value.primary]) : [theme.value.primary],
)
const data = computed<ChartData<'bar' | 'line'>>(() => ({
  labels: props.chart.points.map((point) => label(point.label)),
  datasets: [
    {
      label: props.chart.title,
      data: props.chart.points.map((point) => point.value),
      backgroundColor:
        props.chart.kind === 'bar'
          ? borderColors.value.map((value) => chartFill(value))
          : chartFill(theme.value.primary, 0.14),
      borderColor: props.chart.kind === 'bar' ? borderColors.value : theme.value.primary,
      borderWidth: props.chart.kind === 'line' ? 3 : 2,
      borderSkipped: false,
      hoverBackgroundColor:
        props.chart.kind === 'bar'
          ? borderColors.value.map((value) => chartFill(value, 0.35))
          : chartFill(theme.value.primary, 0.2),
      hoverBorderColor: props.chart.kind === 'bar' ? borderColors.value : theme.value.primary,
      fill: props.chart.kind === 'line' ? 'origin' : false,
      borderRadius: props.chart.kind === 'bar' ? 4 : undefined,
      maxBarThickness: 36,
      pointRadius: 3,
      pointHoverRadius: 5,
      pointBackgroundColor: theme.value.primary,
      pointBorderColor: theme.value.primary,
      tension: 0.2,
    },
  ],
}))
const options = computed<ChartOptions<'bar' | 'line'>>(() => {
  const valueAxis = {
    beginAtZero: true,
    ticks: {
      color: theme.value.text,
      precision: props.chart.points.every((point) => Number.isInteger(point.value)) ? 0 : undefined,
    },
    grid: { color: theme.value.grid },
    border: { display: false },
  }
  const labelAxis = {
    ticks: {
      color: theme.value.text,
      maxRotation: 0,
      autoSkip: true,
      autoSkipPadding: 16,
      maxTicksLimit: props.horizontal ? 12 : 8,
    },
    grid: { display: false },
    border: { display: false },
  }
  return {
    responsive: true,
    maintainAspectRatio: false,
    animation: false,
    locale: 'pt-BR',
    indexAxis: props.horizontal ? 'y' : 'x',
    interaction: { intersect: false, mode: 'index' },
    plugins: {
      legend: { display: false },
      tooltip: {
        callbacks: {
          title: (items) =>
            items.length ? label(props.chart.points[items[0].dataIndex].label, true) : '',
          label: (context) => number(props.chart.points[context.dataIndex].value),
        },
      },
    },
    scales: {
      x: props.horizontal ? valueAxis : labelAxis,
      y: props.horizontal ? labelAxis : valueAxis,
    },
  }
})
</script>

<template>
  <figure class="report-chart">
    <figcaption :class="{ 'sr-only': !showTitle }">{{ chart.title }}</figcaption>
    <div
      v-if="!chart.points.length"
      class="empty-inline"
    >
      Sem dados para este gráfico.
    </div>
    <template v-else>
      <Chart
        :type="chart.kind"
        :data="data"
        :options="options"
        class="report-chart-canvas"
        :canvas-props="{ role: 'img', 'aria-label': chart.title, 'aria-describedby': tableId }"
      />
      <table
        :id="tableId"
        class="sr-only"
      >
        <caption>
          {{
            chart.title
          }}
        </caption>
        <thead>
          <tr>
            <th scope="col">Categoria</th>
            <th scope="col">Valor</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(point, index) in chart.points"
            :key="index"
          >
            <th scope="row">{{ label(point.label, true) }}</th>
            <td>{{ number(point.value) }}</td>
          </tr>
        </tbody>
      </table>
    </template>
  </figure>
</template>
