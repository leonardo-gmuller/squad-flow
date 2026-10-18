<script setup lang="ts">
import AppButton from './AppButton.vue'
import Message from './primevue/Message'
import { KeyIcon } from '@heroicons/vue/24/outline'
import { onUnmounted, ref } from 'vue'
import FormField from './FormField.vue'
import PasswordInput from './PasswordInput.vue'
import { passwordError } from '../utils/password'

withDefaults(
  defineProps<{
    prefix: string
    busy?: boolean
    error?: string
    currentRequired?: boolean
    cancellable?: boolean
    submitLabel?: string
  }>(),
  { submitLabel: 'Salvar nova senha', error: undefined },
)
const emit = defineEmits<{
  submit: [input: { password: string; current_password: string }]
  cancel: []
}>()
const current = ref(''),
  password = ref(''),
  confirmation = ref(''),
  validation = ref('')
function submit() {
  validation.value = passwordError(password.value)
  if (validation.value) return
  if (password.value !== confirmation.value) {
    validation.value = 'A confirmação deve ser igual à nova senha.'
    return
  }
  emit('submit', { password: password.value, current_password: current.value })
}
function clear() {
  current.value = ''
  password.value = ''
  confirmation.value = ''
  validation.value = ''
}
onUnmounted(clear)
defineExpose({ clear })
</script>

<template>
  <form
    class="grid gap-5"
    @submit.prevent="submit"
  >
    <FormField
      v-if="currentRequired"
      :id="`${prefix}-current`"
      label="Senha atual"
      required
    >
      <PasswordInput
        v-model="current"
        :input-id="`${prefix}-current`"
        autocomplete="current-password"
      />
    </FormField>
    <FormField
      :id="`${prefix}-new`"
      label="Nova senha"
      required
      help="Use pelo menos 12 caracteres. Combine letras maiúsculas e minúsculas, números e símbolos."
    >
      <PasswordInput
        v-model="password"
        :input-id="`${prefix}-new`"
        autocomplete="new-password"
        :minlength="12"
        :maxlength="72"
        :described-by="`${prefix}-new-help`"
        feedback
      />
    </FormField>
    <FormField
      :id="`${prefix}-confirmation`"
      label="Confirmar nova senha"
      required
    >
      <PasswordInput
        v-model="confirmation"
        :input-id="`${prefix}-confirmation`"
        autocomplete="new-password"
        :maxlength="72"
      />
    </FormField>
    <Message
      v-if="validation || error"
      severity="error"
      :closable="false"
    >
      {{ validation || error }}
    </Message>
    <div class="flex flex-wrap justify-end gap-2">
      <AppButton
        v-if="cancellable"
        variant="ghost"
        type="button"
        label="Cancelar"
        :disabled="busy"
        @click="emit('cancel')"
      />
      <AppButton
        type="submit"
        :label="submitLabel"
        :loading="busy"
        :icon="KeyIcon"
      />
    </div>
  </form>
</template>
