import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { api } from '../api'
import { useCatalogStore } from './catalog'
import type {
  CreateInput,
  CreatePackage,
  Release,
  ReleaseInput,
  Demand,
  Detail,
  Metrics,
  Status,
  WorkPackage,
} from '../types'

const emptyMetrics = (): Metrics => ({
  total: 0,
  active: 0,
  in_qa: 0,
  returned: 0,
  delivered: 0,
  first_pass_percent: 0,
  average_lead_hours: 0,
  by_kind: {},
  by_status: {},
  by_priority: {},
  overdue: 0,
  unassigned: 0,
  total_returns: 0,
  rework_percent: 0,
  on_time_percent: 0,
  deliveries_with_due_date: 0,
  delivered_30_days: 0,
  median_lead_hours: 0,
  p85_lead_hours: 0,
  average_active_age_hours: 0,
  daily_deliveries: [],
})

export const useWorkspaceStore = defineStore('workspace', () => {
  const catalog = useCatalogStore()
  const items = ref<Demand[]>([])
  const releases = ref<Release[]>([])
  const boardMode = ref(false)
  const dataRevision = ref(0)
  const defaultReleaseId = ref('')
  const activeReleases = computed(() => releases.value.filter((release) => !release.archived_at))
  const releaseById = (id?: string) => releases.value.find((release) => release.id === id)
  function setBoardMode(value: boolean) {
    boardMode.value = value
  }
  const packages = ref<WorkPackage[]>([])
  const metrics = ref<Metrics>(emptyMetrics())
  const detail = ref<Detail | null>(null)
  const selectedId = ref<string | null>(null)
  const createDemandVisible = ref(false)
  const createPackageVisible = ref(false)
  const defaultPackageId = ref('')
  const loading = ref(false)
  const loaded = ref(false)
  const error = ref('')
  const busy = ref(false)
  let revision = 0
  const owners = computed(() =>
    [...new Set(items.value.map((d) => d.assignee).filter(Boolean))].sort(),
  )
  const qaOwners = computed(() =>
    [...new Set(items.value.map((d) => d.reviewer).filter(Boolean))].sort(),
  )
  const packageById = (id?: string) => packages.value.find((p) => p.id === id)
  const itemsInPackage = (id: string) => items.value.filter((i) => i.package_id === id)

  async function load() {
    const current = ++revision
    loading.value = true
    error.value = ''
    try {
      const [demands, workPackages, stats, versions] = await Promise.all([
        boardMode.value ? Promise.resolve([]) : api.list(),
        api.packages(),
        api.metrics(),
        api.releases(),
        catalog.load(),
      ])
      if (current !== revision) return
      items.value = demands ?? []
      packages.value = workPackages ?? []
      metrics.value = stats
      releases.value = versions ?? []
      dataRevision.value++
      loaded.value = true
    } catch (err) {
      if (current === revision) error.value = (err as Error).message
      throw err
    } finally {
      if (current === revision) loading.value = false
    }
  }
  async function selectDemand(id: string) {
    selectedId.value = id
    detail.value = null
    try {
      const result = await api.detail(id)
      if (selectedId.value === id) detail.value = result
    } catch (err) {
      if (selectedId.value === id) {
        selectedId.value = null
        error.value = (err as Error).message
      }
    }
  }
  function closeDemand() {
    selectedId.value = null
    detail.value = null
  }
  async function refresh() {
    await load()
    const id = selectedId.value
    if (id) {
      const result = await api.detail(id)
      if (selectedId.value === id) detail.value = result
    }
  }
  function reset() {
    catalog.reset()
    revision++
    items.value = []
    releases.value = []
    defaultReleaseId.value = ''
    boardMode.value = false
    packages.value = []
    metrics.value = emptyMetrics()
    closeDemand()
    loaded.value = false
    loading.value = false
    error.value = ''
    createDemandVisible.value = false
    createPackageVisible.value = false
  }
  function openCreate(packageId = '', releaseId = '') {
    defaultPackageId.value = packageId
    defaultReleaseId.value = releaseId
    createDemandVisible.value = true
  }
  async function change(action: () => Promise<unknown>, refreshDetail = false) {
    busy.value = true
    try {
      const result = await action()
      await load()
      const id = selectedId.value
      if (refreshDetail && id) {
        const result = await api.detail(id)
        if (selectedId.value === id) detail.value = result
      }
      return result
    } finally {
      busy.value = false
    }
  }
  async function createDemand(input: CreateInput) {
    return change(() => api.create(input)) as Promise<Demand>
  }
  async function createPackage(input: CreatePackage) {
    return change(() => api.createPackage(input)) as Promise<WorkPackage>
  }
  async function editDemand(id: string, input: Record<string, unknown>) {
    return change(() => api.edit(id, input), true)
  }
  async function moveDemand(id: string, status: Status, actor: string, reason: string) {
    return change(() => api.move(id, status, actor, reason), true)
  }
  async function addTest(
    id: string,
    input: {
      title: string
      steps: string
      expected: string
      actual: string
      result: 'passed' | 'failed'
      author: string
    },
  ) {
    return change(() => api.addTest(id, input), true)
  }
  async function createRelease(input: ReleaseInput) {
    return change(() => api.createRelease(input)) as Promise<Release>
  }
  async function editRelease(id: string, input: ReleaseInput) {
    return change(() => api.editRelease(id, input))
  }
  async function archiveRelease(id: string, archived: boolean) {
    return change(() => api.archiveRelease(id, archived))
  }
  return {
    releases,
    activeReleases,
    releaseById,
    createRelease,
    editRelease,
    archiveRelease,
    setBoardMode,
    dataRevision,
    defaultReleaseId,
    items,
    packages,
    metrics,
    detail,
    selectedId,
    createDemandVisible,
    createPackageVisible,
    defaultPackageId,
    loading,
    loaded,
    error,
    busy,
    owners,
    qaOwners,
    packageById,
    itemsInPackage,
    load,
    refresh,
    reset,
    selectDemand,
    closeDemand,
    openCreate,
    createDemand,
    createPackage,
    editDemand,
    moveDemand,
    addTest,
  }
})
