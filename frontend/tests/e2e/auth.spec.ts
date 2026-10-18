import { test, expect } from '@playwright/test'

test('exige login, usa cookies HttpOnly e limpa a sessão ao sair', async ({ page, context }) => {
  const email = process.env.E2E_EMAIL,
    password = process.env.E2E_PASSWORD
  if (!email || !password) throw new Error('Configure E2E_EMAIL e E2E_PASSWORD.')
  await page.goto('/reports')
  await expect(page).toHaveURL(/\/login/)
  await page.getByLabel('Email', { exact: true }).fill(email)
  await page.getByLabel('Senha', { exact: true }).fill(password)
  await expect(page.getByRole('button', { name: 'Mostrar senha', exact: true })).toHaveCSS(
    'position',
    'absolute',
  )
  await page.getByRole('button', { name: 'Mostrar senha', exact: true }).click()
  await expect(page.getByLabel('Senha', { exact: true })).toHaveAttribute('type', 'text')
  await page.getByRole('button', { name: 'Ocultar senha', exact: true }).click()
  await expect(page.getByLabel('Senha', { exact: true })).toHaveAttribute('type', 'password')
  await expect(page.getByLabel('Senha', { exact: true })).toHaveAttribute(
    'autocomplete',
    'current-password',
  )
  await page.getByRole('button', { name: 'Entrar', exact: true }).click()
  await expect(page).toHaveURL(/\/reports$/)
  await expect(page.getByRole('heading', { name: 'Relatórios.', exact: true })).toBeVisible()
  const cookies = await context.cookies()
  for (const name of ['sf_access', 'sf_refresh']) {
    const cookie = cookies.find((c) => c.name === name)
    expect(cookie?.httpOnly).toBe(true)
    expect(cookie?.sameSite).toBe('Strict')
  }
  await expect(
    page.getByRole('navigation', { name: 'Navegação principal' }).getByRole('link'),
  ).toHaveCount(10)
  await expect(page.getByText('ACESSO RÁPIDO', { exact: true })).toHaveCount(0)
  await expect(page.getByRole('button', { name: 'Gerar relatório', exact: true })).toBeDisabled()
  await page.getByRole('button', { name: 'Sair', exact: true }).click()
  await expect(page).toHaveURL(/\/login$/)
  expect((await context.cookies()).filter((c) => c.name.startsWith('sf_'))).toHaveLength(0)
  expect((await page.request.get('/api/v1/demands')).status()).toBe(401)
})
