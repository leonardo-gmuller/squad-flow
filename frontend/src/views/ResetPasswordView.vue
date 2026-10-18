<script setup lang="ts">
import Message from '../components/primevue/Message'
import AppLogo from '../components/AppLogo.vue'
import { ArrowRightIcon } from '@heroicons/vue/24/outline'
import { onMounted, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import InputText from 'primevue/inputtext'
import FormField from '../components/FormField.vue'
import PasswordForm from '../components/PasswordForm.vue'
import { api } from '../api'
import { useAuthStore } from '../stores/auth'
import { resetToken } from '../utils/password'

const route = useRoute(),
  router = useRouter(),
  auth = useAuthStore()
const token = ref(new URLSearchParams(route.hash.slice(1)).get('token') || '')
const busy = ref(false),
  error = ref(''),
  done = ref(false)
onMounted(() => {
  if (route.hash) void router.replace({ path: '/reset-password' })
})
async function reset(input: { password: string }) {
  error.value = ''
  const value = resetToken(token.value)
  if (!value) {
    error.value = 'Cole o link ou código de redefinição entregue pelo administrador.'
    return
  }
  busy.value = true
  try {
    const result = await api.resetPassword(value, input.password)
    if (auth.user?.id === result.user_id) auth.expire()
    token.value = ''
    done.value = true
  } catch (err) {
    error.value = (err as Error).message
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <main class="login-page">
    <section class="surface-card login-card password-reset-card">
      <AppLogo />
      <div class="eyebrow">RECUPERAÇÃO DE ACESSO</div>
      <h1>Redefinir senha<span class="title-period">.</span></h1>
      <template v-if="done">
        <Message
          severity="success"
          :closable="false"
          role="status"
        >
          Senha redefinida. Use a nova senha para entrar. As sessões anteriores foram encerradas.
        </Message>
        <RouterLink
          :to="auth.user ? '/dashboard' : '/login'"
          class="password-access-link"
        >
          {{ auth.user ? 'Voltar ao workspace' : 'Ir para o login' }}
          <ArrowRightIcon
            aria-hidden="true"
            class="app-icon"
          />
        </RouterLink>
      </template>
      <template v-else>
        <p>
          Peça um link de recuperação ao administrador. Ele é compartilhado internamente, vale por
          30 minutos e pode ser usado uma vez.
        </p>
        <FormField
          id="reset-token"
          label="Link ou código de redefinição"
          help="Cole o link ou código que o administrador entregou a você."
          class="mb-5"
        >
          <InputText
            id="reset-token"
            v-model="token"
            type="password"
            autocomplete="off"
            :spellcheck="false"
            fluid
          />
        </FormField>
        <PasswordForm
          prefix="reset-password"
          :busy="busy"
          :error="error"
          submit-label="Redefinir minha senha"
          @submit="reset"
        />
        <RouterLink
          to="/login"
          class="password-access-link"
        >
          Voltar para o login
        </RouterLink>
      </template>
    </section>
  </main>
</template>
