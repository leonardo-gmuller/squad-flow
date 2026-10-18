import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { api } from '../api'
import type { Member, Workstream, WorkstreamInput } from '../types'

export const useCatalogStore = defineStore('catalog', () => {
  const workstreams = ref<Workstream[]>([])
  const members = ref<Member[]>([])
  let revision = 0
  const kindOptions = computed(() =>
    workstreams.value.map((item) => ({
      value: item.code,
      label: item.name + (item.active ? '' : ' · Inativa'),
      color: item.color,
    })),
  )
  const activeKinds = computed(() =>
    kindOptions.value.filter(
      (option) => workstreams.value.find((item) => item.code === option.value)?.active,
    ),
  )
  const memberOptions = computed(() => [
    { id: '', label: 'Não atribuído' },
    ...members.value.map((item) => ({
      id: item.id,
      label:
        members.value.filter((other) => other.name === item.name).length > 1
          ? `${item.name} · ${item.email}`
          : item.name,
    })),
  ])
  const kindLabel = (code: string) =>
    workstreams.value.find((item) => item.code === code)?.name || code
  const kindStyle = (code: string) => {
    const color = workstreams.value.find((item) => item.code === code)?.color || '#8976e8'
    return { color, backgroundColor: `${color}20`, borderColor: `${color}45` }
  }
  const memberName = (id: string) => members.value.find((item) => item.id === id)?.name || id
  async function load() {
    const current = ++revision
    const [streams, people] = await Promise.all([api.workstreams(), api.members()])
    if (current !== revision) return
    workstreams.value = streams
    members.value = people
  }
  async function save(input: WorkstreamInput, id?: string) {
    const saved = id ? await api.updateWorkstream(id, input) : await api.createWorkstream(input)
    await load()
    return saved
  }
  function reset() {
    revision++
    workstreams.value = []
    members.value = []
  }
  return {
    workstreams,
    members,
    kindOptions,
    activeKinds,
    memberOptions,
    kindLabel,
    kindStyle,
    memberName,
    load,
    save,
    reset,
  }
})
