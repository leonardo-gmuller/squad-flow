import { test, expect } from '@playwright/test'

test('botão auxiliar não envia o login e carregamento impede reenvio até o erro', async ({
  page,
}) => {
  let submissions = 0
  let finishLogin!: () => void
  const pendingLogin = new Promise<void>((resolve) => {
    finishLogin = resolve
  })
  await page.route('**/api/v1/**', async (route) => {
    if (new URL(route.request().url()).pathname === '/api/v1/auth/login') {
      submissions++
      await pendingLogin
      await route.fulfill({ status: 400, json: { error: 'Credenciais de teste inválidas' } })
    } else await route.fulfill({ status: 401, json: { error: 'Sem sessão' } })
  })
  await page.goto('/login')
  await page.getByLabel('Email', { exact: true }).fill('button-test@example.com')
  await page.getByLabel('Senha', { exact: true }).fill('Senha-ficticia-123!')
  const reveal = page.getByRole('button', { name: 'Mostrar senha', exact: true })
  await expect(reveal).toHaveAttribute('type', 'button')
  await reveal.click()
  await expect(page.getByLabel('Senha', { exact: true })).toHaveAttribute('type', 'text')
  expect(submissions).toBe(0)
  const submit = page.getByRole('button', { name: 'Entrar', exact: true })
  await submit.press('Enter')
  await expect(submit).toBeDisabled()
  await expect(submit).toHaveAttribute('aria-busy', 'true')
  await expect(submit.locator('svg')).toBeVisible()
  await expect(submit).toContainText('Entrar')
  expect(submissions).toBe(1)
  finishLogin()
  await expect(page.getByText('Credenciais de teste inválidas', { exact: true })).toBeVisible()
  await expect(submit).toBeEnabled()
  await expect(submit).not.toHaveAttribute('aria-busy', 'true')
})

test('botão de diálogo envia o formulário externo e mantém cancelar como ação auxiliar', async ({
  page,
}) => {
  let submissions = 0
  let finishSave!: () => void
  const pendingSave = new Promise<void>((resolve) => {
    finishSave = resolve
  })
  await page.routeWebSocket('**/api/v1/ws', (ws) => ws.close())
  await page.route('**/api/v1/**', async (route) => {
    const path = new URL(route.request().url()).pathname
    if (path === '/api/v1/releases' && route.request().method() === 'POST') {
      submissions++
      expect(route.request().postDataJSON().name).toBe('1.0.0')
      await pendingSave
      await route.fulfill({ status: 400, json: { error: 'Erro de teste ao salvar' } })
    } else
      await route.fulfill({
        json:
          path === '/api/v1/auth/me'
            ? { user: { id: 'button-test', name: 'Pessoa de teste', role: 'admin' } }
            : [],
      })
  })
  await page.goto('/releases')
  await page.getByRole('button', { name: 'Nova versão', exact: true }).click()
  const dialog = page.getByRole('dialog', { name: 'Nova versão', exact: true })
  const cancel = dialog.getByRole('button', { name: 'Cancelar', exact: true })
  await expect(cancel).toHaveAttribute('type', 'button')
  await cancel.click()
  await expect(dialog).toHaveCount(0)
  expect(submissions).toBe(0)
  await page.getByRole('button', { name: 'Nova versão', exact: true }).click()
  await dialog.getByLabel('Nome da versão').fill('1.0.0')
  const save = dialog.getByRole('button', { name: 'Salvar versão', exact: true })
  await expect(save).toHaveAttribute('form', 'release-form')
  await save.click()
  await expect(save).toBeDisabled()
  await expect(save).toHaveAttribute('aria-busy', 'true')
  finishSave()
  await expect(dialog.getByText('Erro de teste ao salvar', { exact: true })).toBeVisible()
  await expect(save).toBeEnabled()
  expect(submissions).toBe(1)
})
