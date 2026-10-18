<script setup lang="ts">
import AppButton from './AppButton.vue'
import {
  CheckCircleIcon,
  EllipsisHorizontalCircleIcon,
  EyeIcon,
  EyeSlashIcon,
} from '@heroicons/vue/24/outline'
import { computed } from 'vue'
import { tooltip } from '../config/tooltip'
import Password from 'primevue/password'

withDefaults(
  defineProps<{
    inputId: string
    autocomplete: 'current-password' | 'new-password'
    minlength?: number
    maxlength?: number
    describedBy?: string
    feedback?: boolean
  }>(),
  { feedback: false, minlength: undefined, maxlength: undefined, describedBy: undefined },
)
const model = defineModel<string>({ default: '' })
const mediumRegex = /^(?=.{12,}$)(?=.*\p{L})(?=.*[\p{N}\p{P}\p{S}]).*$/u
const strongRegex = /^(?=.{12,}$)(?=.*\p{Ll})(?=.*\p{Lu})(?=.*\p{N})(?=.*[\p{P}\p{S}]).*$/u
const suggestions = computed(() => [
  { label: 'Pelo menos 12 caracteres', met: Array.from(model.value).length >= 12 },
  {
    label: 'Letras maiúsculas e minúsculas',
    met: /\p{Lu}/u.test(model.value) && /\p{Ll}/u.test(model.value),
  },
  { label: 'Números', met: /\p{N}/u.test(model.value) },
  { label: 'Símbolos, como !, @ ou #', met: /[\p{P}\p{S}]/u.test(model.value) },
])
</script>

<template>
  <Password
    v-model="model"
    class="app-password"
    :input-id="inputId"
    :feedback="feedback"
    :medium-regex="mediumRegex"
    :strong-regex="strongRegex"
    toggle-mask
    required
    fluid
    :input-props="{ autocomplete, minlength, maxlength, 'aria-describedby': describedBy }"
  >
    <template #maskicon="{ toggleCallback }">
      <AppButton
        v-tooltip.top="tooltip('Ocultar senha')"
        variant="ghost"
        type="button"
        aria-label="Ocultar senha"
        class="password-toggle"
        size="small"
        :icon="EyeSlashIcon"
        @click="toggleCallback"
      />
    </template>
    <template #unmaskicon="{ toggleCallback }">
      <AppButton
        v-tooltip.top="tooltip('Mostrar senha')"
        variant="ghost"
        type="button"
        aria-label="Mostrar senha"
        class="password-toggle"
        size="small"
        :icon="EyeIcon"
        @click="toggleCallback"
      />
    </template>
    <template
      v-if="feedback"
      #footer
    >
      <div class="password-guidance">
        <p>Sugestões para uma senha forte</p>
        <ul>
          <li
            v-for="suggestion in suggestions"
            :key="suggestion.label"
            :class="{ 'is-met': suggestion.met }"
          >
            <component
              :is="suggestion.met ? CheckCircleIcon : EllipsisHorizontalCircleIcon"
              aria-hidden="true"
              class="app-icon"
            />
            <span>
              {{ suggestion.label }}
              <span class="sr-only">· {{ suggestion.met ? 'Atendido' : 'Pendente' }}</span>
            </span>
          </li>
        </ul>
      </div>
    </template>
  </Password>
</template>
