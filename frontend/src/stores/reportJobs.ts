import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { api } from '../api'
import type { ReportJob, ReportRequest } from '../types'

export const useReportJobsStore = defineStore('report-jobs', () => {
  const jobs = ref<ReportJob[]>([]),
    error = ref(''),
    initialized = ref(false)
  const active = computed(() =>
    jobs.value.find((job) => job.status === 'queued' || job.status === 'processing'),
  )
  let timer: ReturnType<typeof setInterval> | undefined,
    owner = '',
    revision = 0,
    loading = false,
    refreshedAt = 0
  async function refresh() {
    if (!owner || loading) return
    const version = revision
    loading = true
    try {
      const rows = await api.reportJobs()
      if (version !== revision) return
      jobs.value = rows
      error.value = ''
      initialized.value = true
      refreshedAt = Date.now()
    } catch (err) {
      if (version === revision) error.value = (err as Error).message
    } finally {
      if (version === revision) loading = false
    }
  }
  function stop() {
    clearInterval(timer)
    owner = ''
    revision++
    loading = false
    jobs.value = []
    error.value = ''
    initialized.value = false
    refreshedAt = 0
  }
  function start(userID: string) {
    if (owner === userID) return
    stop()
    owner = userID
    void refresh()
    timer = setInterval(() => {
      if (active.value || Date.now() - refreshedAt >= 15000) void refresh()
    }, 2500)
  }
  function add(job: ReportJob) {
    jobs.value = [job, ...jobs.value.filter((row) => row.id !== job.id)].slice(0, 50)
  }
  async function submit(request: ReportRequest, key: string) {
    const version = revision
    const job = await api.generateReport(request, key)
    if (version === revision && owner) add(job)
    return job
  }
  async function retry(id: string) {
    const version = revision
    const job = await api.retryReportJob(id)
    if (version === revision && owner) add(job)
    return job
  }
  return { jobs, error, initialized, active, start, stop, refresh, submit, retry }
})
