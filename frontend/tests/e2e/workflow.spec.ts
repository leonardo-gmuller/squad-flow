import { test, expect } from './fixtures'

// Requires the app + API + PostgreSQL running with synthetic seed.
test('cria melhoria, registra reprovação, corrige e conclui QA em novo ciclo', async ({ page }) => {
  const title = `E2E fluxo ${Date.now()}`
  await page.goto('/dashboard')
  await page.getByTestId('new-demand').click()
  await page.getByTestId('create-title').fill(title)
  await page.getByTestId('create-submit').click()
  const drawer = page.getByRole('dialog', { name: 'Detalhes da demanda' })
  await expect(drawer).toContainText(title)
  await page.getByTestId('move-in_progress').click()
  await page.getByTestId('move-in_qa').click()
  await page.getByRole('button', { name: 'Adicionar teste' }).click()
  await page.getByTestId('test-title').fill('Reprovação de campo obrigatório')
  await page.getByTestId('test-result').click()
  await page.getByRole('option', { name: 'Reprovado' }).click()
  await page.getByPlaceholder('Obrigatório para reprovação').fill('Resultado diferente do esperado')
  await page.getByTestId('save-test').click()
  await expect(drawer).toContainText('Devolvidas')
  await page.getByTestId('move-in_progress').click()
  await page.getByTestId('move-in_qa').click()
  await page.getByRole('button', { name: 'Adicionar teste' }).click()
  await page.getByTestId('test-title').fill('Cenário aprovado após correção')
  await page.getByTestId('save-test').click()
  await page.getByTestId('move-done').click()
  await expect(drawer).toContainText('Esta demanda foi concluída.')
})
