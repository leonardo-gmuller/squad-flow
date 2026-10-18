<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import {
  PlusIcon,
  PencilSquareIcon,
  ArrowPathIcon,
  ArchiveBoxIcon,
} from '@heroicons/vue/24/outline'
import InputText from 'primevue/inputtext'
import ColorPicker from 'primevue/colorpicker'
import Column from 'primevue/column'
import Tag from 'primevue/tag'
import AppButton from '../components/AppButton.vue'
import DataTable from '../components/primevue/DataTable'
import Dialog from '../components/primevue/Dialog'
import Message from '../components/primevue/Message'
import FormField from '../components/FormField.vue'
import { useCatalogStore } from '../stores/catalog'
import { useWorkspaceStore } from '../stores/workspace'
import { tooltip } from '../config/tooltip'
import type { Workstream } from '../types'

const catalog = useCatalogStore(),
  workspace = useWorkspaceStore()
const query = ref(''),
  visible = ref(false),
  editing = ref<Workstream | null>(null),
  busy = ref(false),
  error = ref(''),
  success = ref('')
const pending = ref<Workstream | null>(null)
const form = reactive({ name: '', color: '8976e8' })
const filtered = computed(() =>
  catalog.workstreams.filter((item) =>
    item.name.toLocaleLowerCase('pt-BR').includes(query.value.toLocaleLowerCase('pt-BR')),
  ),
)
function confirmDeactivation(item: Workstream) {
  error.value = ''
  pending.value = item
}
function open(item?: Workstream) {
  editing.value = item || null
  form.name = item?.name || ''
  form.color = item?.color.slice(1) || '8976e8'
  error.value = ''
  visible.value = true
}
async function save() {
  busy.value = true
  error.value = ''
  success.value = ''
  try {
    await catalog.save(
      { name: form.name.trim(), color: `#${form.color}`, active: editing.value?.active ?? true },
      editing.value?.id,
    )
    visible.value = false
    success.value = 'Esteira salva.'
    await workspace.refresh()
  } catch (err) {
    error.value = (err as Error).message
  } finally {
    busy.value = false
  }
}
async function setActive(item: Workstream) {
  busy.value = true
  error.value = ''
  success.value = ''
  try {
    await catalog.save({ name: item.name, color: item.color, active: !item.active }, item.id)
    pending.value = null
    success.value = item.active
      ? 'Esteira desativada. O histórico foi preservado.'
      : 'Esteira reativada.'
    await workspace.refresh()
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
      <h1>Esteiras<span class="title-period">.</span></h1>
      <p>Organize as categorias de trabalho disponíveis nas demandas e nas ordens de serviço.</p>
    </div>
    <div class="heading-actions">
      <AppButton
        label="Nova esteira"
        :icon="PlusIcon"
        @click="open()"
      />
    </div>
  </div>
  <Message
    v-if="success"
    severity="success"
    :closable="false"
    class="mb-5"
    role="status"
  >
    {{ success }}
  </Message>
  <Message
    v-if="error && !visible && !pending"
    severity="error"
    :closable="false"
    class="mb-5"
  >
    {{ error }}
  </Message>
  <section class="surface-card list-card">
    <div class="filter-toolbar">
      <InputText
        v-model="query"
        placeholder="Buscar esteira..."
        aria-label="Buscar esteiras"
        fluid
      />
    </div>
    <DataTable
      :value="filtered"
      data-key="id"
      paginator
      :rows="10"
      scrollable
      :table-props="{ 'aria-label': 'Esteiras cadastradas' }"
    >
      <Column header="Esteira">
        <template #body="{ data }">
          <span
            class="kind-pill"
            :style="catalog.kindStyle(data.code)"
          >
            {{ data.name }}
          </span>
        </template>
      </Column>
      <Column header="Situação">
        <template #body="{ data }">
          <Tag
            :value="data.active ? 'Ativa' : 'Inativa'"
            :severity="data.active ? 'success' : 'secondary'"
          />
        </template>
      </Column>
      <Column header="Ações">
        <template #body="{ data }">
          <div class="flex flex-wrap items-center gap-2">
            <AppButton
              v-tooltip.top="tooltip('Editar esteira')"
              variant="ghost"
              size="small"
              :icon="PencilSquareIcon"
              :aria-label="`Editar esteira ${data.name}`"
              @click="open(data)"
            />
            <AppButton
              v-if="data.active"
              v-tooltip.top="tooltip('Desativar esteira')"
              variant="ghost"
              size="small"
              :icon="ArchiveBoxIcon"
              :aria-label="`Desativar esteira ${data.name}`"
              @click="confirmDeactivation(data)"
            />
            <AppButton
              v-else
              variant="ghost"
              size="small"
              label="Reativar"
              :icon="ArrowPathIcon"
              :aria-label="`Reativar esteira ${data.name}`"
              :loading="busy"
              @click="setActive(data)"
            />
          </div>
        </template>
      </Column>
      <template #empty><div class="empty-inline">Nenhuma esteira encontrada.</div></template>
    </DataTable>
  </section>
  <p class="card-footnote mt-5">
    Desativar uma esteira impede novos vínculos. As demandas, ordens de serviço, métricas e
    históricos continuam disponíveis.
  </p>
  <Dialog
    v-model:visible="visible"
    modal
    :header="editing ? 'Editar esteira' : 'Nova esteira'"
    class="w-[min(480px,96vw)]"
  >
    <form
      id="workstream-form"
      class="grid gap-5"
      @submit.prevent="save"
    >
      <FormField
        id="workstream-name"
        label="Nome da esteira"
        required
      >
        <InputText
          id="workstream-name"
          v-model="form.name"
          required
          minlength="2"
          maxlength="80"
          placeholder="Ex.: Infraestrutura"
          fluid
        />
      </FormField>
      <FormField
        id="workstream-color"
        label="Cor"
      >
        <div class="flex items-center gap-3">
          <ColorPicker
            v-model="form.color"
            input-id="workstream-color"
            format="hex"
            aria-label="Cor da esteira"
          />
          <span
            class="kind-pill"
            :style="{ color: `#${form.color}`, backgroundColor: `#${form.color}20` }"
          >
            {{ form.name || 'Prévia da esteira' }}
          </span>
        </div>
      </FormField>
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
        :disabled="busy"
        @click="visible = false"
      />
      <AppButton
        label="Salvar esteira"
        type="submit"
        form="workstream-form"
        :loading="busy"
      />
    </template>
  </Dialog>
  <Dialog
    :visible="Boolean(pending)"
    modal
    header="Desativar esteira"
    class="w-[min(480px,96vw)]"
    @update:visible="
      (value) => {
        if (!value) pending = null
      }
    "
  >
    <p class="text-sm leading-relaxed">
      A esteira
      <strong>{{ pending?.name }}</strong>
      deixará de aparecer nos novos cadastros. Os vínculos e o histórico existentes serão
      preservados.
    </p>
    <Message
      v-if="error"
      severity="error"
      :closable="false"
      class="mt-4"
    >
      {{ error }}
    </Message>
    <template #footer>
      <AppButton
        variant="ghost"
        label="Cancelar"
        :disabled="busy"
        @click="pending = null"
      />
      <AppButton
        label="Desativar esteira"
        :loading="busy"
        @click="pending && setActive(pending)"
      />
    </template>
  </Dialog>
</template>
