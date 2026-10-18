<script setup lang="ts">
import { computed, useSlots, type Component } from 'vue'
import PrimeButton from 'primevue/button'
import { ArrowPathIcon } from '@heroicons/vue/24/outline'

defineOptions({ inheritAttrs: false })

const props = withDefaults(
  defineProps<{
    label?: string
    icon?: Component
    variant?: 'primary' | 'secondary' | 'ghost' | 'danger' | 'link'
    size?: 'small' | 'normal' | 'large'
    type?: 'button' | 'submit' | 'reset'
    loading?: boolean
    disabled?: boolean
    fluid?: boolean
  }>(),
  {
    label: undefined,
    icon: undefined,
    variant: 'primary',
    size: 'normal',
    type: 'button',
    loading: false,
    disabled: false,
    fluid: false,
  },
)

const slots = useSlots()
const iconOnly = computed(() => !props.label && !slots.default && Boolean(props.icon || slots.icon))
const severity = computed(() =>
  props.variant === 'danger'
    ? 'danger'
    : props.variant === 'secondary' || props.variant === 'ghost'
      ? 'secondary'
      : undefined,
)
const primeVariant = computed(() =>
  props.variant === 'secondary'
    ? 'outlined'
    : props.variant === 'ghost'
      ? 'text'
      : props.variant === 'link'
        ? 'link'
        : undefined,
)
</script>

<template>
  <PrimeButton
    v-bind="$attrs"
    :label="label"
    :type="type"
    :loading="loading"
    :disabled="disabled || loading"
    :fluid="fluid"
    :severity="severity"
    :variant="primeVariant"
    :size="size === 'normal' ? undefined : size"
    :aria-busy="loading || undefined"
    class="app-button"
    :class="[
      `app-button--${variant}`,
      `app-button--${size}`,
      { 'app-button--icon-only': iconOnly, 'app-button--fluid': fluid },
    ]"
  >
    <template
      v-if="slots.default"
      #default
    >
      <slot />
    </template>
    <template
      v-if="icon || slots.icon"
      #icon="scope"
    >
      <slot
        name="icon"
        v-bind="scope"
      >
        <component
          :is="icon"
          :class="scope.class"
          class="app-icon"
          aria-hidden="true"
        />
      </slot>
    </template>
    <template #loadingicon="scope">
      <ArrowPathIcon
        :class="scope.class"
        class="app-icon app-icon-spin"
        aria-hidden="true"
      />
    </template>
  </PrimeButton>
</template>
