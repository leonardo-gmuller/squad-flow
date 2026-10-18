<script setup lang="ts">
import AppButton from '../components/AppButton.vue'
import AppLogo from '../components/AppLogo.vue'
import Message from '../components/primevue/Message'
import { ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import InputText from 'primevue/inputtext'
import PasswordInput from '../components/PasswordInput.vue'
import FormField from '../components/FormField.vue'
import { useAuthStore } from '../stores/auth'
import { Form } from '@primevue/forms'
import type { FormSubmitEvent } from '@primevue/forms'
import { zodResolver } from '@primevue/forms/resolvers/zod'
import { loginSchema, type LoginValues } from '../validation/login'
const auth = useAuthStore(),
  router = useRouter(),
  route = useRoute()
const busy = ref(false),
  error = ref('')
const initialValues: LoginValues = { email: '', password: '' }
const resolver = zodResolver(loginSchema)
function clearError() {
  error.value = ''
  auth.error = ''
}
async function submit(event: FormSubmitEvent) {
  if (busy.value) return
  const values = loginSchema.safeParse(event.values)
  if (!event.valid || !values.success) {
    document.getElementById(event.states.email?.invalid ? 'login-email' : 'login-password')?.focus()
    return
  }
  busy.value = true
  clearError()
  try {
    await auth.login(values.data.email, values.data.password)
    event.reset()
    const target = String(route.query.redirect || '/dashboard')
    await router.replace(target.startsWith('/') && !target.startsWith('//') ? target : '/dashboard')
  } catch (err) {
    error.value = (err as Error).message
  } finally {
    busy.value = false
  }
}
</script>
<template>
  <main class="login-page">
    <section class="surface-card login-card">
      <AppLogo />
      <h1>Entre no workspace<span class="title-period">.</span></h1>
      <p>Demandas, qualidade e relatórios em um só lugar.</p>
      <Message
        v-if="route.query.password === 'changed'"
        severity="success"
        :closable="false"
        class="mb-5"
      >
        Senha alterada. Entre novamente com a nova senha.
      </Message>
      <Form
        v-slot="$form"
        :initial-values="initialValues"
        :resolver="resolver"
        :validate-on-value-update="false"
        :validate-on-blur="true"
        class="grid gap-5"
        novalidate
        @input="clearError"
        @submit="submit"
      >
        <FormField
          id="login-email"
          label="Email"
        >
          <InputText
            id="login-email"
            name="email"
            type="email"
            autocomplete="username"
            :disabled="busy"
            required
            fluid
            :invalid="$form.email?.invalid"
            :aria-invalid="$form.email?.invalid || undefined"
            :aria-describedby="$form.email?.invalid ? 'login-email-error' : undefined"
          />
          <Message
            v-if="$form.email?.invalid"
            id="login-email-error"
            severity="error"
            size="small"
            variant="simple"
            role="alert"
            >{{ $form.email.error?.message }}</Message
          >
        </FormField>
        <FormField
          id="login-password"
          label="Senha"
        >
          <PasswordInput
            name="password"
            input-id="login-password"
            autocomplete="current-password"
            :disabled="busy"
            :invalid="$form.password?.invalid"
            :described-by="$form.password?.invalid ? 'login-password-error' : undefined"
            :aria-invalid="$form.password?.invalid || undefined"
          />
          <Message
            v-if="$form.password?.invalid"
            id="login-password-error"
            severity="error"
            size="small"
            variant="simple"
            role="alert"
            >{{ $form.password.error?.message }}</Message
          >
        </FormField>
        <Message
          v-if="error || auth.error"
          severity="error"
          :closable="false"
          role="alert"
        >
          {{ error || auth.error }}
        </Message>
        <AppButton
          label="Entrar"
          type="submit"
          :loading="busy"
          fluid
        />
      </Form>
      <RouterLink
        to="/reset-password"
        class="password-access-link"
      >
        Esqueci minha senha
      </RouterLink>
      <small>Seu acesso é cadastrado pelo administrador do workspace.</small>
    </section>
  </main>
</template>
