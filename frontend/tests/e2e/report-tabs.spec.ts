import { test, expect } from '@playwright/test'

const metrics = {
  total: 5,
  active: 5,
  in_qa: 1,
  returned: 1,
  delivered: 0,
  first_pass_percent: 0,
  average_lead_hours: 0,
  by_kind: { improvement: 2, corrective: 1, rewrite: 1, testing: 1 },
  by_status: { backlog: 2, in_progress: 1, in_qa: 1, changes_requested: 1, done: 0 },
  by_priority: { critical: 1, high: 2, medium: 1, low: 1 },
  overdue: 0,
  unassigned: 0,
  total_returns: 0,
  rework_percent: 0,
  on_time_percent: 0,
  deliveries_with_due_date: 0,
  delivered_30_days: 0,
  median_lead_hours: 0,
  p85_lead_hours: 0,
  average_active_age_hours: 72,
  daily_deliveries: [
    { date: '2026-10-16', count: 0 },
    { date: '2026-10-17', count: 0 },
  ],
}

test.beforeEach(async ({ page }) => {
  await page.route('**/api/v1/**', async (route) => {
    const path = new URL(route.request().url()).pathname
    const data =
      path === '/api/v1/auth/me'
        ? {
            user: {
              id: '55555555-5555-4555-8555-555555555501',
              name: 'Alex',
              email: 'alex@example.test',
              role: 'member',
            },
          }
        : path === '/api/v1/metrics'
          ? metrics
          : path === '/api/v1/reports/status'
            ? { enabled: true, asynchronous: true }
            : []
    await route.fulfill({ json: data })
  })
  await page.routeWebSocket('**/api/v1/ws', (ws) => ws.close())
})

test('abas organizam todas as informações, exibem gráficos reais e preservam o pedido', async ({
  page,
}, testInfo) => {
  await page.goto('/reports')
  await expect(page.getByRole('tab', { name: 'Visão geral', exact: true })).toHaveAttribute(
    'aria-selected',
    'true',
  )
  await expect(page.getByRole('img', { name: 'Demandas por status', exact: true })).toBeVisible()
  await expect(
    page.getByRole('img', { name: 'Demandas por prioridade', exact: true }),
  ).toBeVisible()
  await expect(page.getByRole('heading', { name: 'O que você quer analisar?' })).toBeHidden()
  await page.screenshot({
    path: testInfo.outputPath('reports-overview-desktop.png'),
    fullPage: true,
  })
  await page.getByRole('tab', { name: 'Fluxo e qualidade' }).click()
  await expect(page.getByText('Lead time P85', { exact: true })).toBeVisible()
  await expect(page.getByText('Entregas em 30 dias', { exact: true })).toBeVisible()
  await expect(page.getByRole('img', { name: 'Trabalho em QA e ajustes' })).toBeVisible()
  await expect(
    page.getByText('Uma demanda pode aparecer nos dois grupos.', { exact: false }),
  ).toBeVisible()
  await page.getByRole('tab', { name: 'OS e equipe' }).click()
  await expect(page.getByRole('heading', { name: 'Progresso por OS' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Entregas por responsável' })).toBeVisible()
  await page.getByRole('tab', { name: 'Relatórios com IA' }).click()
  await page.getByLabel('Seu pedido').fill('Analise as demandas de demonstração')
  await page.getByRole('tab', { name: 'Visão geral', exact: true }).click()
  await page.getByRole('tab', { name: 'Relatórios com IA' }).click()
  await expect(page.getByLabel('Seu pedido')).toHaveValue('Analise as demandas de demonstração')
  await expect(page).toHaveURL(/tab=ai/)
  await page.reload()
  await expect(page.getByRole('tab', { name: 'Relatórios com IA' })).toHaveAttribute(
    'aria-selected',
    'true',
  )
})

test('abas e gráficos funcionam em celular sem transbordar a página', async ({
  page,
}, testInfo) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await page.goto('/reports')
  await expect(page.getByRole('img', { name: 'Demandas por status', exact: true })).toBeVisible()
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(391)
  await page.screenshot({
    path: testInfo.outputPath('reports-overview-mobile.png'),
    fullPage: true,
  })
  await page.getByRole('tab', { name: 'Relatórios com IA' }).click()
  await expect(page.getByLabel('Seu pedido')).toBeVisible()
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(391)
  await page.screenshot({ path: testInfo.outputPath('reports-ai-mobile.png'), fullPage: true })
})
