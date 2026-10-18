<script setup lang="ts">
import MemberSelect from './MemberSelect.vue'
import { useCatalogStore } from '../stores/catalog'
import Drawer from './primevue/Drawer'
import AppButton from './AppButton.vue'
import Select from './primevue/Select'
import Message from './primevue/Message'
import {
  ArrowRightIcon,
  ArrowUturnLeftIcon,
  CheckCircleIcon,
  CheckIcon,
  ClipboardDocumentCheckIcon,
  ClockIcon,
  DocumentTextIcon,
  PencilSquareIcon,
  PlusIcon,
  SparklesIcon,
  UserCircleIcon,
  XCircleIcon,
} from '@heroicons/vue/24/outline'
import { computed, reactive, ref, watch } from 'vue'
import { useToast } from 'primevue/usetoast'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import CalendarDateInput from './CalendarDateInput.vue'
import FormField from './FormField.vue'
import { useWorkspaceStore } from '../stores/workspace'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'
import { priorities, statusLabel, priorityLabel, transitions } from '../config/catalog'
import { dateLabel, dateTime, demandCode, initials } from '../utils/format'
import type { Priority, Status } from '../types'

const catalog = useCatalogStore()
const store = useWorkspaceStore(),
  toast = useToast()
const auth = useAuthStore(),
  router = useRouter()
const visible = computed({
  get: () => Boolean(store.selectedId),
  set: (v) => {
    if (!v) store.closeDemand()
  },
})
const demand = computed(() => store.detail?.demand)
const activeTests = computed(
  () => store.detail?.tests.filter((t) => t.cycle === demand.value?.qa_cycle) || [],
)
const editing = ref(false),
  showTest = ref(false),
  actor = computed(() => auth.user?.name || ''),
  moveReason = ref(''),
  error = ref('')
function report() {
  if (!demand.value) return
  const id = demand.value.id
  store.closeDemand()
  void router.push({ path: '/reports', query: { demand_id: id } })
}
const editDue = ref('')
const releaseOptions = computed(() => [
  { id: '', name: 'Sem versão' },
  ...store.releases
    .filter((r) => !r.archived_at || r.id === demand.value?.release_id)
    .map((r) => ({ id: r.id, name: r.name + (r.archived_at ? ' · Arquivada' : '') })),
])
const editForm = reactive({
  release_id: '',
  title: '',
  description: '',
  priority: 'medium' as Priority,
  assignee_id: '',
  reviewer_id: '',
  product: '',
  reference: '',
  acceptance_criteria: '',
  event_codes: '',
})
const testForm = reactive({
  title: '',
  steps: '',
  expected: '',
  actual: '',
  result: 'passed' as 'passed' | 'failed',
  author: '',
})
watch(
  () => store.selectedId,
  () => {
    editing.value = false
    showTest.value = false
    error.value = ''
    moveReason.value = ''
    testForm.title = ''
    testForm.steps = ''
    testForm.expected = ''
    testForm.actual = ''
    testForm.result = 'passed'
  },
)
function toggleEdit() {
  if (editing.value) {
    editing.value = false
    return
  }
  startEdit()
}
function startEdit() {
  if (!demand.value) return
  Object.assign(editForm, {
    release_id: demand.value.release_id || '',
    title: demand.value.title,
    description: demand.value.description,
    priority: demand.value.priority,
    assignee_id: demand.value.assignee_id || (demand.value.assignee ? 'legacy' : ''),
    reviewer_id: demand.value.reviewer_id || (demand.value.reviewer ? 'legacy' : ''),
    product: demand.value.product,
    reference: demand.value.reference,
    acceptance_criteria: demand.value.acceptance_criteria,
    event_codes: demand.value.event_codes.join(', '),
  })
  editDue.value = demand.value.due_date?.slice(0, 10) || ''
  editing.value = true
}
async function saveEdit() {
  if (!demand.value) return
  if (editForm.assignee_id === 'legacy' || editForm.reviewer_id === 'legacy') {
    error.value = 'Vincule os responsáveis antigos a usuários cadastrados ou escolha Não atribuído.'
    return
  }
  try {
    error.value = ''
    await store.editDemand(demand.value.id, {
      ...editForm,
      event_codes: editForm.event_codes
        .split(',')
        .map((c) => c.trim())
        .filter(Boolean),
      due_date: editDue.value ? `${editDue.value}T12:00:00Z` : null,
    })
    editing.value = false
    toast.add({ severity: 'success', summary: 'Alterações salvas', life: 2500 })
  } catch (e) {
    error.value = (e as Error).message
  }
}
async function move(next: Status) {
  if (!demand.value) return
  if (next === 'changes_requested' && !moveReason.value.trim()) {
    error.value = 'Informe o motivo da devolução.'
    return
  }
  try {
    error.value = ''
    await store.moveDemand(demand.value.id, next, actor.value.trim(), moveReason.value)
    moveReason.value = ''
    toast.add({
      severity: 'success',
      summary: next === 'done' ? 'Entrega concluída' : 'Status atualizado',
      detail: statusLabel(next),
      life: 2500,
    })
  } catch (e) {
    error.value = (e as Error).message
  }
}
async function saveTest() {
  if (!demand.value) return
  try {
    error.value = ''
    if (testForm.result === 'failed' && !testForm.actual.trim()) {
      error.value = 'Descreva o resultado observado ao reprovar.'
      return
    }
    await store.addTest(demand.value.id, { ...testForm, author: testForm.author || actor.value })
    showTest.value = false
    testForm.title = ''
    testForm.steps = ''
    testForm.expected = ''
    testForm.actual = ''
    testForm.result = 'passed'
    toast.add({ severity: 'success', summary: 'Cenário registrado', life: 2700 })
  } catch (e) {
    error.value = (e as Error).message
  }
}
const moveText: Record<Status, string> = {
  backlog: 'Retornar ao backlog',
  in_progress: 'Iniciar desenvolvimento',
  in_qa: 'Enviar para QA',
  changes_requested: 'Devolver para correção',
  done: 'Aprovar e concluir',
}
</script>
<template>
  <Drawer
    v-model:visible="visible"
    position="right"
    :style="{ width: 'min(680px,100vw)' }"
    class="task-drawer"
    :pt="{ content: { class: '!p-0' } }"
    :show-close-icon="true"
    :close-button-props="{ 'aria-label': 'Fechar detalhes' }"
    header="Detalhes da demanda"
    role="dialog"
    aria-label="Detalhes da demanda"
  >
    <div
      v-if="!store.detail"
      class="py-20 text-center text-sm text-slate-500"
    >
      Carregando detalhes da demanda...
    </div>
    <div
      v-else-if="demand"
      class="drawer-inner"
    >
      <div class="drawer-topline">
        <span class="ticket-key">{{ demandCode(demand.number) }}</span>
        <span class="flex items-center gap-2">
          <span
            class="status-pill"
            :class="demand.status"
          >
            <span></span>
            {{ statusLabel(demand.status) }}
          </span>
        </span>
      </div>
      <h2 class="drawer-title">{{ demand.title }}</h2>
      <div class="drawer-badges">
        <span
          class="kind-pill"
          :style="catalog.kindStyle(demand.kind)"
        >
          {{ catalog.kindLabel(demand.kind) }}
        </span>
        <span
          class="priority-badge"
          :class="demand.priority"
        >
          <span></span>
          {{ priorityLabel(demand.priority) }}
        </span>
        <span
          v-if="demand.reference"
          class="light-pill"
        >
          {{ demand.reference }}
        </span>
      </div>
      <div class="drawer-actions-top">
        <AppButton
          variant="secondary"
          label="Editar entrega"
          size="small"
          :icon="PencilSquareIcon"
          @click="toggleEdit"
        />
        <span class="text-xs text-slate-400">Atualizada {{ dateLabel(demand.updated_at) }}</span>
      </div>
      <Message
        v-if="error"
        severity="error"
        :closable="false"
        class="mt-4"
      >
        {{ error }}
      </Message>
      <form
        v-if="editing"
        class="edit-block grid gap-5"
        @submit.prevent="saveEdit"
      >
        <Message
          v-if="
            (demand.assignee && !demand.assignee_id) || (demand.reviewer && !demand.reviewer_id)
          "
          severity="warn"
          :closable="false"
        >
          Há responsáveis antigos sem vínculo com uma conta. Selecione os usuários cadastrados para
          atualizar a atribuição.
        </Message>
        <h3 class="block-title">Editar informações</h3>
        <FormField
          id="edit-title"
          label="Título"
        >
          <InputText
            id="edit-title"
            v-model="editForm.title"
            required
            minlength="5"
            fluid
          />
        </FormField>
        <div class="grid grid-cols-1 gap-5 sm:grid-cols-2">
          <FormField
            id="edit-priority"
            label="Prioridade"
          >
            <Select
              v-model="editForm.priority"
              input-id="edit-priority"
              aria-labelledby="edit-priority-label"
              :options="priorities"
              option-label="label"
              option-value="value"
              fluid
            />
          </FormField>
          <FormField
            id="edit-due"
            label="Prazo"
          >
            <CalendarDateInput
              v-model="editDue"
              input-id="edit-due"
            />
          </FormField>
          <FormField
            id="edit-assignee"
            label="Desenvolvedor"
          >
            <MemberSelect
              v-model="editForm.assignee_id"
              input-id="edit-assignee"
              labelled-by="edit-assignee-label"
              :legacy-name="!demand.assignee_id ? demand.assignee : undefined"
            />
          </FormField>
          <FormField
            id="edit-reviewer"
            label="QA"
          >
            <MemberSelect
              v-model="editForm.reviewer_id"
              input-id="edit-reviewer"
              labelled-by="edit-reviewer-label"
              :legacy-name="!demand.reviewer_id ? demand.reviewer : undefined"
            />
          </FormField>
          <FormField
            id="edit-release"
            label="Versão de entrega"
          >
            <Select
              v-model="editForm.release_id"
              input-id="edit-release"
              aria-labelledby="edit-release-label"
              :options="releaseOptions"
              option-label="name"
              option-value="id"
              filter
              fluid
            />
          </FormField>
          <FormField
            id="edit-product"
            label="Produto"
          >
            <InputText
              id="edit-product"
              v-model="editForm.product"
              fluid
            />
          </FormField>
          <FormField
            id="edit-reference"
            label="Referência externa"
          >
            <InputText
              id="edit-reference"
              v-model="editForm.reference"
              fluid
            />
          </FormField>
          <FormField
            id="edit-events"
            label="Eventos (separados por vírgula)"
            class="col-span-full"
          >
            <InputText
              id="edit-events"
              v-model="editForm.event_codes"
              fluid
            />
          </FormField>
        </div>
        <FormField
          id="edit-description"
          label="Descrição"
        >
          <Textarea
            id="edit-description"
            v-model="editForm.description"
            rows="3"
            auto-resize
            fluid
          />
        </FormField>
        <FormField
          id="edit-acceptance"
          label="Critérios de aceite"
        >
          <Textarea
            id="edit-acceptance"
            v-model="editForm.acceptance_criteria"
            rows="3"
            auto-resize
            fluid
          />
        </FormField>
        <div class="flex justify-end gap-2">
          <AppButton
            variant="ghost"
            label="Cancelar"
            @click="editing = false"
          />
          <AppButton
            label="Salvar mudanças"
            type="submit"
            :loading="store.busy"
          />
        </div>
      </form>
      <section class="drawer-section">
        <h3 class="block-title">
          <DocumentTextIcon style="width: 17px; height: 17px; flex-shrink: 0" />
          Contexto da entrega
        </h3>
        <p class="rich-description">
          {{ demand.description || 'Nenhuma descrição técnica cadastrada.' }}
        </p>
        <div
          v-if="demand.acceptance_criteria"
          class="acceptance-box"
        >
          <strong>Critérios de aceite</strong>
          <p>{{ demand.acceptance_criteria }}</p>
        </div>
      </section>
      <section class="drawer-section">
        <h3 class="block-title">
          <UserCircleIcon style="width: 17px; height: 17px; flex-shrink: 0" />
          Responsáveis e informações
        </h3>
        <div class="metadata-grid">
          <div>
            <span>Desenvolvedor</span>
            <strong>
              <span class="avatar-mini">{{ initials(demand.assignee) }}</span>
              {{ demand.assignee || 'Não atribuído' }}
            </strong>
          </div>
          <div>
            <span>Responsável QA</span>
            <strong>
              <span class="avatar-mini qa-avatar">{{ initials(demand.reviewer) }}</span>
              {{ demand.reviewer || 'Não atribuído' }}
            </strong>
          </div>
          <div>
            <span>OS agrupadora</span>
            <strong>
              {{ store.packageById(demand.package_id)?.reference || 'Ticket avulso / sem OS' }}
            </strong>
          </div>
          <div>
            <span>Versão de entrega</span>
            <strong>
              {{
                store.releaseById(demand.release_id)?.name || demand.release_name || 'Sem versão'
              }}
              <small v-if="store.releaseById(demand.release_id)?.archived_at">· Arquivada</small>
            </strong>
          </div>
          <div>
            <span>Produto</span>
            <strong>{{ demand.product || '—' }}</strong>
          </div>
          <div>
            <span>Prazo</span>
            <strong>{{ dateLabel(demand.due_date) }}</strong>
          </div>
          <div>
            <span>Devoluções</span>
            <strong>{{ demand.return_count }}</strong>
          </div>
        </div>
        <div
          v-if="demand.event_codes?.length"
          class="mt-4 flex flex-wrap gap-2"
        >
          <span
            v-for="code in demand.event_codes"
            :key="code"
            class="light-pill"
          >
            {{ code }}
          </span>
        </div>
      </section>
      <section class="drawer-section">
        <h3 class="block-title">
          <ArrowUturnLeftIcon style="width: 17px; height: 17px; flex-shrink: 0" />
          Movimentação
        </h3>
        <div
          v-if="demand.status === 'done'"
          class="complete-banner"
        >
          <CheckCircleIcon style="width: 18px; height: 18px; flex-shrink: 0" />
          Esta demanda foi concluída.
          <AppButton
            variant="ghost"
            label="Gerar relatório desta demanda"
            size="small"
            :icon="SparklesIcon"
            @click="report"
          />
        </div>
        <template v-else>
          <FormField
            id="move-actor"
            label="Responsável pela movimentação"
          >
            <InputText
              id="move-actor"
              :model-value="actor"
              fluid
              readonly
            />
          </FormField>
          <div
            v-if="demand.status === 'in_qa'"
            class="qa-tip"
          >
            <ClipboardDocumentCheckIcon style="width: 17px; height: 17px; flex-shrink: 0" />
            A conclusão exige um teste aprovado e nenhum reprovado no ciclo atual.
          </div>
          <FormField
            v-if="demand.status === 'in_qa'"
            id="move-reason"
            label="Motivo da devolução"
            class="mt-3"
          >
            <Textarea
              id="move-reason"
              v-model="moveReason"
              rows="2"
              fluid
              placeholder="Obrigatório se a demanda voltar para desenvolvimento"
            />
          </FormField>
          <div class="mt-4 flex flex-wrap gap-2">
            <AppButton
              v-for="next in transitions[demand.status]"
              :key="next"
              :data-testid="`move-${next}`"
              :label="moveText[next]"
              :variant="next === 'changes_requested' ? 'danger' : 'primary'"
              size="small"
              :loading="store.busy"
              :icon="next === 'changes_requested' ? ArrowUturnLeftIcon : ArrowRightIcon"
              @click="move(next)"
            />
          </div>
        </template>
      </section>
      <section class="drawer-section">
        <div class="flex items-center justify-between gap-3">
          <h3 class="block-title">
            <ClipboardDocumentCheckIcon style="width: 17px; height: 17px; flex-shrink: 0" />
            Homologação
            <span class="section-count">
              {{ activeTests.length }} no ciclo {{ demand.qa_cycle }}
            </span>
          </h3>
          <AppButton
            v-if="demand.status === 'in_qa'"
            variant="ghost"
            label="Adicionar teste"
            size="small"
            :icon="PlusIcon"
            @click="showTest = !showTest"
          />
        </div>
        <form
          v-if="showTest && demand.status === 'in_qa'"
          class="edit-block grid gap-5"
          @submit.prevent="saveTest"
        >
          <FormField
            id="test-title"
            label="Cenário"
            required
          >
            <InputText
              id="test-title"
              v-model="testForm.title"
              minlength="3"
              required
              fluid
              data-testid="test-title"
              placeholder="Ex.: Registro sem informação obrigatória"
            />
          </FormField>
          <FormField
            id="test-steps"
            label="Passos executados"
          >
            <Textarea
              id="test-steps"
              v-model="testForm.steps"
              rows="2"
              fluid
            />
          </FormField>
          <FormField
            id="test-expected"
            label="Resultado esperado"
          >
            <Textarea
              id="test-expected"
              v-model="testForm.expected"
              rows="2"
              fluid
            />
          </FormField>
          <FormField
            id="test-result"
            label="Resultado"
          >
            <Select
              v-model="testForm.result"
              input-id="test-result"
              aria-labelledby="test-result-label"
              :options="[
                { label: 'Aprovado', value: 'passed' },
                { label: 'Reprovado', value: 'failed' },
              ]"
              option-label="label"
              option-value="value"
              fluid
              data-testid="test-result"
            />
          </FormField>
          <FormField
            id="test-actual"
            label="Resultado observado"
            :required="testForm.result === 'failed'"
          >
            <Textarea
              id="test-actual"
              v-model="testForm.actual"
              :required="testForm.result === 'failed'"
              rows="2"
              fluid
              placeholder="Obrigatório para reprovação"
            />
          </FormField>
          <AppButton
            label="Registrar cenário"
            type="submit"
            data-testid="save-test"
            :loading="store.busy"
            :icon="CheckIcon"
          />
        </form>
        <div
          v-if="!store.detail.tests.length"
          class="empty-inline"
        >
          Nenhum cenário registrado. Os testes aparecerão aqui, separados por ciclo.
        </div>
        <div
          v-for="test in [...store.detail.tests].reverse()"
          :key="test.id"
          class="qa-test"
        >
          <span
            class="qa-result-icon"
            :class="test.result"
          >
            <CheckCircleIcon
              v-if="test.result === 'passed'"
              style="width: 17px; height: 17px; flex-shrink: 0"
            />
            <XCircleIcon
              v-else
              style="width: 17px; height: 17px; flex-shrink: 0"
            />
          </span>
          <div class="min-w-0 flex-1">
            <strong>{{ test.title }}</strong>
            <small>
              Ciclo {{ test.cycle }} · {{ test.author || 'Equipe' }} ·
              {{ dateLabel(test.created_at) }}
            </small>
            <p
              v-if="test.result === 'failed' && test.actual"
              class="test-error-note"
            >
              {{ test.actual }}
            </p>
          </div>
          <span
            class="qa-result-label"
            :class="test.result"
          >
            {{ test.result === 'passed' ? 'Aprovado' : 'Reprovado' }}
          </span>
        </div>
      </section>
      <section class="drawer-section last">
        <h3 class="block-title">
          <ClockIcon style="width: 17px; height: 17px; flex-shrink: 0" />
          Histórico de movimentações
        </h3>
        <div
          v-if="!store.detail.transitions.length"
          class="empty-inline"
        >
          O histórico começa na primeira movimentação.
        </div>
        <div class="history-list">
          <div
            v-for="event in [...store.detail.transitions].reverse()"
            :key="event.id"
            class="history-event"
          >
            <span class="history-node"></span>
            <div>
              <strong>
                {{ statusLabel(event.from_status) }} → {{ statusLabel(event.to_status) }}
              </strong>
              <small>{{ event.actor || 'Equipe' }} · {{ dateTime(event.created_at) }}</small>
              <p v-if="event.reason">{{ event.reason }}</p>
            </div>
          </div>
        </div>
      </section>
    </div>
  </Drawer>
</template>
