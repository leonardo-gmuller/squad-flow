import { test, expect } from './fixtures'
import type { Demand, Detail, Status, User } from '../../src/types'

test.use({ viewport: { width: 1920, height: 1080 } })
test('arraste persiste histórico, devolução e aprovação de QA na API real', async ({ page }) => {
  test.setTimeout(90_000)
  const headers = { 'X-SquadFlow-Client': 'web' },
    marker = Date.now()
  const actor: { user: User } = await (await page.request.get('/api/v1/auth/me')).json()
  const actorLabel = `${actor.user.name} (${actor.user.id})`
  const releaseResponse = await page.request.post('/api/v1/releases', {
    headers,
    data: { name: `Drag E2E ${marker}` },
  })
  expect(releaseResponse.status()).toBe(201)
  const release = await releaseResponse.json()
  const created = await page.request.post('/api/v1/demands', {
    headers,
    data: {
      title: `Entrega drag ${marker}`,
      kind: 'improvement',
      priority: 'medium',
      release_id: release.id,
    },
  })
  expect(created.status()).toBe(201)
  const item: Demand = await created.json()
  let fullLists = 0
  page.on('request', (request) => {
    if (request.method() === 'GET' && new URL(request.url()).pathname === '/api/v1/demands')
      fullLists++
  })
  await page.goto(`/board?release_id=${release.id}`)
  const card = (status: Status) => page.getByTestId(`board-${status}`).locator('.work-card')
  async function drop(from: Status, to: Status) {
    await expect(card(from)).toBeVisible()
    const source = (await card(from).boundingBox())!,
      target = (await page.getByTestId(`board-${to}`).boundingBox())!
    await page.mouse.move(source.x + source.width / 2, source.y + 60)
    await page.mouse.down()
    await page.mouse.move(target.x + target.width / 2, target.y + 100, { steps: 12 })
    await expect(page.locator('.kanban-drag-preview')).toBeVisible()
    await expect(page.getByTestId(`board-${to}`)).toHaveClass(/is-drop-over/)
    await page.mouse.up()
  }
  for (const [from, to] of [
    ['backlog', 'in_progress'],
    ['in_progress', 'in_qa'],
  ] as [Status, Status][]) {
    await drop(from, to)
    await expect(card(to)).toBeVisible()
  }
  await drop('in_qa', 'changes_requested')
  const dialog = page.getByRole('dialog', { name: 'Mover entrega', exact: true })
  await dialog.getByLabel(/^Motivo da devolução/).fill('Corrigir validação antes de aprovar')
  await dialog.getByRole('button', { name: 'Mover entrega', exact: true }).click()
  await expect(card('changes_requested')).toBeVisible()
  await expect(page.locator('.p-dialog-mask')).toHaveCount(0)
  for (const [from, to] of [
    ['changes_requested', 'in_progress'],
    ['in_progress', 'in_qa'],
  ] as [Status, Status][]) {
    await drop(from, to)
    await expect(card(to)).toBeVisible()
  }
  const rejected = page.waitForResponse(
    (response) =>
      response.request().method() === 'POST' && response.url().endsWith(`/demands/${item.id}/move`),
  )
  await drop('in_qa', 'done')
  expect((await rejected).status()).toBe(409)
  await expect(
    page.getByRole('alert').filter({ hasText: 'ao menos um teste aprovado' }),
  ).toBeVisible()
  await expect(card('in_qa')).toBeVisible()
  const passed = await page.request.post(`/api/v1/demands/${item.id}/tests`, {
    headers,
    data: { title: 'Validação aprovada após correção', result: 'passed', author: 'Nome ignorado' },
  })
  expect(passed.status()).toBe(201)
  await drop('in_qa', 'done')
  await expect(card('done')).toBeVisible()
  await expect(page.getByTestId('board-done').getByRole('button', { name: /^Mover / })).toHaveCount(
    0,
  )
  const detail: Detail = await (await page.request.get(`/api/v1/demands/${item.id}`)).json()
  expect(detail.demand.status).toBe('done')
  expect(detail.demand.return_count).toBe(1)
  expect(detail.demand.qa_cycle).toBe(2)
  expect(detail.transitions).toHaveLength(6)
  expect(detail.transitions.every((transition) => transition.actor === actorLabel)).toBe(true)
  expect(
    detail.transitions.find((transition) => transition.to_status === 'changes_requested')?.reason,
  ).toBe('Corrigir validação antes de aprovar')
  expect(detail.tests[0].author).toBe(actorLabel)
  await page.reload()
  await expect(card('done')).toBeVisible()
  expect(fullLists).toBe(0)
})
