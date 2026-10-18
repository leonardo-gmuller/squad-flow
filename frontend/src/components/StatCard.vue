<script setup lang="ts">
import type { Component } from 'vue'
defineProps<{
  label: string
  value: string | number
  helper: string
  icon: Component
  tone?: string
}>()
</script>
<template>
  <div class="stat-card">
    <div class="flex items-start justify-between">
      <span class="stat-label">{{ label }}</span>
      <span
        class="stat-symbol"
        :class="tone"
      >
        <component
          :is="icon"
          style="width: 19px; height: 19px; flex-shrink: 0"
        />
      </span>
    </div>
    <div class="stat-number">{{ value }}</div>
    <div class="stat-helper">
      <span
        class="stat-small-dot"
        :class="tone"
      ></span>
      {{ helper }}
    </div>
  </div>
</template>
