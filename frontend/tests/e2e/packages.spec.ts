import { test, expect } from './fixtures'

test('cria OS compartilhada e ticket corretivo avulso', async ({ page }) => {
  const reference = `OS-E2E-${Date.now()}`
  await page.goto('/packages')
  await page.getByTestId('new-package').click()
  await page.getByTestId('package-reference').fill(reference)
  await page.getByTestId('package-title').fill('Regras de validação de processo')
  await page.getByTestId('save-package').click()
  await expect(page.getByText(reference).first()).toBeVisible()
  await page.getByTestId('new-package-item').click()
  await page.getByTestId('create-title').fill('Implementar validação no XML')
  await page.getByLabel('Desenvolvedor', { exact: true }).click()
  await page.getByRole('option', { name: 'Daniel', exact: true }).click()
  await page.getByTestId('create-submit').click()
  await expect(page.getByRole('dialog', { name: 'Detalhes da demanda' })).toContainText(
    'Implementar validação no XML',
  )
  await page.getByRole('button', { name: 'Fechar detalhes' }).click()
  await page.getByTestId('new-demand').click()
  await page.getByTestId('create-title').fill('Corrigir falha no gateway')
  await page.getByTestId('create-kind').click()
  await page.getByRole('option', { name: 'Corretivas' }).click()
  await page.getByTestId('create-submit').click()
  await expect(page.getByRole('dialog', { name: 'Detalhes da demanda' })).toContainText(
    'Ticket avulso / sem OS',
  )
})

test('navegação SPA mantém URL por módulo', async ({ page }) => {
  await page.goto('/dashboard')
  await page
    .getByRole('navigation', { name: 'Navegação principal' })
    .getByRole('link', { name: 'Kanban', exact: true })
    .click()
  await expect(page).toHaveURL(/\/board$/)
  await page.getByTestId('nav-packages').click()
  await expect(page).toHaveURL(/\/packages$/)
  await page.reload()
  await expect(page.getByRole('heading', { name: /OS e iniciativas/ }).first()).toBeVisible()
})
