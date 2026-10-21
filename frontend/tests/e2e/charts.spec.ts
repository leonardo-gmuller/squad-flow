import { catalogResponses } from './catalog-fixtures'
import { test, expect, type Locator, type Page } from '@playwright/test'
import type { Metrics } from '../../src/types'

test.use({ timezoneId: 'America/Sao_Paulo' })
test.beforeEach(async ({ page }) => {
  await page.addInitScript(() => localStorage.setItem('squadflow:dark', 'false'))
})

const metrics: Metrics = {
  total: 10,
  active: 6,
  in_qa: 2,
  returned: 1,
  delivered: 4,
  first_pass_percent: 75,
  average_lead_hours: 24,
  by_kind: { improvement: 4, corrective: 3, testing: 2, rewrite: 1 },
  by_status: {},
  by_priority: {},
  overdue: 0,
  unassigned: 0,
  total_returns: 1,
  rework_percent: 10,
  on_time_percent: 100,
  deliveries_with_due_date: 4,
  delivered_30_days: 4,
  median_lead_hours: 24,
  p85_lead_hours: 48,
  average_active_age_hours: 12,
  daily_deliveries: Array.from({ length: 30 }, (_, i) => ({
    date: `2026-09-${String(i + 1).padStart(2, '0')}`,
    count: i % 3,
  })),
}
const report = {
  id: 'test-report',
  created_at: '2026-10-13T12:00:00Z',
  content: {
    title: 'Análise dos gráficos',
    summary: 'Dados de teste.',
    metrics: [],
    sections: [],
    charts: [
      {
        title: 'Variação por categoria',
        kind: 'bar',
        points: [
          { label: 'Redução', value: -2.5 },
          { label: 'Estável', value: 0 },
          { label: 'Aumento', value: 3.25 },
        ],
      },
      {
        title: 'Tendência',
        kind: 'line',
        points: metrics.daily_deliveries.map((day) => ({ label: day.date, value: day.count })),
      },
      { title: 'Período vazio', kind: 'line', points: [] },
      {
        title: 'Datas em inglês',
        kind: 'bar',
        points: [
          { label: 'October 16, 2026', value: 1 },
          { label: 'November 2026', value: 2 },
          { label: 'Monday', value: 3 },
        ],
      },
    ],
    tables: [
      {
        title: 'Datas do relatório',
        columns: ['Data', 'Horário'],
        rows: [['October 16, 2026', '2026-10-13T12:00:00Z']],
      },
    ],
    recommendations: [],
    missing_data: [],
    source_ids: [],
  },
}

async function mockWorkspace(page: Page, stats = metrics) {
  await page.routeWebSocket('**/api/v1/ws', (ws) => ws.close())
  await page.route('**/api/v1/**', async (route) => {
    const path = new URL(route.request().url()).pathname
    const responses: Record<string, unknown> = {
      ...catalogResponses,
      '/api/v1/auth/me': { user: { id: 'test-user', name: 'Pessoa de teste', role: 'member' } },
      '/api/v1/metrics': stats,
      '/api/v1/packages': [
        {
          id: 'test-package',
          reference: 'OS-10',
          title: 'OS de teste',
          initiative: 'Evolução',
          kind: 'improvement',
          done: 1,
          total: 2,
          in_qa: 0,
        },
      ],
      '/api/v1/reports/status': { enabled: true },
      '/api/v1/reports': [report],
      '/api/v1/reports/test-report': report,
    }
    await route.fulfill({ json: responses[path] ?? [] })
  })
}

async function expectPainted(canvas: Locator) {
  await expect(canvas).toBeVisible()
  await expect
    .poll(() =>
      canvas.evaluate((element: HTMLCanvasElement) => {
        const context = element.getContext('2d')!
        return context
          .getImageData(0, 0, element.width, element.height)
          .data.some((value, index) => index % 4 === 3 && value > 0)
      }),
    )
    .toBe(true)
}

test('dashboard desenha gráficos, atualiza o tema e cabe no celular', async ({ page }) => {
  const errors: string[] = []
  page.on('pageerror', (error) => errors.push(error.message))
  await mockWorkspace(page)
  await page.goto('/dashboard')
  const distribution = page.getByRole('img', { name: 'Distribuição por esteira', exact: true })
  const progress = page.getByRole('img', { name: 'Progresso das entregas', exact: true })
  await expectPainted(distribution)
  await expectPainted(progress)
  await expect(progress).toHaveAccessibleDescription('4 concluídas, 6 em aberto, 40% concluído')
  await expect(page.getByRole('table', { name: 'Distribuição por esteira' })).toContainText(
    'Corretivas3',
  )
  const lightImage = await distribution.evaluate((canvas: HTMLCanvasElement) => canvas.toDataURL())
  await page.getByRole('button', { name: 'Tema escuro', exact: true }).click()
  await expect(page.locator('html')).toHaveClass('app-dark')
  await expect
    .poll(() => distribution.evaluate((canvas: HTMLCanvasElement) => canvas.toDataURL()))
    .not.toBe(lightImage)
  await expectPainted(progress)
  await page
    .locator('.dashboard-grid')
    .screenshot({ path: 'test-results/charts-dashboard-dark.png' })
  await page.setViewportSize({ width: 390, height: 844 })
  await expectPainted(distribution)
  await expect
    .poll(() => page.evaluate(() => document.documentElement.scrollWidth))
    .toBeLessThanOrEqual(390)
  await page
    .locator('.dashboard-grid')
    .screenshot({ path: 'test-results/charts-dashboard-mobile.png' })
  expect(errors).toEqual([])
})

test('relatórios desenham linhas e barras com valores negativos e usam ProgressBar nas OS', async ({
  page,
}) => {
  const errors: string[] = []
  page.on('pageerror', (error) => errors.push(error.message))
  await mockWorkspace(page)
  await page.goto('/reports')
  await expectPainted(
    page.getByRole('img', { name: 'Entregas por dia · últimos 30 dias (UTC)', exact: true }),
  )
  const dailyDates = page.getByRole('table', {
    name: 'Entregas por dia · últimos 30 dias (UTC)',
    exact: true,
  })
  await expect(
    dailyDates.getByRole('rowheader', { name: '1 de setembro de 2026', exact: true }),
  ).toBeAttached()
  await page.getByRole('tab', { name: 'OS e equipe', exact: true }).click()
  await expectPainted(page.getByRole('img', { name: 'Volume por esteira', exact: true }))
  await expect(page.getByRole('progressbar', { name: 'Progresso da OS OS-10' })).toHaveAttribute(
    'aria-valuenow',
    '50',
  )
  await page.getByRole('tab', { name: 'Relatórios com IA', exact: true }).click()
  await page.getByLabel('Seus relatórios salvos').click()
  await page.getByRole('option', { name: /Análise dos gráficos/ }).click()
  const rendered = page.getByRole('article', { name: 'Relatório gerado' })
  await expectPainted(rendered.getByRole('img', { name: 'Variação por categoria', exact: true }))
  await expectPainted(rendered.getByRole('img', { name: 'Tendência', exact: true }))
  await expect(rendered.getByRole('table', { name: 'Variação por categoria' })).toContainText(
    'Redução-2,5',
  )
  await expectPainted(rendered.getByRole('img', { name: 'Datas em inglês', exact: true }))
  const dates = rendered.getByRole('table', { name: 'Datas em inglês', exact: true })
  await expect(dates).toContainText('16 de outubro de 2026')
  await expect(dates).toContainText('novembro de 2026')
  await expect(dates).toContainText('segunda-feira')
  await expect(rendered.locator('.overline')).toContainText('13/10/2026')
  const dateTable = rendered.getByRole('table', { name: 'Datas do relatório', exact: true })
  await expect(
    dateTable.getByRole('cell', { name: '16 de outubro de 2026', exact: true }),
  ).toBeVisible()
  await expect(dateTable.getByRole('cell', { name: '13/10/2026 09:00', exact: true })).toBeVisible()
  await expect(rendered.getByText('Sem dados para este gráfico.', { exact: true })).toBeVisible()
  await page.setViewportSize({ width: 390, height: 844 })
  await expect
    .poll(() => page.evaluate(() => document.documentElement.scrollWidth))
    .toBeLessThanOrEqual(390)
  await rendered.screenshot({ path: 'test-results/charts-reports-mobile.png' })
  await page.goto('/packages')
  await expect(page.getByRole('progressbar', { name: 'Progresso da OS OS-10' })).toHaveAttribute(
    'aria-valuenow',
    '50',
  )
  expect(errors).toEqual([])
})

test('dashboard vazio mantém progresso zero e renderiza sem erros', async ({ page }) => {
  const errors: string[] = []
  page.on('pageerror', (error) => errors.push(error.message))
  await mockWorkspace(page, {
    ...metrics,
    total: 0,
    active: 0,
    delivered: 0,
    by_kind: {},
    daily_deliveries: [],
  })
  await page.goto('/dashboard')
  const progress = page.getByRole('img', { name: 'Progresso das entregas', exact: true })
  await expectPainted(progress)
  await expect(progress).toHaveAccessibleDescription('0 concluídas, 0 em aberto, 0% concluído')
  await expectPainted(page.getByRole('img', { name: 'Distribuição por esteira', exact: true }))
  expect(errors).toEqual([])
})

test('gráfico de esteiras preserva o nome cadastrado mesmo quando parece uma data', async ({
  page,
}) => {
  await mockWorkspace(page, { ...metrics, by_kind: { custom: 2 } })
  await page.route('**/api/v1/workstreams', (route) =>
    route.fulfill({
      json: [
        {
          id: '66666666-6666-4666-8666-666666666605',
          code: 'custom',
          name: 'Monday',
          color: '#227d8d',
          active: true,
          created_at: '2026-10-16T12:00:00Z',
          updated_at: '2026-10-16T12:00:00Z',
        },
      ],
    }),
  )
  await page.goto('/dashboard')
  await expectPainted(page.getByRole('img', { name: 'Distribuição por esteira', exact: true }))
  const table = page.getByRole('table', { name: 'Distribuição por esteira', exact: true })
  await expect(table.getByRole('rowheader', { name: 'Monday', exact: true })).toBeAttached()
  await expect(table).toContainText('Monday2')
})
