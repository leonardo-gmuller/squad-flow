<script setup lang="ts">
import AppButton from '../components/AppButton.vue'
import Message from '../components/primevue/Message'
import DataTable from '../components/primevue/DataTable'
import Dialog from '../components/primevue/Dialog'
import {
  ArrowDownIcon,
  ArrowUpIcon,
  ArrowsUpDownIcon,
  CheckIcon,
  ClipboardDocumentIcon,
  KeyIcon,
  LinkIcon,
} from '@heroicons/vue/24/outline'
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import InputText from 'primevue/inputtext'
import PasswordInput from '../components/PasswordInput.vue'
import Column from 'primevue/column'
import Tag from 'primevue/tag'
import FormField from '../components/FormField.vue'
import PasswordForm from '../components/PasswordForm.vue'
import { useAuthStore } from '../stores/auth'
import { tooltip } from '../config/tooltip'
import { dateTime } from '../utils/format'
import { passwordError } from '../utils/password'
import { api } from '../api'
import { useCatalogStore } from '../stores/catalog'
import type { User } from '../types'
const users = ref<User[]>([]),
  busy = ref(false),
  error = ref(''),
  success = ref('')
const query = ref(''),
  first = ref(0)
const filteredUsers = computed(() =>
  users.value.filter((user) =>
    `${user.name} ${user.email}`
      .toLocaleLowerCase('pt-BR')
      .includes(query.value.toLocaleLowerCase('pt-BR')),
  ),
)
watch(
  query,
  () => {
    first.value = 0
  },
  { flush: 'sync' },
)
const form = reactive({ name: '', email: '', password: '' })
const auth = useAuthStore(),
  router = useRouter()
const selectedUser = ref<User | null>(null),
  passwordVisible = ref(false),
  resetVisible = ref(false),
  actionBusy = ref(false),
  actionError = ref(''),
  copied = ref(false)
const reset = ref<{ token: string; expires_at: string } | null>(null)
const passwordForm = ref<InstanceType<typeof PasswordForm>>()
const resetLink = computed(() =>
  reset.value
    ? `${window.location.origin}/reset-password#token=${encodeURIComponent(reset.value.token)}`
    : '',
)
watch(resetVisible, (visible) => {
  if (!visible) {
    reset.value = null
    copied.value = false
    selectedUser.value = null
    actionError.value = ''
  }
})
watch(passwordVisible, (visible) => {
  if (!visible) {
    passwordForm.value?.clear()
    selectedUser.value = null
    actionError.value = ''
  }
})
function openPassword(user: User) {
  selectedUser.value = user
  actionError.value = ''
  passwordVisible.value = true
}
function openReset(user: User) {
  selectedUser.value = user
  reset.value = null
  copied.value = false
  actionError.value = ''
  resetVisible.value = true
}
async function changePassword(input: { password: string }) {
  if (!selectedUser.value) return
  actionBusy.value = true
  actionError.value = ''
  try {
    const own = selectedUser.value.id === auth.user?.id
    await api.setUserPassword(selectedUser.value.id, input.password)
    passwordForm.value?.clear()
    passwordVisible.value = false
    if (own) {
      auth.expire()
      await router.replace({ path: '/login', query: { password: 'changed' } })
    } else success.value = 'Senha alterada. As sessões anteriores do usuário foram encerradas.'
  } catch (err) {
    actionError.value = (err as Error).message
  } finally {
    actionBusy.value = false
  }
}
async function generateReset() {
  if (!selectedUser.value) return
  actionBusy.value = true
  actionError.value = ''
  try {
    reset.value = await api.createPasswordReset(selectedUser.value.id)
    copied.value = false
  } catch (err) {
    actionError.value = (err as Error).message
  } finally {
    actionBusy.value = false
  }
}
async function copyReset() {
  try {
    await navigator.clipboard.writeText(resetLink.value)
    copied.value = true
  } catch {
    actionError.value =
      'Não foi possível copiar automaticamente. Selecione o link e copie manualmente.'
  }
}
onMounted(async () => {
  try {
    users.value = await api.users()
  } catch (err) {
    error.value = (err as Error).message
  }
})
async function create() {
  error.value = ''
  success.value = ''
  error.value = passwordError(form.password)
  if (error.value) return
  busy.value = true
  try {
    const user = await api.createUser(form)
    users.value.push(user)
    query.value = ''
    first.value = Math.floor((users.value.length - 1) / 10) * 10
    Object.assign(form, { name: '', email: '', password: '' })
    success.value = 'Usuário cadastrado.'
    await useCatalogStore().load()
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
      <div class="eyebrow">ADMINISTRAÇÃO</div>
      <h1>Usuários<span class="title-period">.</span></h1>
      <p>Cadastre o acesso das pessoas ao workspace compartilhado.</p>
    </div>
  </div>
  <div class="grid min-w-0 grid-cols-1 gap-5 xl:grid-cols-[minmax(280px,1fr)_minmax(0,2fr)]">
    <section class="surface-card p-6">
      <h2 class="text-lg font-semibold">Novo usuário</h2>
      <form
        class="mt-5 grid gap-5"
        @submit.prevent="create"
      >
        <FormField
          id="user-name"
          label="Nome"
        >
          <InputText
            id="user-name"
            v-model="form.name"
            minlength="2"
            maxlength="120"
            autocomplete="name"
            required
            fluid
          />
        </FormField>
        <FormField
          id="user-email"
          label="Email"
        >
          <InputText
            id="user-email"
            v-model="form.email"
            type="email"
            autocomplete="email"
            required
            fluid
          />
        </FormField>
        <FormField
          id="user-password"
          label="Senha inicial"
          help="Use pelo menos 12 caracteres. Combine letras maiúsculas e minúsculas, números e símbolos para uma senha mais forte."
        >
          <PasswordInput
            v-model="form.password"
            input-id="user-password"
            autocomplete="new-password"
            :minlength="12"
            :maxlength="72"
            described-by="user-password-help"
            feedback
          />
        </FormField>
        <Message
          v-if="error"
          severity="error"
          :closable="false"
        >
          {{ error }}
        </Message>
        <Message
          v-if="success"
          severity="success"
          :closable="false"
          role="status"
        >
          {{ success }}
        </Message>
        <AppButton
          type="submit"
          label="Cadastrar usuário"
          :loading="busy"
        />
      </form>
    </section>
    <section class="surface-card min-w-0 overflow-hidden p-6">
      <h2 class="mb-5 text-lg font-semibold">Acessos cadastrados</h2>
      <InputText
        v-model="query"
        aria-label="Buscar usuários"
        placeholder="Buscar por nome ou e-mail..."
        class="mb-4"
        fluid
      />
      <DataTable
        v-model:first="first"
        :value="filteredUsers"
        data-key="id"
        size="small"
        scrollable
        paginator
        :rows="10"
        :rows-per-page-options="[10, 20, 50]"
      >
        <Column
          field="name"
          header="Nome"
          sortable
        >
          <template #sorticon="{ sortOrder, class: iconClass }">
            <component
              :is="
                sortOrder === 1 ? ArrowUpIcon : sortOrder === -1 ? ArrowDownIcon : ArrowsUpDownIcon
              "
              :class="iconClass"
              class="app-icon"
            />
          </template>
        </Column>
        <Column
          field="email"
          header="Email"
          sortable
        >
          <template #sorticon="{ sortOrder, class: iconClass }">
            <component
              :is="
                sortOrder === 1 ? ArrowUpIcon : sortOrder === -1 ? ArrowDownIcon : ArrowsUpDownIcon
              "
              :class="iconClass"
              class="app-icon"
            />
          </template>
        </Column>
        <Column
          field="role"
          header="Perfil"
        >
          <template #body="{ data }">
            <Tag
              :value="data.role === 'admin' ? 'Administrador' : 'Usuário'"
              :severity="data.role === 'admin' ? 'primary' : 'secondary'"
            />
          </template>
        </Column>
        <Column header="Acesso">
          <template #body="{ data }">
            <div class="flex flex-wrap gap-1">
              <AppButton
                variant="ghost"
                label="Alterar senha"
                size="small"
                :aria-label="`Alterar senha de ${data.name}`"
                :icon="KeyIcon"
                @click="openPassword(data)"
              />
              <AppButton
                v-tooltip.top="tooltip('Gerar link de redefinição')"
                variant="ghost"
                size="small"
                :aria-label="`Gerar link de redefinição para ${data.name}`"
                :icon="LinkIcon"
                @click="openReset(data)"
              />
            </div>
          </template>
        </Column>
        <template #empty>Nenhum usuário cadastrado.</template>
      </DataTable>
    </section>
  </div>
  <Dialog
    v-model:visible="passwordVisible"
    modal
    header="Alterar senha do usuário"
    :style="{ width: 'min(480px,calc(100vw - 24px))' }"
    :closable="!actionBusy"
    :close-on-escape="!actionBusy"
  >
    <p class="mb-4 text-sm">{{ selectedUser?.name }} · {{ selectedUser?.email }}</p>
    <Message
      severity="info"
      :closable="false"
      class="mb-5"
    >
      A nova senha encerrará as sessões anteriores e invalidará links de redefinição pendentes.
    </Message>
    <PasswordForm
      ref="passwordForm"
      prefix="admin-password"
      :busy="actionBusy"
      :error="actionError"
      cancellable
      @cancel="passwordVisible = false"
      @submit="changePassword"
    />
  </Dialog>
  <Dialog
    v-model:visible="resetVisible"
    modal
    header="Redefinição interna de senha"
    :style="{ width: 'min(560px,calc(100vw - 24px))' }"
    :closable="!actionBusy"
    :close-on-escape="!actionBusy"
  >
    <p class="mb-4 text-sm">{{ selectedUser?.name }} · {{ selectedUser?.email }}</p>
    <template v-if="reset">
      <Message
        severity="success"
        :closable="false"
        class="mb-4"
      >
        Link gerado. Compartilhe-o diretamente com esta pessoa pelo canal interno da equipe.
      </Message>
      <FormField
        id="password-reset-link"
        label="Link de redefinição"
        :help="`Válido até ${dateTime(reset.expires_at)}. Pode ser usado apenas uma vez.`"
      >
        <InputText
          id="password-reset-link"
          :model-value="resetLink"
          readonly
          fluid
          @focus="($event.target as HTMLInputElement).select()"
        />
      </FormField>
      <p class="mt-3 text-xs leading-relaxed text-[var(--p-text-muted-color)]">
        O link fica disponível nesta janela. Gerar outro link ou alterar a senha invalida os
        anteriores. As sessões serão encerradas quando a pessoa definir a nova senha.
      </p>
      <AppButton
        :label="copied ? 'Link copiado' : 'Copiar link'"
        class="mt-4"
        :icon="copied ? CheckIcon : ClipboardDocumentIcon"
        @click="copyReset"
      />
    </template>
    <template v-else>
      <Message
        severity="info"
        :closable="false"
      >
        O usuário escolherá a própria senha por um link válido por 30 minutos. Links anteriores
        serão invalidados. Nenhum e-mail será enviado.
      </Message>
      <AppButton
        label="Gerar link"
        class="mt-5"
        :loading="actionBusy"
        :icon="LinkIcon"
        @click="generateReset"
      />
    </template>
    <Message
      v-if="actionError"
      severity="error"
      :closable="false"
      class="mt-4"
    >
      {{ actionError }}
    </Message>
  </Dialog>
</template>
