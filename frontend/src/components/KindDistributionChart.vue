<script setup lang="ts">
import { useCatalogStore } from '../stores/catalog'
import { computed } from 'vue'

import type { Metrics } from '../types'
import ReportChart from './ReportChart.vue'

const catalog = useCatalogStore()
const props = defineProps<{ counts: Metrics['by_kind']; title: string }>()
const colors = computed(() => catalog.kindOptions.map((item) => item.color))
const chart = computed(() => ({
  title: props.title,
  kind: 'bar' as const,
  points: catalog.kindOptions.map((kind) => ({
    label: kind.label,
    value: props.counts[kind.value] ?? 0,
  })),
}))
</script>

<template>
  <ReportChart
    :chart="chart"
    horizontal
    :colors="colors"
    :show-title="false"
    :localize-dates="false"
    class="kind-distribution-chart"
  />
</template>
