import { test, expect } from './fixtures'

test('renderiza indicadores, tabelas e gráficos de relatório salvo', async ({ page }) => {
  const report = {
    id: '00000000-0000-4000-8000-000000000001',
    demand_id: '',
    prompt: 'Analise as entregas',
    created_at: '2026-10-13T12:00:00Z',
    content: {
      title: 'Análise de entregas',
      summary: 'Três demandas finalizadas no período.',
      metrics: [{ label: 'Concluídas', value: 3, unit: 'demandas' }],
      sections: [{ title: 'Evidências de QA', body: 'Os testes foram consultados.' }],
      charts: [
        { title: 'Entregas por esteira', kind: 'bar', points: [{ label: 'Melhorias', value: 3 }] },
      ],
      tables: [
        { title: 'Visão por equipe', columns: ['Equipe', 'Entregas'], rows: [['Squad', '3']] },
      ],
      recommendations: ['Revisar critérios de aceite.'],
      missing_data: ['Sem estimativas de esforço.'],
      source_ids: [],
    },
  }
  await page.route('**/api/v1/reports/status', (route) =>
    route.fulfill({ json: { enabled: true } }),
  )
  await page.route('**/api/v1/reports', (route) => route.fulfill({ json: [report] }))
  await page.route(`**/api/v1/reports/${report.id}`, (route) => route.fulfill({ json: report }))
  await page.goto('/reports')
  await page.getByRole('tab', { name: 'Fluxo e qualidade' }).click()
  await expect(page.getByText('Lead time P85', { exact: true })).toBeVisible()
  await expect(page.getByText('Entregas em 30 dias', { exact: true })).toBeVisible()
  await page.getByRole('tab', { name: 'Relatórios com IA' }).click()
  await page.getByLabel('Seus relatórios salvos').click()
  await page.getByRole('option', { name: /Análise de entregas/ }).click()
  const rendered = page.getByRole('article', { name: 'Relatório gerado' })
  await expect(rendered.getByRole('heading', { name: 'Análise de entregas' })).toBeVisible()
  await expect(rendered.getByRole('img', { name: 'Entregas por esteira' })).toBeVisible()
  await expect(rendered.getByRole('cell', { name: 'Squad', exact: true })).toBeVisible()
  await expect(rendered.getByText('Sem estimativas de esforço.', { exact: true })).toBeVisible()
  await page.evaluate(() => window.scrollTo(0, 0))
  await page.screenshot({ path: 'test-results/reports-desktop.png', fullPage: true })
  await page.setViewportSize({ width: 390, height: 844 })
  await expect(page.getByRole('button', { name: 'Abrir navegação' })).toBeVisible()
  await expect(page.locator('.shell-sidebar')).toHaveCSS('transform', 'matrix(1, 0, 0, 1, -273, 0)')
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(391)
  await page.screenshot({ path: 'test-results/reports-mobile.png', fullPage: true })
})

test('admin cadastra membro e somente admin vê a gestão de usuários', async ({ page }) => {
  const email = `e2e-${Date.now()}@example.com`,
    password = 'e2e-member-test-password'
  await page.goto('/users')
  await page.getByLabel('Nome', { exact: true }).fill('Membro E2E')
  await page.getByLabel('Email', { exact: true }).fill(email)
  await page.getByLabel('Senha inicial', { exact: true }).fill(password)
  await page.getByRole('button', { name: 'Cadastrar usuário' }).click()
  await expect(page.getByRole('cell', { name: email, exact: true })).toBeVisible()
  await page.getByRole('button', { name: 'Sair', exact: true }).click()
  await expect(page).toHaveURL(/\/login$/)
  await expect(page.getByRole('heading', { name: 'Entre no workspace.' })).toBeVisible()
  await page.getByLabel('Email', { exact: true }).fill(email)
  await page.getByLabel('Senha', { exact: true }).fill(password)
  await page.getByRole('button', { name: 'Entrar', exact: true }).click()
  await expect(page).toHaveURL(/\/dashboard$/)
  await expect(page.getByRole('link', { name: 'Usuários', exact: true })).toHaveCount(0)
  expect((await page.request.get('/api/v1/users')).status()).toBe(403)
  await page.goto('/users')
  await expect(page).toHaveURL(/\/dashboard$/)
})
