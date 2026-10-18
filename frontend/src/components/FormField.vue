<script setup lang="ts">
defineProps<{ id: string; label: string; required?: boolean; help?: string }>()
</script>

<template>
  <div class="flex min-w-0 flex-col gap-2">
    <label
      :id="`${id}-label`"
      :for="id"
      class="text-sm font-medium text-[var(--p-text-color)]"
    >
      {{ label }}
      <span
        v-if="required"
        class="text-red-500"
        aria-hidden="true"
      >
        *
      </span>
    </label>
    <slot />
    <small
      v-if="help"
      :id="`${id}-help`"
      class="text-xs leading-relaxed text-[var(--p-text-muted-color)]"
    >
      {{ help }}
    </small>
  </div>
</template>
