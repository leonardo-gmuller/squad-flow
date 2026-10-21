import { test, expect } from './fixtures'
import type { Release } from '../../src/types'

// These tests use the real API and a database with synthetic data.
test('vincula demandas, consulta entregas e arquiva/restaura uma versão sem perder o histórico', async ({
  page,
}) => {
  test.setTimeout(90_000)
  const marker = Date.now().toString(),
    name = `v-e2e-${marker}`,
    title = `Entrega versionada ${marker}`
  await page.goto('/releases')
  await page.getByRole('button', { name: 'Nova versão', exact: true }).click()
  await page.getByRole('textbox', { name: 'Nome da versão', exact: true }).fill(name)
  await page.getByLabel('Descrição', { exact: true }).fill('Notas da versão de teste')
  const createdVersion = page.waitForResponse(
    (response) =>
      response.request().method() === 'POST' &&
      new URL(response.url()).pathname === '/api/v1/releases',
  )
  await page.getByRole('button', { name: 'Salvar versão', exact: true }).click()
  const version: Release = await (await createdVersion).json()
  await expect(page.getByRole('heading', { name, exact: true })).toBeVisible()
  await page.getByRole('button', { name: 'Adicionar demanda', exact: true }).click()
  await expect(page.getByLabel('Versão de entrega', { exact: true })).toContainText(name)
  await page.getByTestId('create-title').fill(title)
  const createdDemand = page.waitForResponse(
    (response) =>
      response.request().method() === 'POST' &&
      new URL(response.url()).pathname === '/api/v1/demands',
  )
  await page.getByTestId('create-submit').click()
  const demand = await (await createdDemand).json()
  expect(demand.release_id).toBe(version.id)
  const drawer = page.getByRole('dialog', { name: 'Detalhes da demanda', exact: true })
  await expect(drawer).toContainText(name)
  await drawer.getByRole('button', { name: 'Editar entrega', exact: true }).click()
  await drawer.getByLabel('Versão de entrega', { exact: true }).click()
  await page.getByRole('option', { name: 'Sem versão', exact: true }).click()
  const cleared = page.waitForResponse(
    (response) =>
      response.request().method() === 'PUT' &&
      new URL(response.url()).pathname === `/api/v1/demands/${demand.id}`,
  )
  await drawer.getByRole('button', { name: 'Salvar mudanças', exact: true }).click()
  expect((await cleared).request().postDataJSON().release_id).toBe('')
  await expect(drawer.getByRole('button', { name: 'Salvar mudanças', exact: true })).toHaveCount(0)
  await drawer.getByRole('button', { name: 'Editar entrega', exact: true }).click()
  await drawer.getByLabel('Versão de entrega', { exact: true }).click()
  await page.getByRole('option', { name, exact: true }).click()
  await drawer.getByRole('button', { name: 'Salvar mudanças', exact: true }).click()
  await expect(drawer.getByRole('button', { name: 'Salvar mudanças', exact: true })).toHaveCount(0)
  for (const status of ['in_progress', 'in_qa']) await page.getByTestId(`move-${status}`).click()
  await drawer.getByRole('button', { name: 'Adicionar teste', exact: true }).click()
  await page.getByTestId('test-title').fill('Cenário da entrega versionada')
  await page.getByTestId('save-test').click()
  await page.getByTestId('move-done').click()
  await expect(drawer).toContainText('Esta demanda foi concluída.')
  await page.getByRole('button', { name: 'Fechar detalhes', exact: true }).click()
  const pending = await page.request.post('/api/v1/demands', {
    headers: { 'X-SquadFlow-Client': 'web' },
    data: {
      title: `Entrega pendente ${marker}`,
      kind: 'corrective',
      priority: 'medium',
      release_id: version.id,
    },
  })
  expect(pending.status()).toBe(201)
  await expect(
    page.getByRole('progressbar', { name: `Progresso da versão ${name}`, exact: true }),
  ).toHaveAttribute('aria-valuenow', '50')
  await page.getByRole('button', { name: `Arquivar versão ${name}`, exact: true }).click()
  const confirmation = page.getByRole('dialog', { name: 'Arquivar versão', exact: true })
  await expect(confirmation).toContainText('1 demanda aberta')
  await confirmation.getByRole('button', { name: 'Arquivar versão', exact: true }).click()
  await expect(confirmation).toHaveCount(0)
  await page.goto(`/board?release_id=${version.id}`)
  await expect(
    page.getByTestId('board-done').getByRole('button', { name: new RegExp(title) }),
  ).toHaveCount(0)
  await expect(page.getByRole('status').filter({ hasText: '0 entregas' })).toBeVisible()
  await page.getByRole('button', { name: 'Filtros', exact: true }).click()
  await page.getByLabel('Visibilidade das versões', { exact: true }).click()
  await page.getByRole('option', { name: 'Versões arquivadas', exact: true }).click()
  await expect(
    page.getByTestId('board-done').getByRole('button', { name: new RegExp(title) }),
  ).toBeVisible()
  await expect(
    page
      .getByTestId('board-backlog')
      .getByRole('button', { name: new RegExp(`^Abrir .*Entrega pendente ${marker}`) }),
  ).toBeVisible()
  await page.goto('/releases')
  await page.getByLabel('Situação das versões', { exact: true }).click()
  await page.getByRole('option', { name: 'Arquivadas', exact: true }).click()
  await page.getByRole('button', { name: `Restaurar versão ${name}`, exact: true }).click()
  await expect(
    page.getByRole('button', { name: `Restaurar versão ${name}`, exact: true }),
  ).toHaveCount(0)
  await page.goto(`/board?release_id=${version.id}`)
  await expect(
    page.getByTestId('board-done').getByRole('button', { name: new RegExp(title) }),
  ).toBeVisible()
  await page
    .getByTestId('board-done')
    .getByRole('button', { name: new RegExp(title) })
    .click()
  await expect(drawer).toContainText('Esta demanda foi concluída.')
  await expect(drawer).toContainText('Cenário da entrega versionada')
  await expect(drawer).toContainText('Em QA → Concluídas')
  await page.getByRole('button', { name: 'Fechar detalhes', exact: true }).click()
  await page.setViewportSize({ width: 390, height: 844 })
  await expect(page.locator('.shell-sidebar')).toHaveCSS('transform', 'matrix(1, 0, 0, 1, -273, 0)')
  await page.screenshot({
    path: 'test-results/releases-board-mobile.png',
    fullPage: true,
    animations: 'disabled',
  })
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(390)
  await page.goto(`/releases?release_id=${version.id}`)
  await expect(
    page.getByRole('table', { name: `Demandas da versão ${name}`, exact: true }),
  ).toContainText(title)
  await page.screenshot({
    path: 'test-results/releases-mobile.png',
    fullPage: true,
    animations: 'disabled',
  })
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(390)
})

test('Kanban carrega 20 cartões por coluna, pagina e filtra sem baixar a lista completa', async ({
  page,
}) => {
  // Other tests write to the same workspace; their events would reset the loaded pages.
  await page.routeWebSocket('**/api/v1/ws', (ws) => ws.close())
  const marker = `Paginacao-${Date.now()}`
  const versionResponse = await page.request.post('/api/v1/releases', {
    headers: { 'X-SquadFlow-Client': 'web' },
    data: { name: marker },
  })
  expect(versionResponse.status()).toBe(201)
  const version: Release = await versionResponse.json()
  await Promise.all(
    Array.from({ length: 42 }, async (_, index) => {
      const response = await page.request.post('/api/v1/demands', {
        headers: { 'X-SquadFlow-Client': 'web' },
        data: {
          title: `${marker} item ${index}`,
          kind: 'improvement',
          priority: 'medium',
          assignee_id: '55555555-5555-4555-8555-555555555501',
          release_id: version.id,
        },
      })
      expect(response.status()).toBe(201)
    }),
  )
  const unversioned = await page.request.post('/api/v1/demands', {
    headers: { 'X-SquadFlow-Client': 'web' },
    data: { title: `${marker} sem versão`, kind: 'corrective', priority: 'medium' },
  })
  expect(unversioned.status()).toBe(201)
  let fullListRequests = 0
  page.on('request', (request) => {
    if (request.method() === 'GET' && new URL(request.url()).pathname === '/api/v1/demands')
      fullListRequests++
  })
  await page.goto(`/board?release_id=${version.id}&q=${marker}`)
  const backlog = page.getByTestId('board-backlog')
  await expect(backlog.locator('.work-card')).toHaveCount(20)
  await expect(backlog.locator('.column-count')).toHaveText('42')
  await backlog.getByRole('button', { name: 'Carregar mais em Backlog', exact: true }).click()
  await expect(backlog.locator('.work-card')).toHaveCount(40)
  await backlog.getByRole('button', { name: 'Carregar mais em Backlog', exact: true }).click()
  await expect(backlog.locator('.work-card')).toHaveCount(42)
  await expect(
    backlog.getByRole('button', { name: 'Carregar mais em Backlog', exact: true }),
  ).toHaveCount(0)
  await page.getByLabel('Filtrar por versão', { exact: true }).click()
  await page.getByRole('option', { name: 'Sem versão', exact: true }).click()
  await expect(backlog.locator('.work-card')).toHaveCount(1)
  await expect(backlog).toContainText(`${marker} sem versão`)
  await page
    .getByLabel('Buscar demandas no quadro', { exact: true })
    .fill('busca sem resultado 123xyz')
  await expect(backlog.locator('.work-card')).toHaveCount(0)
  await expect(backlog).toContainText('Nenhuma entrega nesta etapa.')
  expect(fullListRequests).toBe(0)
})
