<script setup lang="ts">
import Message from '../components/primevue/Message'
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import PasswordForm from '../components/PasswordForm.vue'
import { useAuthStore } from '../stores/auth'
import { api } from '../api'

const auth = useAuthStore(),
  router = useRouter()
const busy = ref(false),
  error = ref('')
const passwordForm = ref<InstanceType<typeof PasswordForm>>()
async function change(input: { password: string; current_password: string }) {
  busy.value = true
  error.value = ''
  try {
    await api.changePassword(input.current_password, input.password)
    passwordForm.value?.clear()
    auth.expire()
    await router.replace({ path: '/login', query: { password: 'changed' } })
  } catch (err) {
    error.value = (err as Error).message
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div class="page-heading">
    <div>
      <div class="eyebrow">SEU ACESSO</div>
      <h1>Minha conta<span class="title-period">.</span></h1>
      <p>{{ auth.user?.name }} · {{ auth.user?.email }}</p>
    </div>
  </div>
  <section class="surface-card account-password-card">
    <h2>Alterar minha senha</h2>
    <Message
      severity="info"
      :closable="false"
      class="my-5"
    >
      Após a troca, suas sessões serão encerradas. Entre novamente com a nova senha.
    </Message>
    <PasswordForm
      ref="passwordForm"
      prefix="account-password"
      current-required
      :busy="busy"
      :error="error"
      @submit="change"
    />
  </section>
</template>
