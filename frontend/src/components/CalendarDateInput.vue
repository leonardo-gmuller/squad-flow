<script setup lang="ts">
import DatePicker from './primevue/DatePicker'
import { calendarDate, calendarDateString } from '../utils/calendar'

defineProps<{
  inputId: string
  min?: string
  max?: string
  disabled?: boolean
  invalid?: boolean
}>()
const model = defineModel<string>({ default: '' })

function update(value: Date | (Date | null)[] | null | undefined) {
  model.value = value instanceof Date ? calendarDateString(value) : ''
}
</script>

<template>
  <DatePicker
    :input-id="inputId"
    :model-value="calendarDate(model)"
    :min-date="calendarDate(min) ?? undefined"
    :max-date="calendarDate(max) ?? undefined"
    :disabled="disabled"
    :invalid="invalid"
    date-format="dd/mm/yy"
    placeholder="dd/mm/aaaa"
    show-icon
    icon-display="input"
    show-button-bar
    :manual-input="false"
    fluid
    @update:model-value="update"
  />
</template>
