import { test, expect } from '@playwright/test'

const report = {
  id: '00000000-0000-4000-8000-000000000001',
  demand_id: '',
  prompt: 'Resuma as entregas',
  created_at: '2026-10-13T12:00:00Z',
  content: {
    title: 'Relatório gerado com sucesso',
    summary: 'Análise pronta para visualização.',
    metrics: [],
    sections: [],
    charts: [],
    tables: [],
    recommendations: [],
    missing_data: [],
    source_ids: [],
  },
}
const metrics = {
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
}
const queued = (request: Record<string, unknown>, id = report.id) => ({
  id,
  request,
  status: 'queued',
  attempts: 0,
  max_attempts: 3,
  available_at: report.created_at,
  progress_agent: 'fila',
  progress_message: 'Aguardando processamento',
  error_message: '',
  report_id: '',
  created_at: report.created_at,
  updated_at: report.created_at,
  completed_at: null as string | null,
})

type MockJob = ReturnType<typeof queued>
async function mockAPI(
  page: import('@playwright/test').Page,
  generate: (route: import('@playwright/test').Route) => Promise<void>,
) {
  const state = { jobs: [] as MockJob[], reports: [] as (typeof report)[] }
  await page.route('**/api/v1/**', async (route) => {
    const path = new URL(route.request().url()).pathname
    if (path === '/api/v1/reports/generate') {
      await generate(route)
      return
    }
    const data =
      path === '/api/v1/auth/me'
        ? {
            user: {
              id: '55555555-5555-4555-8555-555555555501',
              name: 'Usuário de teste',
              email: 'test@example.com',
              role: 'member',
            },
          }
        : path === '/api/v1/metrics'
          ? metrics
          : path === '/api/v1/reports/status'
            ? { enabled: true, asynchronous: true }
            : path === '/api/v1/reports/jobs'
              ? state.jobs
              : path === `/api/v1/reports/${report.id}`
                ? report
                : path === '/api/v1/reports'
                  ? state.reports
                  : []
    await route.fulfill({ json: data })
  })
  return state
}

test('pedido pela tela usa a fila HTTP mesmo com WebSocket conectado', async ({ page }) => {
  let receivedPrompt = '',
    httpCalls = 0,
    socketOpened = false,
    socketReportRequests = 0
  const state = await mockAPI(page, async (route) => {
    httpCalls++
    const request = route.request().postDataJSON()
    receivedPrompt = request.prompt
    const job = queued(request)
    state.jobs = [job]
    await route.fulfill({ status: 202, json: job })
  })
  await page.routeWebSocket('**/api/v1/ws', (ws) => {
    socketOpened = true
    ws.onMessage((message) => {
      if (JSON.parse(String(message)).type === 'report.generate') socketReportRequests++
    })
  })
  await page.goto('/reports?tab=ai')
  await expect.poll(() => socketOpened).toBe(true)
  await page.getByLabel('Seu pedido').fill('Resuma as entregas')
  await page.getByRole('button', { name: 'Gerar relatório', exact: true }).click()
  await expect(page.getByTestId(`report-job-${report.id}`)).toContainText('Na fila')
  state.jobs = [
    {
      ...state.jobs[0],
      status: 'completed',
      report_id: report.id,
      completed_at: report.created_at,
    },
  ]
  state.reports = [report]
  await expect(page.getByRole('article', { name: 'Relatório gerado' })).toContainText(
    report.content.title,
  )
  expect(receivedPrompt).toBe('Resuma as entregas')
  expect(httpCalls).toBe(1)
  expect(socketReportRequests).toBe(0)
})

test('falha ao registrar pedido aparece na tela e em notificação', async ({ page }) => {
  await page.routeWebSocket('**/api/v1/ws', (ws) => ws.close())
  await mockAPI(page, async (route) =>
    route.fulfill({
      status: 503,
      json: { error: 'Não foi possível registrar a solicitação agora.' },
    }),
  )
  await page.goto('/reports?tab=ai')
  await page.getByLabel('Seu pedido').fill('Resuma as entregas')
  await page.getByRole('button', { name: 'Gerar relatório', exact: true }).click()
  await expect(
    page.getByText('Não foi possível registrar a solicitação', { exact: true }),
  ).toBeVisible()
  await expect(page.locator('.ai-reports .report-generation-error')).toContainText(
    'Não foi possível registrar a solicitação agora.',
  )
})

test('calendários validam o período e a continuação envia o relatório anterior como contexto', async ({
  page,
}) => {
  const requests: Record<string, unknown>[] = []
  await page.routeWebSocket('**/api/v1/ws', (ws) => ws.close())
  const state = await mockAPI(page, async (route) => {
    const request = route.request().postDataJSON()
    requests.push(request)
    const job = queued(
      request,
      requests.length === 1 ? report.id : '00000000-0000-4000-8000-000000000002',
    )
    await route.fulfill({ status: 202, json: job })
    state.jobs = [
      { ...job, status: 'completed', report_id: report.id, completed_at: report.created_at },
    ]
    state.reports = [report]
  })
  await page.goto('/reports?tab=ai')
  const month = await page.evaluate(() => {
    const date = new Date()
    return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`
  })
  async function choose(label: string, day: string) {
    await page.getByLabel(label, { exact: true }).click()
    const calendar = page.getByRole('dialog', { name: 'Escolher data', exact: true })
    await calendar.getByRole('gridcell', { name: day, exact: true }).click()
    await expect(calendar).toHaveCount(0)
  }
  await choose('De', '10')
  await choose('Até', '15')
  await choose('De', '20')
  await expect(
    page.getByText('A data final deve ser igual ou posterior à data inicial.', { exact: true }),
  ).toBeVisible()
  await expect(page.getByRole('button', { name: 'Gerar relatório', exact: true })).toBeDisabled()
  await choose('Até', '25')
  await page.getByLabel('Seu pedido').fill('Analise as entregas no período')
  await page.getByRole('button', { name: 'Gerar relatório', exact: true }).click()
  await expect(page.getByRole('article', { name: 'Relatório gerado' })).toBeVisible()
  expect(requests[0]).toMatchObject({ from: `${month}-20`, to: `${month}-25` })
  await page.getByRole('checkbox', { name: /Usar o resumo/ }).check()
  await page.getByLabel('Seu pedido').fill('Detalhe os gargalos encontrados')
  await page.getByRole('button', { name: 'Gerar relatório', exact: true }).click()
  await expect.poll(() => requests.length).toBe(2)
  expect(requests[1].previous_report_id).toBe(report.id)
})
