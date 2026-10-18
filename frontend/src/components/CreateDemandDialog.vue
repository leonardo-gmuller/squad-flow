<script setup lang="ts">
import MemberSelect from './MemberSelect.vue'
import { useCatalogStore } from '../stores/catalog'
import Dialog from './primevue/Dialog'
import AppButton from './AppButton.vue'
import Select from './primevue/Select'
import Message from './primevue/Message'
import { PlusIcon } from '@heroicons/vue/24/outline'
import { computed, reactive, ref, watch } from 'vue'
import { useToast } from 'primevue/usetoast'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import FormField from './FormField.vue'
import CalendarDateInput from './CalendarDateInput.vue'
import { useWorkspaceStore } from '../stores/workspace'
import { priorities } from '../config/catalog'
import type { CreateInput } from '../types'
const catalog = useCatalogStore()
const store = useWorkspaceStore()
const toast = useToast()
const visible = computed({
  get: () => store.createDemandVisible,
  set: (v: boolean) => {
    store.createDemandVisible = v
  },
})
const initial = (): CreateInput => ({
  title: '',
  description: '',
  kind: catalog.activeKinds[0]?.value || '',
  priority: 'medium',
  assignee_id: '',
  reviewer_id: '',
  reference: '',
  event_codes: [],
  package_id: '',
  release_id: '',
  product: '',
  acceptance_criteria: '',
  due_date: null,
})
const form = reactive<CreateInput>(initial())
const codes = ref('')
const due = ref('')
const error = ref('')
watch(
  () => store.createDemandVisible,
  (open) => {
    if (open) {
      void catalog.load().catch((err) => {
        error.value = (err as Error).message
      })
      Object.assign(form, initial())
      codes.value = ''
      due.value = ''
      error.value = ''
      form.package_id = store.defaultPackageId
      form.release_id = store.defaultReleaseId
      packageChanged()
    }
  },
)
function packageChanged() {
  const p = store.packageById(form.package_id)
  if (p) {
    form.kind = p.kind
    form.reference = p.reference
  }
}
async function submit() {
  if (!catalog.activeKinds.some((item) => item.value === form.kind)) {
    error.value =
      'Selecione uma esteira ativa. Se necessário, solicite o cadastro ao administrador.'
    return
  }
  if (form.title.trim().length < 5) {
    error.value = 'O título precisa ter pelo menos 5 caracteres.'
    return
  }
  try {
    error.value = ''
    const result = await store.createDemand({
      ...form,
      title: form.title.trim(),
      event_codes: codes.value
        .split(',')
        .map((c) => c.trim())
        .filter(Boolean),
      due_date: due.value ? `${due.value}T12:00:00Z` : null,
    })
    visible.value = false
    toast.add({
      severity: 'success',
      summary: 'Entrega criada',
      detail: 'A demanda foi adicionada ao fluxo.',
      life: 3500,
    })
    await store.selectDemand(result.id)
  } catch (err) {
    error.value = (err as Error).message
  }
}
</script>
<template>
  <Dialog
    v-model:visible="visible"
    modal
    header="Nova entrega"
    class="w-[min(760px,96vw)]"
    :breakpoints="{ '640px': '96vw' }"
    data-testid="create-demand-dialog"
  >
    <p class="mb-6 text-sm leading-relaxed text-[var(--p-text-muted-color)]">
      Cadastre uma unidade de trabalho. Ela pode pertencer a uma OS ou ser um ticket independente.
    </p>
    <form
      id="demand-form"
      class="grid gap-5"
      @submit.prevent="submit"
    >
      <FormField
        id="new-title"
        label="Título da demanda"
        required
      >
        <InputText
          id="new-title"
          v-model="form.title"
          data-testid="create-title"
          required
          minlength="5"
          maxlength="180"
          placeholder="Ex.: Adequar validação do evento S-2299"
          fluid
        />
      </FormField>
      <FormField
        id="new-description"
        label="Descrição técnica"
      >
        <Textarea
          id="new-description"
          v-model="form.description"
          rows="3"
          auto-resize
          fluid
          placeholder="Objetivo, comportamento esperado e contexto técnico..."
        />
      </FormField>
      <div class="grid grid-cols-1 gap-5 sm:grid-cols-2">
        <FormField
          id="new-kind"
          label="Esteira"
        >
          <Select
            v-model="form.kind"
            input-id="new-kind"
            aria-labelledby="new-kind-label"
            :options="catalog.activeKinds"
            option-label="label"
            option-value="value"
            data-testid="create-kind"
            fluid
          />
        </FormField>
        <FormField
          id="new-priority"
          label="Prioridade"
        >
          <Select
            v-model="form.priority"
            input-id="new-priority"
            aria-labelledby="new-priority-label"
            :options="priorities"
            option-label="label"
            option-value="value"
            fluid
          />
        </FormField>
        <FormField
          id="new-assignee"
          label="Desenvolvedor"
        >
          <MemberSelect
            v-model="form.assignee_id"
            input-id="new-assignee"
            labelled-by="new-assignee-label"
          />
        </FormField>
        <FormField
          id="new-reviewer"
          label="Responsável QA"
        >
          <MemberSelect
            v-model="form.reviewer_id"
            input-id="new-reviewer"
            labelled-by="new-reviewer-label"
          />
        </FormField>
        <FormField
          id="new-package"
          label="OS vinculada"
        >
          <Select
            v-model="form.package_id"
            placeholder="Ticket avulso / sem OS"
            input-id="new-package"
            aria-labelledby="new-package-label"
            :options="[
              { id: '', label: 'Ticket avulso / sem OS' },
              ...store.packages.map((p) => ({
                id: p.id,
                label: `${p.reference || 'OS'} · ${p.title}`,
              })),
            ]"
            option-label="label"
            option-value="id"
            data-testid="create-package"
            filter
            fluid
            @change="packageChanged"
          />
        </FormField>
        <FormField
          id="new-reference"
          label="OS / referência externa"
        >
          <InputText
            id="new-reference"
            v-model="form.reference"
            placeholder="OS-12345 ou ticket externo"
            fluid
          />
        </FormField>
        <FormField
          id="new-release"
          label="Versão de entrega"
        >
          <Select
            v-model="form.release_id"
            input-id="new-release"
            aria-labelledby="new-release-label"
            :options="[{ id: '', name: 'Sem versão' }, ...store.activeReleases]"
            option-label="name"
            option-value="id"
            filter
            fluid
          />
        </FormField>
        <FormField
          id="new-product"
          label="Produto / área"
        >
          <InputText
            id="new-product"
            v-model="form.product"
            placeholder="eSocial, Folha, API..."
            fluid
          />
        </FormField>
        <FormField
          id="new-events"
          label="Eventos impactados"
        >
          <InputText
            id="new-events"
            v-model="codes"
            placeholder="S-2200, S-2205"
            fluid
          />
        </FormField>
        <FormField
          id="due-date"
          label="Data limite"
        >
          <CalendarDateInput
            v-model="due"
            input-id="due-date"
          />
        </FormField>
      </div>
      <FormField
        id="new-acceptance"
        label="Critérios de aceite"
      >
        <Textarea
          id="new-acceptance"
          v-model="form.acceptance_criteria"
          rows="3"
          auto-resize
          fluid
          placeholder="Regras e evidências necessárias para a homologação..."
        />
      </FormField>
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
        form="demand-form"
        label="Criar entrega"
        data-testid="create-submit"
        :loading="store.busy"
        :icon="PlusIcon"
      />
    </template>
  </Dialog>
</template>
