<script setup lang="ts">
import { computed } from 'vue'
import Select from './primevue/Select'
import { useCatalogStore } from '../stores/catalog'

const props = defineProps<{ inputId: string; labelledBy: string; legacyName?: string }>()
const model = defineModel<string>({ default: '' })
const catalog = useCatalogStore()
const options = computed(() =>
  props.legacyName
    ? [
        { id: 'legacy', label: `${props.legacyName} · Sem vínculo`, disabled: true },
        ...catalog.memberOptions,
      ]
    : catalog.memberOptions,
)
</script>

<template>
  <Select
    v-model="model"
    :input-id="inputId"
    :aria-labelledby="labelledBy"
    :options="options"
    option-disabled="disabled"
    option-label="label"
    option-value="id"
    filter
    placeholder="Selecione um usuário"
    empty-filter-message="Nenhum usuário encontrado"
    fluid
  />
</template>
