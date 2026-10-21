import { test, expect } from '@playwright/test'

test('PrimeVue Forms valida campos com Zod sem enviar credenciais inválidas', async ({ page }) => {
  let submissions = 0
  await page.route('**/api/v1/**', (route) => {
    if (new URL(route.request().url()).pathname === '/api/v1/auth/login') submissions++
    return route.fulfill({
      status: 401,
      json: {
        type: 'srn:error:unauthorized',
        code: 'auth:unauthorized',
        message: 'Sua sessão expirou.',
      },
    })
  })
  await page.goto('/login')
  await page.getByRole('button', { name: 'Entrar', exact: true }).click()
  await expect(page.getByText('Informe seu email.', { exact: true })).toBeVisible()
  await expect(page.getByText('Informe sua senha.', { exact: true })).toBeVisible()
  await expect(page.getByLabel('Email', { exact: true })).toBeFocused()
  await page.getByLabel('Email', { exact: true }).fill('email-invalido')
  await page.getByLabel('Senha', { exact: true }).fill('senha existente')
  await page.getByRole('button', { name: 'Entrar', exact: true }).click()
  await expect(page.getByText('Informe um email válido.', { exact: true })).toBeVisible()
  await page.getByLabel('Email', { exact: true }).fill('test@example.com')
  await page.getByLabel('Senha', { exact: true }).fill('á'.repeat(37))
  await page.getByRole('button', { name: 'Entrar', exact: true }).click()
  await expect(page.getByText('A senha excede o tamanho permitido.', { exact: true })).toBeVisible()
  expect(submissions).toBe(0)
})

test('credenciais incorretas mostram o erro público e permitem corrigir o pedido', async ({
  page,
}) => {
  let submissions = 0,
    refreshes = 0
  await page.route('**/api/v1/**', (route) => {
    const path = new URL(route.request().url()).pathname
    if (path === '/api/v1/auth/refresh') refreshes++
    if (path === '/api/v1/auth/login') {
      submissions++
      return route.fulfill({
        status: 401,
        json: {
          type: 'srn:error:unauthorized',
          code: 'auth:invalid-credentials',
          message: 'Email ou senha inválidos.',
        },
      })
    }
    return route.fulfill({
      status: 401,
      json: {
        type: 'srn:error:unauthorized',
        code: 'auth:unauthorized',
        message: 'Sua sessão expirou.',
      },
    })
  })
  const initialRefresh = page.waitForResponse(
    (response) => new URL(response.url()).pathname === '/api/v1/auth/refresh',
  )
  await page.goto('/login')
  await initialRefresh
  const previousRefreshes = refreshes
  await page.getByLabel('Email', { exact: true }).fill('test@example.com')
  await page.getByLabel('Senha', { exact: true }).fill('senha incorreta')
  await page.getByRole('button', { name: 'Entrar', exact: true }).click()
  await expect(page.getByRole('alert')).toContainText('Email ou senha inválidos.')
  await expect(page).toHaveURL(/\/login$/)
  await expect(page.getByRole('button', { name: 'Entrar', exact: true })).toBeEnabled()
  expect(submissions).toBe(1)
  expect(refreshes).toBe(previousRefreshes)
  await page.getByLabel('Senha', { exact: true }).fill('senha corrigida')
  await expect(page.getByText('Email ou senha inválidos.', { exact: true })).toBeHidden()
})

test('envio válido normaliza somente o email e preserva a senha', async ({ page }) => {
  const password = ' senha existente! '
  let received: Record<string, string> = {}
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
  await page.routeWebSocket('**/api/v1/ws', (ws) => ws.close())
  await page.route('**/api/v1/**', (route) => {
    const path = new URL(route.request().url()).pathname
    if (path === '/api/v1/auth/login') {
      received = route.request().postDataJSON()
      return route.fulfill({
        json: {
          user: { id: 'test-member', name: 'Teste', email: 'test@example.com', role: 'member' },
        },
      })
    }
    if (path === '/api/v1/auth/me' || path === '/api/v1/auth/refresh')
      return route.fulfill({
        status: 401,
        json: {
          message: 'Sessão expirada.',
          code: 'auth:unauthorized',
          type: 'srn:error:unauthorized',
        },
      })
    return route.fulfill({ json: path === '/api/v1/metrics' ? metrics : [] })
  })
  await page.goto('/login')
  await page.getByLabel('Email', { exact: true }).fill('  test@example.com  ')
  await page.getByLabel('Senha', { exact: true }).fill(password)
  await page.getByRole('button', { name: 'Entrar', exact: true }).click()
  await expect(page).toHaveURL(/\/dashboard$/)
  expect(received).toEqual({ email: 'test@example.com', password })
})
