import { test, expect } from './fixtures'
import type { User, Demand } from '../../src/types'

const headers = { 'X-SquadFlow-Client': 'web' }

test('membro recebe notificações, lê, pagina e abre a demanda; central funciona no celular', async ({
  page,
  request,
}) => {
  test.setTimeout(90_000)
  const email = `notifications-${Date.now()}@example.test`,
    password = 'Notificacoes-Teste-2026!'
  const login = await request.post('/api/v1/auth/login', {
    headers,
    data: { email: process.env.E2E_EMAIL, password: process.env.E2E_PASSWORD },
  })
  expect(login.status()).toBe(200)
  const createdUser = await request.post('/api/v1/users', {
    headers,
    data: { name: 'Pessoa das notificações', email, password },
  })
  expect(createdUser.status()).toBe(201)
  const user: User = await createdUser.json()
  await page.goto('/dashboard')
  await page.getByRole('button', { name: 'Sair', exact: true }).click()
  await page.getByLabel('Email', { exact: true }).fill(email)
  await page.getByLabel('Senha', { exact: true }).fill(password)
  await page.getByRole('button', { name: 'Entrar', exact: true }).click()
  await expect(page).toHaveURL(/\/dashboard$/)
  const bell = () => page.getByRole('button', { name: /^Notificações(?:,|$)/ })
  await expect(bell()).toHaveAccessibleName('Notificações')
  await expect(page.getByText('Ao vivo', { exact: true })).toHaveCount(0)
  async function assign(title: string): Promise<Demand> {
    const response = await request.post('/api/v1/demands', {
      headers,
      data: { title, kind: 'improvement', priority: 'medium', assignee_id: user.id },
    })
    expect(response.status()).toBe(201)
    return response.json()
  }
  const item = await assign('Revisar validação do cadastro de clientes')
  await expect(bell()).toHaveAccessibleName('Notificações, 1 não lida')
  await page.screenshot({
    path: 'test-results/notifications-bell-desktop.png',
    animations: 'disabled',
  })
  await bell().click()
  const center = page.getByRole('dialog', { name: 'Central de notificações', exact: true })
  await expect(center).toContainText('Nova atribuição')
  await expect(center).toContainText(item.title)
  await center.getByRole('button', { name: 'Não lidas', exact: true }).click()
  await center
    .getByRole('button', { name: 'Marcar como lida: Nova atribuição', exact: true })
    .click()
  await expect(
    center.getByRole('heading', { name: 'Nenhuma notificação não lida', exact: true }),
  ).toBeVisible()
  await expect(bell()).toHaveAccessibleName('Notificações')
  const moved = await request.post(`/api/v1/demands/${item.id}/move`, {
    headers,
    data: { status: 'in_progress' },
  })
  expect(moved.status()).toBe(200)
  await expect(bell()).toHaveAccessibleName('Notificações, 1 não lida')
  await expect(center).toContainText('Etapa atualizada')
  await center.getByRole('button', { name: /^Abrir notificação: Etapa atualizada\./ }).click()
  await expect(center).toHaveCount(0)
  const drawer = page.getByRole('dialog', { name: 'Detalhes da demanda', exact: true })
  await expect(drawer).toContainText(item.title)
  await page.getByRole('button', { name: 'Fechar detalhes', exact: true }).click()
  await expect(bell()).toHaveAccessibleName('Notificações')
  await page.reload()
  await expect(bell()).toHaveAccessibleName('Notificações')
  const persisted = await page.request.get('/api/v1/notifications?filter=unread')
  expect((await persisted.json()).items).toEqual([])
  await Promise.all(
    Array.from({ length: 22 }, (_, index) => assign(`Acompanhar entrega de integração ${index}`)),
  )
  await expect(bell()).toHaveAccessibleName('Notificações, 22 não lidas')
  await bell().click()
  await center.getByRole('button', { name: 'Não lidas', exact: true }).click()
  await expect(center.getByRole('listitem')).toHaveCount(20)
  await center.getByRole('button', { name: 'Carregar mais notificações', exact: true }).click()
  await expect(center.getByRole('listitem')).toHaveCount(22)
  await expect(
    center.getByRole('button', { name: 'Carregar mais notificações', exact: true }),
  ).toHaveCount(0)
  await center.getByRole('button', { name: 'Marcar todas como lidas', exact: true }).click()
  await expect(
    center.getByRole('heading', { name: 'Nenhuma notificação não lida', exact: true }),
  ).toBeVisible()
  await expect(bell()).toHaveAccessibleName('Notificações')
  await assign('Conferir critérios de aceite da próxima versão')
  await expect(center).toContainText('Conferir critérios de aceite da próxima versão')
  await page.screenshot({
    path: 'test-results/notifications-desktop-dark.png',
    animations: 'disabled',
  })
  await page.setViewportSize({ width: 390, height: 844 })
  await expect
    .poll(() => page.evaluate(() => document.documentElement.scrollWidth))
    .toBeLessThanOrEqual(390)
  await page.screenshot({
    path: 'test-results/notifications-mobile-dark.png',
    animations: 'disabled',
  })
  await center.getByRole('button', { name: 'Fechar notificações', exact: true }).click()
  await page.getByRole('button', { name: 'Tema claro', exact: true }).click()
  await bell().click()
  await expect(center).toContainText('Conferir critérios de aceite da próxima versão')
  await page.screenshot({
    path: 'test-results/notifications-mobile-light.png',
    animations: 'disabled',
  })
})

test('admin não tem botão nem acesso à central de notificações', async ({ page }) => {
  let inboxRequests = 0
  page.on('request', (req) => {
    if (new URL(req.url()).pathname.startsWith('/api/v1/notifications')) inboxRequests++
  })
  await page.goto('/dashboard')
  await expect(page.getByRole('heading', { name: 'Visão geral.', exact: true })).toBeVisible()
  await expect(page.getByRole('button', { name: /^Notificações(?:,|$)/ })).toHaveCount(0)
  await expect(page.getByText('Ao vivo', { exact: true })).toHaveCount(0)
  expect(inboxRequests).toBe(0)
  expect((await page.request.get('/api/v1/notifications')).status()).toBe(403)
  expect((await page.request.get('/api/v1/notifications/unread-count')).status()).toBe(403)
  expect(
    (await page.request.post('/api/v1/notifications/read-all', { headers, data: {} })).status(),
  ).toBe(403)
})

test('central mostra falha e permite tentar novamente sem perder a sessão', async ({ page }) => {
  let fail = true
  await page.routeWebSocket('**/api/v1/ws', (ws) => ws.close())
  await page.route('**/api/v1/**', async (route) => {
    const path = new URL(route.request().url()).pathname
    if (path === '/api/v1/auth/me')
      return route.fulfill({
        json: {
          user: {
            id: '11111111-1111-4111-8111-111111111111',
            name: 'Pessoa de teste',
            role: 'member',
          },
        },
      })
    if (path === '/api/v1/notifications/unread-count') return route.fulfill({ json: { count: 0 } })
    if (path === '/api/v1/notifications')
      return fail
        ? route.fulfill({ status: 503, json: { error: 'Não foi possível carregar notificações' } })
        : route.fulfill({ json: { items: [], unread_count: 0, next_cursor: '' } })
    return route.fulfill({ json: [] })
  })
  await page.goto('/account')
  await page.getByRole('button', { name: 'Notificações', exact: true }).click()
  const center = page.getByRole('dialog', { name: 'Central de notificações', exact: true })
  await expect(
    center.getByText('Não foi possível carregar notificações', { exact: true }),
  ).toBeVisible()
  fail = false
  await center.getByRole('button', { name: 'Tentar novamente', exact: true }).click()
  await expect(
    center.getByRole('heading', { name: 'Nenhuma notificação por enquanto', exact: true }),
  ).toBeVisible()
  await expect(page).toHaveURL(/\/account$/)
  await page.keyboard.press('Escape')
  await expect(center).toHaveCount(0)
  await expect(page.getByRole('button', { name: 'Notificações', exact: true })).toBeFocused()
})
