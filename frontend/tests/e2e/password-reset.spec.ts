import { test, expect } from './fixtures'
import type { Page } from '@playwright/test'

const headers = { 'X-SquadFlow-Client': 'web' }
const original = 'OriginalTeste123!'
test.use({ actionTimeout: 10_000 })
async function createMember(page: Page) {
  const suffix = `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
  const response = await page.request.post('/api/v1/users', {
    headers,
    data: {
      name: `Recuperação ${suffix}`,
      email: `reset-${suffix}@example.com`,
      password: original,
    },
  })
  expect(response.status()).toBe(201)
  return response.json() as Promise<{ id: string; name: string; email: string }>
}
async function login(page: Page, email: string, password: string) {
  await page.goto('/login')
  await page.getByLabel('Email', { exact: true }).fill(email)
  await page.getByLabel('Senha', { exact: true }).fill(password)
  await page.getByRole('button', { name: 'Entrar', exact: true }).click()
}

test('admin altera senha e compartilha reset interno de uso único', async ({
  page,
  browser,
  baseURL,
}) => {
  test.setTimeout(90_000)
  const user = await createMember(page)
  const memberContext = await browser.newContext({ baseURL })
  try {
    const member = await memberContext.newPage()
    const urls: string[] = []
    member.on('request', (request) => urls.push(request.url()))
    await login(member, user.email, original)
    await expect(member).toHaveURL(new RegExp('/dashboard$'))
    await page.goto('/users')
    await page.getByRole('textbox', { name: 'Buscar usuários', exact: true }).fill(user.email)
    await page.getByRole('button', { name: `Alterar senha de ${user.name}`, exact: true }).click()
    const change = page.getByRole('dialog', { name: 'Alterar senha do usuário', exact: true })
    await change.getByLabel(/^Nova senha/).fill('AdminAlterada123!')
    await change.getByLabel(/^Confirmar nova senha/).fill('OutraSenha123!')
    await change.getByRole('button', { name: 'Salvar nova senha', exact: true }).click()
    await expect(change).toContainText('A confirmação deve ser igual à nova senha.')
    await change.getByLabel(/^Confirmar nova senha/).fill('AdminAlterada123!')
    await change.getByRole('button', { name: 'Salvar nova senha', exact: true }).click()
    await expect(change).toBeHidden()
    expect((await member.request.get('/api/v1/auth/me')).status()).toBe(401)
    await login(member, user.email, 'AdminAlterada123!')
    await expect(member).toHaveURL(new RegExp('/dashboard$'))

    await page
      .getByRole('button', { name: `Gerar link de redefinição para ${user.name}`, exact: true })
      .click()
    const reset = page.getByRole('dialog', { name: 'Redefinição interna de senha', exact: true })
    await reset.getByRole('button', { name: 'Gerar link', exact: true }).click()
    const link = await reset
      .getByRole('textbox', { name: 'Link de redefinição', exact: true })
      .inputValue()
    expect(link).toContain('/reset-password#token=')
    expect((await member.request.get('/api/v1/auth/me')).status()).toBe(200)
    await member.goto(link)
    await expect(member).toHaveURL(new RegExp('/reset-password$'))
    await member.getByLabel(/^Nova senha/).fill('UsuarioEscolheu123!')
    await member.getByLabel(/^Confirmar nova senha/).fill('UsuarioEscolheu123!')
    await member.getByRole('button', { name: 'Redefinir minha senha', exact: true }).click()
    await expect(member.getByRole('status')).toContainText('Senha redefinida.')
    const rawToken = new URLSearchParams(new URL(link).hash.slice(1)).get('token')!
    expect(urls.every((url) => !url.includes(rawToken))).toBe(true)
    await member.getByRole('link', { name: 'Ir para o login', exact: true }).click()
    await member.getByLabel('Email', { exact: true }).fill(user.email)
    await member.getByLabel('Senha', { exact: true }).fill('UsuarioEscolheu123!')
    await member.getByRole('button', { name: 'Entrar', exact: true }).click()
    await expect(member).toHaveURL(new RegExp('/dashboard$'))
    await member.goto(link)
    await member.getByLabel(/^Nova senha/).fill('NaoPodeReusar123!')
    await member.getByLabel(/^Confirmar nova senha/).fill('NaoPodeReusar123!')
    await member.getByRole('button', { name: 'Redefinir minha senha', exact: true }).click()
    await expect(member.getByRole('alert')).toContainText('inválido ou expirado')
    expect((await page.request.get('/api/v1/users')).status()).toBe(200)
  } finally {
    await memberContext.close().catch(() => {})
  }
})

test('membro altera sua senha com confirmação da senha atual', async ({ page }) => {
  test.setTimeout(60_000)
  const user = await createMember(page)
  const response = await page.request.post('/api/v1/auth/login', {
    headers,
    data: { email: user.email, password: original },
  })
  expect(response.status()).toBe(200)
  await page.goto('/account')
  await page.getByLabel(/^Senha atual/).fill('senha-incorreta')
  await page.getByLabel(/^Nova senha/).fill('MinhaAlteracao123!')
  await page.getByLabel(/^Confirmar nova senha/).fill('MinhaAlteracao123!')
  await page.getByRole('button', { name: 'Salvar nova senha', exact: true }).click()
  await expect(page.getByRole('alert').filter({ hasText: 'senha atual incorreta' })).toContainText(
    'senha atual incorreta',
  )
  expect((await page.request.get('/api/v1/auth/me')).status()).toBe(200)
  await page.getByLabel(/^Senha atual/).fill(original)
  await page.getByRole('button', { name: 'Salvar nova senha', exact: true }).click()
  await expect(page).toHaveURL(new RegExp('/login[?]password=changed$'))
  expect((await page.request.get('/api/v1/auth/me')).status()).toBe(401)
  await page.getByLabel('Email', { exact: true }).fill(user.email)
  await page.getByLabel('Senha', { exact: true }).fill('MinhaAlteracao123!')
  await page.getByRole('button', { name: 'Entrar', exact: true }).click()
  await expect(page).toHaveURL(new RegExp('/dashboard$'))
  expect(
    (
      await page.request.put(`/api/v1/users/${user.id}/password`, {
        headers,
        data: { password: 'ProibidaSenha123!' },
      })
    ).status(),
  ).toBe(403)
  expect(
    (
      await page.request.post(`/api/v1/users/${user.id}/password-reset`, { headers, data: {} })
    ).status(),
  ).toBe(403)
})
