import { catalogResponses } from './catalog-fixtures'
import { test, expect, type Locator, type Page } from '@playwright/test'
import type { Demand, Status } from '../../src/types'
import { stages } from '../../src/config/catalog'

test.use({ viewport: { width: 1920, height: 1080 }, actionTimeout: 10_000 })

async function workspace(page: Page, status: Status = 'backlog', reject = false) {
  let item = {
    id: 'drag-demand',
    number: 1,
    title: 'Entrega arrastável',
    description: 'Movimentação pelo quadro',
    kind: 'improvement',
    priority: 'medium',
    status,
    assignee: 'Pessoa de teste',
    reviewer: 'QA',
    reference: '',
    event_codes: [],
    package_id: '',
    release_id: '',
    release_name: '',
    product: '',
    acceptance_criteria: '',
    due_date: null,
    qa_cycle: status === 'in_qa' ? 1 : 0,
    return_count: 0,
    created_at: '2026-10-16T12:00:00Z',
    updated_at: '2026-10-16T12:00:00Z',
    completed_at: null,
  } as Demand
  const moves: { status: Status; actor: string; reason: string }[] = []
  const errors: string[] = [],
    boardQueries: URLSearchParams[] = []
  let fullListRequests = 0
  page.on('pageerror', (error) => errors.push(error.message))
  await page.routeWebSocket('**/api/v1/ws', (socket) => socket.close())
  await page.route('**/api/v1/**', async (route) => {
    const request = route.request(),
      url = new URL(request.url()),
      path = url.pathname
    if (path === '/api/v1/demands/drag-demand/move') {
      const input = request.postDataJSON()
      moves.push(input)
      if (reject)
        return route.fulfill({
          status: 409,
          json: {
            error:
              input.status === 'done'
                ? 'QA requires at least one passing test and no failures in current cycle'
                : 'Não foi possível salvar a movimentação.',
          },
        })
      item = {
        ...item,
        status: input.status,
        return_count: item.return_count + Number(input.status === 'changes_requested'),
      }
      return route.fulfill({ json: item })
    }
    if (path === '/api/v1/board') {
      boardQueries.push(url.searchParams)
      return route.fulfill({
        json: {
          owners: ['Pessoa de teste'],
          columns: stages.map((stage) => ({
            status: stage.value,
            total: Number(stage.value === item.status),
            items: stage.value === item.status ? [item] : [],
          })),
        },
      })
    }
    if (path === '/api/v1/demands' && request.method() === 'GET') fullListRequests++
    const responses: Record<string, unknown> = {
      ...catalogResponses,
      '/api/v1/auth/me': { user: { id: 'user', name: 'Pessoa autenticada', role: 'member' } },
      '/api/v1/demands/drag-demand': { demand: item, tests: [], transitions: [] },
      '/api/v1/metrics': {
        total: 1,
        active: 1,
        delivered: 0,
        by_kind: {},
        by_status: {},
        by_priority: {},
        daily_deliveries: [],
      },
    }
    await route.fulfill({ json: responses[path] ?? [] })
  })
  await page.goto('/board?q=Entrega&archive=active')
  await expect(card(page, status)).toBeVisible()
  return {
    moves,
    errors,
    boardQueries,
    get fullListRequests() {
      return fullListRequests
    },
  }
}
function card(page: Page, status: Status) {
  return page.getByTestId(`board-${status}`).locator('.work-card')
}
async function startDrag(page: Page, source: Locator, target: Locator) {
  await source.scrollIntoViewIfNeeded()
  const from = (await source.boundingBox())!,
    to = (await target.boundingBox())!
  await page.mouse.move(from.x + from.width / 2, from.y + 60)
  await page.mouse.down()
  await page.mouse.move(to.x + to.width / 2, to.y + 100, { steps: 12 })
  await expect(page.locator('.kanban-drag-preview')).toBeVisible()
}
async function drop(page: Page, source: Status, target: Status) {
  await startDrag(page, card(page, source), page.getByTestId(`board-${target}`))
  await page.mouse.up()
}

test('arrasta entre etapas, destaca destinos e mantém filtros e persistência do servidor', async ({
  page,
}) => {
  const state = await workspace(page)
  await expect(page.getByRole('button', { name: /^Mover .*Entrega arrastável$/ })).toHaveCSS(
    'position',
    'absolute',
  )
  await startDrag(page, card(page, 'backlog'), page.getByTestId('board-in_progress'))
  await expect(page.getByTestId('board-in_progress')).toHaveClass(/is-drop-over/)
  await expect(page.getByTestId('board-done')).toHaveClass(/is-drop-blocked/)
  await page.screenshot({ path: 'test-results/board-drag-desktop.png', animations: 'disabled' })
  await page.mouse.up()
  await expect(card(page, 'in_progress')).toBeVisible()
  await expect(card(page, 'backlog')).toHaveCount(0)
  await expect(page.getByRole('dialog', { name: 'Detalhes da demanda' })).toHaveCount(0)
  expect(state.moves).toEqual([{ status: 'in_progress', actor: 'Pessoa autenticada', reason: '' }])
  expect(
    state.boardQueries.every(
      (query) => query.get('q') === 'Entrega' && query.get('limit') === '20',
    ),
  ).toBe(true)
  expect(state.fullListRequests).toBe(0)
  await page.reload()
  await expect(card(page, 'in_progress')).toBeVisible()
  expect(state.errors).toEqual([])
})

test('descarta destinos inválidos, cancelamento por Escape e drop na própria coluna', async ({
  page,
}) => {
  const state = await workspace(page)
  await drop(page, 'backlog', 'done')
  await expect(page.locator('.kanban-drag-preview')).toHaveCount(0)
  await startDrag(page, card(page, 'backlog'), page.getByTestId('board-in_progress'))
  await page.keyboard.press('Escape')
  await page.mouse.up()
  await expect(page.locator('body')).not.toHaveClass(/kanban-dragging/)
  await drop(page, 'backlog', 'backlog')
  await expect(card(page, 'backlog')).toBeVisible()
  expect(state.moves).toEqual([])
  expect(state.errors).toEqual([])
})

test('devolução exige motivo e cancelar preserva a etapa original', async ({ page }) => {
  const state = await workspace(page, 'in_qa')
  await drop(page, 'in_qa', 'changes_requested')
  const dialog = page.getByRole('dialog', { name: 'Mover entrega', exact: true })
  await expect(dialog).toBeVisible()
  expect(state.moves).toEqual([])
  await dialog.getByRole('button', { name: 'Cancelar', exact: true }).click()
  await expect(dialog).toHaveCount(0)
  await expect(page.locator('.p-dialog-mask')).toHaveCount(0)
  await expect(card(page, 'in_qa')).toBeVisible()
  await drop(page, 'in_qa', 'changes_requested')
  await dialog.getByLabel(/^Motivo da devolução/).fill('   ')
  await dialog.getByRole('button', { name: 'Mover entrega', exact: true }).click()
  await expect(dialog).toContainText('Informe o motivo da devolução.')
  expect(state.moves).toEqual([])
  await dialog.getByLabel(/^Motivo da devolução/).fill('  Corrigir validação do formulário  ')
  await dialog.getByRole('button', { name: 'Mover entrega', exact: true }).click()
  await expect(dialog).toHaveCount(0)
  await expect(card(page, 'changes_requested')).toBeVisible()
  expect(state.moves).toEqual([
    {
      status: 'changes_requested',
      actor: 'Pessoa autenticada',
      reason: 'Corrigir validação do formulário',
    },
  ])
  expect(state.errors).toEqual([])
})

test('rejeição da API mantém o card e informa a regra de aprovação de QA', async ({ page }) => {
  const state = await workspace(page, 'in_qa', true)
  await drop(page, 'in_qa', 'done')
  await expect(page.getByRole('alert')).toContainText('ao menos um teste aprovado')
  await expect(card(page, 'in_qa')).toBeVisible()
  await expect(card(page, 'done')).toHaveCount(0)
  expect(state.moves).toHaveLength(1)
  expect(state.errors).toEqual([])
})

test('teclado abre a movimentação e oferece somente a próxima etapa válida', async ({ page }) => {
  const state = await workspace(page)
  const handle = page.getByRole('button', { name: /^Mover .*Entrega arrastável$/ })
  await handle.focus()
  await page.keyboard.press('Enter')
  const dialog = page.getByRole('dialog', { name: 'Mover entrega', exact: true })
  await expect(dialog).toBeVisible()
  await dialog.getByRole('combobox', { name: /^Mover para/ }).click()
  await expect(page.getByRole('option')).toHaveCount(1)
  await page.getByRole('option', { name: 'Em desenvolvimento', exact: true }).click()
  await expect(page.getByRole('listbox')).toHaveCount(0)
  await dialog.getByRole('button', { name: 'Mover entrega', exact: true }).press('Enter')
  await expect(card(page, 'in_progress')).toBeVisible()
  expect(state.moves).toHaveLength(1)
  expect(state.errors).toEqual([])
})

test.describe('toque no celular', () => {
  test.use({ viewport: { width: 390, height: 844 }, hasTouch: true, isMobile: true })
  test('arrasta pela alça e rola horizontalmente até a coluna de destino', async ({ page }) => {
    const state = await workspace(page)
    const handle = page.getByRole('button', { name: /^Mover .*Entrega arrastável$/ })
    const bounds = (await handle.boundingBox())!
    const cdp = await page.context().newCDPSession(page)
    const touch = async (type: 'touchStart' | 'touchMove' | 'touchEnd', x = 0, y = 0) =>
      cdp.send('Input.dispatchTouchEvent', {
        type,
        touchPoints: type === 'touchEnd' ? [] : [{ x, y, id: 1 }],
      })
    const y = bounds.y + bounds.height / 2
    await touch('touchStart', bounds.x + bounds.width / 2, y)
    await touch('touchMove', 370, y)
    await expect(page.locator('.kanban-drag-preview')).toBeVisible()
    await expect
      .poll(() => page.locator('.kanban-board').evaluate((element) => element.scrollLeft))
      .toBeGreaterThan(180)
    await touch('touchMove', 190, y)
    const target = page.getByTestId('board-in_progress')
    await target.scrollIntoViewIfNeeded()
    const destination = (await target.boundingBox())!
    await touch('touchMove', destination.x + destination.width / 2, destination.y + 100)
    await expect(target).toHaveClass(/is-drop-over/)
    await page.screenshot({ path: 'test-results/board-drag-mobile.png', animations: 'disabled' })
    await touch('touchEnd')
    await expect(card(page, 'in_progress')).toBeVisible()
    expect(state.moves).toHaveLength(1)
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(390)
    expect(state.errors).toEqual([])
    await cdp.detach()
  })
})
