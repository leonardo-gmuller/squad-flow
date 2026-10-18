import { test, expect } from './fixtures'

test.use({ timezoneId: 'Pacific/Kiritimati', viewport: { width: 390, height: 844 } })

test('calendário preserva o prazo ao criar, editar e limpar uma demanda no celular', async ({
  page,
}) => {
  await page.goto('/dashboard')
  await page.getByTestId('new-demand').click()
  await page.getByTestId('create-title').fill(`Calendário E2E ${Date.now()}`)
  const month = await page.evaluate(() => {
    const date = new Date()
    return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`
  })
  await page.getByLabel('Data limite', { exact: true }).click()
  const calendar = page.getByRole('dialog', { name: 'Escolher data', exact: true })
  await expect(calendar.getByRole('button', { name: 'Hoje', exact: true })).toBeVisible()
  await calendar.getByRole('gridcell', { name: '15', exact: true }).click()
  await expect(calendar).toHaveCount(0)
  await expect(page.getByLabel('Data limite', { exact: true })).toHaveValue(
    `15/${month.slice(5)}/${month.slice(0, 4)}`,
  )
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(391)
  await page
    .getByRole('dialog', { name: 'Nova entrega', exact: true })
    .locator('.p-dialog-content')
    .evaluate((element) => {
      element.scrollTop = 0
    })
  await page.screenshot({ path: 'test-results/primevue-demand-mobile.png', animations: 'disabled' })

  const created = page.waitForResponse(
    (response) =>
      response.request().method() === 'POST' &&
      new URL(response.url()).pathname === '/api/v1/demands',
  )
  await page.getByTestId('create-submit').click()
  const response = await created
  expect(response.request().postDataJSON().due_date).toBe(`${month}-15T12:00:00Z`)
  expect(response.status()).toBe(201)
  const demand = await response.json()
  const drawer = page.getByRole('dialog', { name: 'Detalhes da demanda', exact: true })
  await expect(drawer).toBeVisible()
  await drawer.getByRole('button', { name: 'Editar entrega' }).click()
  await expect(drawer.getByLabel('Prazo', { exact: true })).toHaveValue(
    `15/${month.slice(5)}/${month.slice(0, 4)}`,
  )
  await drawer.getByLabel('Prazo', { exact: true }).click()
  await calendar.getByRole('gridcell', { name: '20', exact: true }).click()
  const updated = page.waitForResponse(
    (response) =>
      response.request().method() === 'PUT' && response.url().endsWith(`/demands/${demand.id}`),
  )
  await drawer.getByRole('button', { name: 'Salvar mudanças' }).click()
  expect((await updated).request().postDataJSON().due_date).toBe(`${month}-20T12:00:00Z`)
  await expect(drawer.getByRole('button', { name: 'Salvar mudanças' })).toHaveCount(0)
  await drawer.getByRole('button', { name: 'Editar entrega' }).click()
  await drawer.getByLabel('Prazo', { exact: true }).click()
  await calendar.getByRole('button', { name: 'Limpar', exact: true }).click()
  const cleared = page.waitForResponse(
    (response) =>
      response.request().method() === 'PUT' && response.url().endsWith(`/demands/${demand.id}`),
  )
  await drawer.getByRole('button', { name: 'Salvar mudanças' }).click()
  expect((await cleared).request().postDataJSON().due_date).toBeNull()
})
