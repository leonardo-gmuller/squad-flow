import type { Kind, Priority, Status } from '../types'
import { useCatalogStore } from '../stores/catalog'
export const stages: { value: Status; label: string; subtitle: string; tone: string }[] = [
  { value: 'backlog', label: 'Backlog', subtitle: 'Aguardando início', tone: 'slate' },
  { value: 'in_progress', label: 'Em desenvolvimento', subtitle: 'Em execução', tone: 'blue' },
  { value: 'in_qa', label: 'Em QA', subtitle: 'Em homologação', tone: 'purple' },
  {
    value: 'changes_requested',
    label: 'Devolvidas',
    subtitle: 'Precisam de ajustes',
    tone: 'orange',
  },
  { value: 'done', label: 'Concluídas', subtitle: 'Entregas aprovadas', tone: 'green' },
]
export const priorities: { value: Priority; label: string }[] = [
  { value: 'low', label: 'Baixa' },
  { value: 'medium', label: 'Média' },
  { value: 'high', label: 'Alta' },
  { value: 'critical', label: 'Crítica' },
]
export const transitions: Record<Status, Status[]> = {
  backlog: ['in_progress'],
  in_progress: ['in_qa'],
  in_qa: ['changes_requested', 'done'],
  changes_requested: ['in_progress'],
  done: [],
}
export const kindLabel = (v: Kind) => useCatalogStore().kindLabel(v)
export const statusLabel = (v: Status) => stages.find((s) => s.value === v)?.label ?? v
export const priorityLabel = (v: Priority) => priorities.find((p) => p.value === v)?.label ?? v

export const kindStyle = (v: Kind) => useCatalogStore().kindStyle(v)
