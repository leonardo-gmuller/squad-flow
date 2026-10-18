<script setup lang="ts">
import MemberSelect from './MemberSelect.vue'
import { useCatalogStore } from '../stores/catalog'
import Dialog from './primevue/Dialog'
import AppButton from './AppButton.vue'
import Select from './primevue/Select'
import Checkbox from './primevue/Checkbox'
import Message from './primevue/Message'
import { PlusIcon } from '@heroicons/vue/24/outline'
import { computed, reactive, ref, watch } from 'vue'
import { useToast } from 'primevue/usetoast'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import FormField from './FormField.vue'
import { useRouter } from 'vue-router'

import { useWorkspaceStore } from '../stores/workspace'
import type { CreatePackage } from '../types'
const catalog = useCatalogStore()
const store = useWorkspaceStore(),
  toast = useToast(),
  router = useRouter()
const visible = computed({
  get: () => store.createPackageVisible,
  set: (v) => (store.createPackageVisible = v),
})
const init = (): CreatePackage => ({
  reference: '',
  title: '',
  description: '',
  initiative: '',
  kind: catalog.activeKinds[0]?.value || '',
  create_first_item: true,
  release_id: '',
  assignee_id: '',
  reviewer_id: '',
  product: '',
})
const form = reactive<CreatePackage>(init())
const error = ref('')
watch(
  () => visible.value,
  (v) => {
    if (v) {
      void catalog.load().catch((err) => {
        error.value = (err as Error).message
      })
      Object.assign(form, init())
      error.value = ''
    }
  },
)
async function submit() {
  if (!catalog.activeKinds.some((item) => item.value === form.kind)) {
    error.value =
      'Selecione uma esteira ativa. Se necessário, solicite o cadastro ao administrador.'
    return
  }
  try {
    error.value = ''
    if (form.title.trim().length < 5) {
      error.value = 'Informe um título com ao menos 5 caracteres.'
      return
    }
    const result = await store.createPackage({ ...form, title: form.title.trim() })
    visible.value = false
    toast.add({
      severity: 'success',
      summary: 'OS cadastrada',
      detail: 'A OS está pronta para receber entregas.',
      life: 3000,
    })
    router.push({ path: '/packages', query: { os: result.id } })
  } catch (e) {
    error.value = (e as Error).message
  }
}
</script>
<template>
  <Dialog
    v-model:visible="visible"
    modal
    header="Criar OS / agrupamento"
    class="w-[min(660px,96vw)]"
  >
    <p class="mb-6 text-sm leading-relaxed text-[var(--p-text-muted-color)]">
      A OS agrupa uma ou várias entregas, inclusive de desenvolvedores diferentes. Uma corretiva
      pode ficar avulsa.
    </p>
    <form
      id="package-form"
      class="grid gap-5"
      @submit.prevent="submit"
    >
      <div class="grid grid-cols-1 gap-5 sm:grid-cols-2">
        <FormField
          id="package-reference"
          label="Número / referência"
        >
          <InputText
            id="package-reference"
            v-model="form.reference"
            placeholder="OS-12345"
            fluid
            data-testid="package-reference"
          />
        </FormField>
        <FormField
          id="package-kind"
          label="Esteira"
        >
          <Select
            v-model="form.kind"
            input-id="package-kind"
            aria-labelledby="package-kind-label"
            :options="catalog.activeKinds"
            option-label="label"
            option-value="value"
            fluid
          />
        </FormField>
      </div>
      <FormField
        id="package-title"
        label="Título"
        required
      >
        <InputText
          id="package-title"
          v-model="form.title"
          required
          minlength="5"
          fluid
          placeholder="Ex.: Atendimento à nova Nota Técnica"
          data-testid="package-title"
        />
      </FormField>
      <FormField
        id="package-initiative"
        label="Iniciativa (opcional)"
      >
        <InputText
          id="package-initiative"
          v-model="form.initiative"
          fluid
          placeholder="Ex.: NT 07/2026, evolução contínua..."
        />
      </FormField>
      <FormField
        id="package-description"
        label="Descrição e escopo"
      >
        <Textarea
          id="package-description"
          v-model="form.description"
          rows="4"
          auto-resize
          fluid
          placeholder="Contexto da OS, objetivo e dependências..."
        />
      </FormField>
      <div
        class="flex items-start gap-3 rounded-xl border border-[var(--p-content-border-color)] bg-[var(--p-content-hover-background)] p-4"
      >
        <Checkbox
          v-model="form.create_first_item"
          binary
          input-id="first-item"
          aria-describedby="first-item-help"
        />
        <div>
          <label
            for="first-item"
            class="cursor-pointer text-sm font-medium"
          >
            Criar a primeira entrega automaticamente
          </label>
          <p
            id="first-item-help"
            class="mt-1 text-xs leading-relaxed text-[var(--p-text-muted-color)]"
          >
            Útil para OS executadas inteiramente por uma pessoa; você poderá adicionar mais itens
            depois.
          </p>
        </div>
      </div>
      <div
        v-if="form.create_first_item"
        class="grid grid-cols-1 gap-5 sm:grid-cols-2"
      >
        <FormField
          id="package-assignee"
          label="Desenvolvedor inicial"
        >
          <MemberSelect
            v-model="form.assignee_id"
            input-id="package-assignee"
            labelled-by="package-assignee-label"
          />
        </FormField>
        <FormField
          id="package-reviewer"
          label="QA inicial"
        >
          <MemberSelect
            v-model="form.reviewer_id"
            input-id="package-reviewer"
            labelled-by="package-reviewer-label"
          />
        </FormField>
        <FormField
          id="package-release"
          label="Versão da primeira entrega"
          class="col-span-full"
        >
          <Select
            v-model="form.release_id"
            input-id="package-release"
            aria-labelledby="package-release-label"
            :options="[{ id: '', name: 'Sem versão' }, ...store.activeReleases]"
            option-label="name"
            option-value="id"
            filter
            fluid
          />
        </FormField>
        <FormField
          id="package-product"
          label="Produto"
          class="col-span-full"
        >
          <InputText
            id="package-product"
            v-model="form.product"
            placeholder="Ex.: eSocial Cloud"
            fluid
          />
        </FormField>
      </div>
      <Message
        v-if="store.loaded && !catalog.activeKinds.length"
        severity="warn"
        :closable="false"
      >
        Nenhuma esteira ativa. Solicite o cadastro ao administrador.
      </Message>
      <Message
        v-if="error"
        severity="error"
        :closable="false"
      >
        {{ error }}
      </Message>
    </form>
    <template #footer>
      <AppButton
        variant="ghost"
        label="Cancelar"
        @click="visible = false"
      />
      <AppButton
        type="submit"
        form="package-form"
        label="Criar OS"
        data-testid="save-package"
        :loading="store.busy"
        :icon="PlusIcon"
      />
    </template>
  </Dialog>
</template>
