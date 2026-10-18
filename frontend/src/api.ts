import type {
  NotificationPage,
  Workstream,
  WorkstreamInput,
  Member,
  Board,
  BoardFilter,
  Release,
  ReleaseInput,
  CreateInput,
  Demand,
  Detail,
  Metrics,
  Status,
  TestCase,
  WorkPackage,
  CreatePackage,
  User,
  GeneratedReport,
  ReportRequest,
  ReportJob,
} from './types'

export class APIError extends Error {
  constructor(
    message: string,
    public status: number,
    public code = '',
    public type = '',
    public requestID = '',
  ) {
    super(message)
    this.name = 'APIError'
  }
}
let refreshPromise: Promise<void> | null = null
async function doRefresh() {
  const response = await fetch('/api/v1/auth/refresh', {
    method: 'POST',
    credentials: 'same-origin',
    headers: { 'Content-Type': 'application/json', 'X-SquadFlow-Client': 'web' },
    body: '{}',
  })
  if (!response.ok) throw new APIError('Sua sessão expirou. Entre novamente.', response.status)
}
export function refreshSession(): Promise<void> {
  if (!refreshPromise) {
    refreshPromise = (async () => {
      if (navigator.locks)
        await navigator.locks.request('squadflow:refresh', async () => {
          await doRefresh()
        })
      else await doRefresh()
    })().finally(() => {
      refreshPromise = null
    })
  }
  return refreshPromise
}

async function request<T>(
  path: string,
  init?: RequestInit,
  retry = true,
  format: 'json' | 'text' = 'json',
): Promise<T> {
  const response = await fetch(`/api/v1${path}`, {
    ...init,
    credentials: 'same-origin',
    headers: { 'Content-Type': 'application/json', 'X-SquadFlow-Client': 'web', ...init?.headers },
  })
  if (!response.ok) {
    if (response.status === 401 && retry && !['/auth/login', '/auth/logout'].includes(path)) {
      try {
        await refreshSession()
      } catch (error) {
        window.dispatchEvent(new Event('squadflow:session-expired'))
        throw error
      }
      return request<T>(path, init, false, format)
    }
    const body = (await response.json().catch(() => ({}))) as {
      type?: string
      code?: string
      message?: string
      error?: string
    } | null
    if (response.status === 401 && path !== '/auth/login')
      window.dispatchEvent(new Event('squadflow:session-expired'))
    const message =
      typeof body?.message === 'string'
        ? body.message
        : typeof body?.error === 'string'
          ? body.error
          : `Erro HTTP ${response.status}`
    throw new APIError(
      message,
      response.status,
      typeof body?.code === 'string' ? body.code : '',
      typeof body?.type === 'string' ? body.type : '',
      response.headers.get('X-Request-ID') || '',
    )
  }
  return (format === 'text' ? response.text() : response.json()) as Promise<T>
}
export const api = {
  notifications: (filter: 'all' | 'unread', before = '', signal?: AbortSignal) =>
    request<NotificationPage>(
      `/notifications?${new URLSearchParams({ filter, limit: '20', ...(before ? { before } : {}) })}`,
      { signal },
    ),
  notificationCount: () => request<{ count: number }>('/notifications/unread-count'),
  readNotification: (id: string) =>
    request<{ read_at: string }>(`/notifications/${encodeURIComponent(id)}/read`, {
      method: 'POST',
      body: '{}',
    }),
  readAllNotifications: () =>
    request<{ read_at: string }>('/notifications/read-all', { method: 'POST', body: '{}' }),
  workstreams: () => request<Workstream[]>('/workstreams'),
  members: () => request<Member[]>('/members'),
  createWorkstream: (input: WorkstreamInput) =>
    request<Workstream>('/workstreams', { method: 'POST', body: JSON.stringify(input) }),
  updateWorkstream: (id: string, input: WorkstreamInput) =>
    request<Workstream>(`/workstreams/${encodeURIComponent(id)}`, {
      method: 'PUT',
      body: JSON.stringify(input),
    }),
  me: () => request<{ user: User }>('/auth/me'),
  login: (email: string, password: string) =>
    request<{ user: User }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }),
  logout: () => request('/auth/logout', { method: 'POST', body: '{}' }),
  users: () => request<User[]>('/users'),
  createUser: (payload: { name: string; email: string; password: string }) =>
    request<User>('/users', { method: 'POST', body: JSON.stringify(payload) }),
  setUserPassword: (id: string, password: string) =>
    request(`/users/${encodeURIComponent(id)}/password`, {
      method: 'PUT',
      body: JSON.stringify({ password }),
    }),
  createPasswordReset: (id: string) =>
    request<{ token: string; expires_at: string }>(
      `/users/${encodeURIComponent(id)}/password-reset`,
      { method: 'POST', body: '{}' },
    ),
  changePassword: (current_password: string, password: string) =>
    request('/auth/password', {
      method: 'PUT',
      body: JSON.stringify({ current_password, password }),
    }),
  resetPassword: (token: string, password: string) =>
    request<{ ok: boolean; user_id: string }>('/auth/password-reset', {
      method: 'POST',
      body: JSON.stringify({ token, password }),
    }),
  reportStatus: () => request<{ enabled: boolean }>('/reports/status'),
  reports: (demandId = '') =>
    request<GeneratedReport[]>(
      `/reports${demandId ? `?demand_id=${encodeURIComponent(demandId)}` : ''}`,
    ),
  report: (id: string) => request<GeneratedReport>(`/reports/${encodeURIComponent(id)}`),
  generateReport: (payload: ReportRequest, key: string = crypto.randomUUID()) =>
    request<ReportJob>('/reports/generate', {
      method: 'POST',
      headers: { 'Idempotency-Key': key },
      body: JSON.stringify(payload),
    }),
  reportJobs: () => request<ReportJob[]>('/reports/jobs'),
  reportJob: (id: string) => request<ReportJob>(`/reports/jobs/${encodeURIComponent(id)}`),
  retryReportJob: (id: string, key: string = crypto.randomUUID()) =>
    request<ReportJob>(`/reports/jobs/${encodeURIComponent(id)}/retry`, {
      method: 'POST',
      headers: { 'Idempotency-Key': key },
      body: '{}',
    }),
  reportHTML: (id: string) =>
    request<string>(`/reports/${encodeURIComponent(id)}/html?download=1`, undefined, true, 'text'),
  async exportCSV() {
    await api.me()
    const response = await fetch('/api/v1/reports/demands.csv', { credentials: 'same-origin' })
    if (!response.ok) throw new APIError('Não foi possível exportar as demandas.', response.status)
    const url = URL.createObjectURL(await response.blob())
    const link = document.createElement('a')
    link.href = url
    link.download = 'squadflow-demandas.csv'
    link.click()
    setTimeout(() => URL.revokeObjectURL(url), 1000)
  },
  releases: () => request<Release[]>('/releases'),
  createRelease: (input: ReleaseInput) =>
    request<Release>('/releases', { method: 'POST', body: JSON.stringify(input) }),
  editRelease: (id: string, input: ReleaseInput) =>
    request(`/releases/${encodeURIComponent(id)}`, { method: 'PUT', body: JSON.stringify(input) }),
  archiveRelease: (id: string, archived: boolean) =>
    request(`/releases/${encodeURIComponent(id)}/archive`, {
      method: 'POST',
      body: JSON.stringify({ archived }),
    }),
  board: (filter: BoardFilter, signal?: AbortSignal) => {
    const query = new URLSearchParams(
      Object.entries(filter)
        .filter(([, value]) => value !== undefined && value !== '')
        .map(([key, value]) => [key, String(value)]),
    )
    return request<Board>(`/board?${query}`, { signal })
  },
  packages: () => request<WorkPackage[]>('/packages'),
  createPackage: (payload: CreatePackage) =>
    request<WorkPackage>('/packages', { method: 'POST', body: JSON.stringify(payload) }),

  list: (query = '') => request<Demand[]>(`/demands${query}`),
  metrics: () => request<Metrics>('/metrics'),
  detail: (id: string) => request<Detail>(`/demands/${encodeURIComponent(id)}`),
  create: (payload: CreateInput) =>
    request<Demand>('/demands', { method: 'POST', body: JSON.stringify(payload) }),
  edit: (id: string, payload: Record<string, unknown>) =>
    request<Demand>(`/demands/${encodeURIComponent(id)}`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    }),
  move: (id: string, status: Status, actor: string, reason = '') =>
    request<Demand>(`/demands/${encodeURIComponent(id)}/move`, {
      method: 'POST',
      body: JSON.stringify({ status, actor, reason }),
    }),
  addTest: (
    id: string,
    payload: {
      title: string
      steps: string
      expected: string
      actual: string
      result: 'passed' | 'failed'
      author: string
    },
  ) =>
    request<TestCase>(`/demands/${encodeURIComponent(id)}/tests`, {
      method: 'POST',
      body: JSON.stringify(payload),
    }),
}
