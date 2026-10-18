import { test, expect } from './fixtures'
import type { Demand, Workstream } from '../../src/types'

test('admin cadastra, renomeia e desativa esteira; atribuição seleciona contas cadastradas', async ({
  page,
}) => {
  test.setTimeout(60_000)
  const marker = Date.now(),
    name = `Infraestrutura ${marker}`,
    renamed = `Plataforma ${marker}`
  await page.goto('/workstreams')
  await page.getByRole('button', { name: 'Nova esteira', exact: true }).click()
  const form = page.getByRole('dialog', { name: 'Nova esteira', exact: true })
  await form.getByRole('textbox', { name: 'Nome da esteira', exact: true }).fill(name)
  const created = page.waitForResponse(
    (response) =>
      response.request().method() === 'POST' &&
      new URL(response.url()).pathname === '/api/v1/workstreams',
  )
  await form.getByRole('button', { name: 'Salvar esteira', exact: true }).click()
  const response = await created
  expect(response.status()).toBe(201)
  const stream: Workstream = await response.json()
  await expect(form).toHaveCount(0)
  await expect(
    page.getByRole('table', { name: 'Esteiras cadastradas', exact: true }),
  ).toContainText(name)
  await page.goto('/tickets')
  await page.getByRole('button', { name: 'Nova entrega', exact: true }).click()
  const demandDialog = page.getByRole('dialog', { name: 'Nova entrega', exact: true })
  await demandDialog
    .getByRole('textbox', { name: 'Título da demanda', exact: true })
    .fill(`Entrega da plataforma ${marker}`)
  await demandDialog.getByLabel('Esteira', { exact: true }).click()
  await page.getByRole('option', { name, exact: true }).click()
  await demandDialog.getByLabel('Desenvolvedor', { exact: true }).click()
  await page.getByRole('option', { name: 'Alex', exact: true }).click()
  await expect(page.getByRole('option', { name: 'Alex', exact: true })).toHaveCount(0)
  await demandDialog.getByLabel('Responsável QA', { exact: true }).click()
  await page.getByRole('option', { name: 'Marina', exact: true }).click()
  const posted = page.waitForResponse(
    (response) =>
      response.request().method() === 'POST' &&
      new URL(response.url()).pathname === '/api/v1/demands',
  )
  await demandDialog.getByRole('button', { name: 'Criar entrega', exact: true }).click()
  const demandResponse = await posted
  expect(demandResponse.status()).toBe(201)
  const item: Demand = await demandResponse.json()
  expect(demandResponse.request().postDataJSON()).toMatchObject({
    kind: stream.code,
    assignee_id: '55555555-5555-4555-8555-555555555501',
    reviewer_id: '55555555-5555-4555-8555-555555555504',
  })
  const drawer = page.getByRole('dialog', { name: 'Detalhes da demanda', exact: true })
  await expect(drawer).toContainText(name)
  await expect(drawer).toContainText('Alex')
  await expect(drawer).toContainText('Marina')
  await drawer.getByRole('button', { name: 'Editar entrega', exact: true }).click()
  await drawer.getByLabel('Desenvolvedor', { exact: true }).click()
  await page.getByRole('option', { name: 'Daniel', exact: true }).click()
  await drawer.getByRole('button', { name: 'Salvar mudanças', exact: true }).click()
  await expect(drawer.getByRole('button', { name: 'Salvar mudanças', exact: true })).toHaveCount(0)
  await expect(drawer).toContainText('Daniel')
  await page.getByRole('button', { name: 'Fechar detalhes', exact: true }).click()
  await page.goto('/dashboard')
  await expect(
    page.getByRole('table', { name: 'Distribuição por esteira', exact: true }),
  ).toContainText(name)
  await page.goto('/board')
  await page.getByRole('button', { name: 'Filtros', exact: true }).click()
  await page.getByLabel('Filtrar por esteira', { exact: true }).click()
  await page.getByRole('option', { name, exact: true }).click()
  await page.getByLabel('Filtrar por responsável', { exact: true }).click()
  await page.getByRole('option', { name: 'Daniel', exact: true }).click()
  await expect(page.locator('.work-card')).toHaveCount(1)
  await expect(page.locator('.work-card')).toContainText(item.title)
  await page.goto('/workstreams')
  await page.getByRole('button', { name: `Editar esteira ${name}`, exact: true }).click()
  const edit = page.getByRole('dialog', { name: 'Editar esteira', exact: true })
  await edit.getByRole('textbox', { name: 'Nome da esteira', exact: true }).fill(renamed)
  await edit.getByRole('button', { name: 'Salvar esteira', exact: true }).click()
  await expect(edit).toHaveCount(0)
  await page.getByRole('button', { name: `Desativar esteira ${renamed}`, exact: true }).click()
  const deactivate = page.getByRole('dialog', { name: 'Desativar esteira', exact: true })
  await deactivate.getByRole('button', { name: 'Desativar esteira', exact: true }).click()
  await expect(deactivate).toHaveCount(0)
  await expect(page.getByRole('row').filter({ hasText: renamed })).toContainText('Inativa')
  const saved = await page.request.get(`/api/v1/demands/${item.id}`)
  expect((await saved.json()).demand).toMatchObject({
    kind: stream.code,
    assignee_id: '55555555-5555-4555-8555-555555555503',
  })
  await page.goto('/tickets')
  await page.getByRole('button', { name: 'Nova entrega', exact: true }).click()
  await page
    .getByRole('dialog', { name: 'Nova entrega', exact: true })
    .getByLabel('Esteira', { exact: true })
    .click()
  await expect(page.getByRole('option', { name: renamed, exact: true })).toHaveCount(0)
  await page.keyboard.press('Escape')
  await page
    .getByRole('dialog', { name: 'Nova entrega', exact: true })
    .getByRole('button', { name: 'Cancelar', exact: true })
    .click()
  await page.goto('/workstreams')
  await page.getByRole('button', { name: `Reativar esteira ${renamed}`, exact: true }).click()
  await expect(page.getByRole('row').filter({ hasText: renamed })).toContainText('Ativa')
  await page.screenshot({ path: 'test-results/workstreams-desktop.png', fullPage: true })
  await page.setViewportSize({ width: 390, height: 844 })
  await expect
    .poll(() => page.evaluate(() => document.documentElement.scrollWidth))
    .toBeLessThanOrEqual(390)
  await page.screenshot({ path: 'test-results/workstreams-mobile.png', fullPage: true })
})

test('membro usa o catálogo e as atribuições sem acessar a administração', async ({ page }) => {
  const marker = Date.now(),
    email = `catalog-member-${marker}@example.test`,
    password = 'Catalogo-Teste-2026!'
  const response = await page.request.post('/api/v1/users', {
    headers: { 'X-SquadFlow-Client': 'web' },
    data: { name: 'Membro do catálogo', email, password },
  })
  expect(response.status()).toBe(201)
  await page.goto('/dashboard')
  await page.getByRole('button', { name: 'Sair', exact: true }).click()
  await page.getByLabel('Email', { exact: true }).fill(email)
  await page.getByLabel('Senha', { exact: true }).fill(password)
  await page.getByRole('button', { name: 'Entrar', exact: true }).click()
  await expect(page).toHaveURL(/\/dashboard$/)
  await expect(page.getByRole('link', { name: 'Esteiras', exact: true })).toHaveCount(0)
  expect((await page.request.get('/api/v1/workstreams')).status()).toBe(200)
  expect((await page.request.get('/api/v1/members')).status()).toBe(200)
  const blocked = await page.request.post('/api/v1/workstreams', {
    headers: { 'X-SquadFlow-Client': 'web' },
    data: { name: 'Não autorizado', color: '#8976e8' },
  })
  expect(blocked.status()).toBe(403)
  await page.goto('/workstreams')
  await expect(page).toHaveURL(/\/dashboard$/)
})
