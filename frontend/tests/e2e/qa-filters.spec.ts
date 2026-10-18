import { test, expect, type Page } from '@playwright/test'
import type { Demand } from '../../src/types'
import { catalogResponses } from './catalog-fixtures'

const people = [
  { id: 'qa-one', name: 'Marina', email: 'marina@example.test' },
  { id: 'qa-two', name: 'Marina', email: 'marina2@example.test' },
  { id: 'dev-one', name: 'Alex', email: 'alex@example.test' },
  { id: 'dev-two', name: 'Daniel', email: 'daniel@example.test' },
]
function demand(number: number, input: Partial<Demand>): Demand {
  return {
    id: `qa-${number}`,
    number,
    title: '',
    description: 'Conferir regras e evidências da entrega.',
    kind: 'improvement',
    priority: 'medium',
    status: 'in_qa',
    assignee_id: 'dev-one',
    assignee: 'Alex',
    reviewer_id: 'qa-one',
    reviewer: 'Marina',
    reference: 'OS-120',
    event_codes: ['S-2200'],
    package_id: '',
    release_id: 'release-active',
    release_name: '2.0.0',
    product: 'Folha',
    acceptance_criteria: '',
    due_date: null,
    qa_cycle: 1,
    return_count: 0,
    created_at: '2026-10-17T12:00:00Z',
    updated_at: '2026-10-17T12:00:00Z',
    completed_at: null,
    ...input,
  }
}
const items = [
  demand(1, { title: 'Validar cálculo de férias' }),
  demand(2, {
    title: 'Corrigir integração do evento',
    kind: 'corrective',
    reviewer_id: 'qa-two',
    release_id: 'release-archived',
    release_name: '1.0.0',
  }),
  demand(3, {
    title: 'Revisar retorno da API',
    status: 'changes_requested',
    assignee_id: 'dev-two',
    assignee: 'Daniel',
    return_count: 1,
  }),
  demand(4, {
    title: 'Concluir validação do XML',
    status: 'done',
    completed_at: '2026-10-17T15:00:00Z',
  }),
  demand(5, {
    title: 'Entrega sem responsáveis',
    kind: 'testing',
    assignee_id: '',
    assignee: '',
    reviewer_id: '',
    reviewer: '',
    release_id: '',
    release_name: '',
  }),
  demand(6, { title: 'Ainda não entrou no QA', status: 'backlog' }),
]
async function mockWorkspace(page: Page) {
  const errors: string[] = []
  page.on('pageerror', (error) => errors.push(error.message))
  await page.routeWebSocket('**/api/v1/ws', (ws) => ws.close())
  await page.route('**/api/v1/**', (route) => {
    const path = new URL(route.request().url()).pathname
    const responses: Record<string, unknown> = {
      ...catalogResponses,
      '/api/v1/auth/me': { user: { ...people[0], role: 'member' } },
      '/api/v1/members': people,
      '/api/v1/demands': items,
      '/api/v1/demands/qa-1': { demand: items[0], tests: [], transitions: [] },
      '/api/v1/metrics': { in_qa: 3 },
      '/api/v1/notifications/unread-count': { count: 0 },
      '/api/v1/releases': [
        { id: 'release-active', name: '2.0.0', archived_at: null },
        { id: 'release-archived', name: '1.0.0', archived_at: '2026-10-16T12:00:00Z' },
      ],
    }
    return route.fulfill({ json: responses[path] ?? [] })
  })
  await page.goto('/qa')
  await expect(page.getByRole('tab', { name: 'Aguardando QA 3', exact: true })).toBeVisible()
  return errors
}
async function choose(page: Page, field: string, value: string) {
  await page.getByRole('combobox', { name: field, exact: true }).click()
  await page.getByRole('option', { name: value, exact: true }).click()
  await expect(page.getByRole('listbox')).toHaveCount(0)
}
const panel = (page: Page) => page.getByRole('tabpanel')

test('combina busca e filtros, conta cada etapa e preserva os filtros ao abrir a demanda', async ({
  page,
}) => {
  const errors = await mockWorkspace(page)
  await expect(panel(page).locator('.work-card')).toHaveCount(3)
  await expect(page.getByRole('combobox', { name: 'Esteira', exact: true })).toContainText(
    'Todas as esteiras',
  )
  await choose(page, 'Esteira', 'Melhorias')
  await page.getByRole('button', { name: 'Minhas revisões', exact: true }).click()
  await expect(page.getByRole('button', { name: 'Minhas revisões', exact: true })).toHaveAttribute(
    'aria-pressed',
    'true',
  )
  await expect(panel(page).locator('.work-card')).toHaveCount(1)
  await expect(page.getByRole('tab', { name: 'Devolvidas 1', exact: true })).toBeVisible()
  await expect(page.getByRole('tab', { name: 'Concluídas 1', exact: true })).toBeVisible()
  const search = page.getByRole('textbox', { name: 'Buscar demandas de QA', exact: true })
  await search.fill('CALCULO FERIAS')
  await expect(panel(page)).toContainText('Validar cálculo de férias')
  await expect(page.getByRole('tab', { name: 'Devolvidas 0', exact: true })).toBeVisible()
  await search.fill('SF-0001')
  await expect(panel(page).locator('.work-card')).toHaveCount(1)
  await page
    .getByRole('button', { name: 'Abrir SF-0001 Validar cálculo de férias', exact: true })
    .click()
  const drawer = page.getByRole('dialog', { name: 'Detalhes da demanda', exact: true })
  await expect(drawer).toContainText('Validar cálculo de férias')
  await page.getByRole('button', { name: 'Fechar detalhes', exact: true }).click()
  await expect(search).toHaveValue('SF-0001')
  await page.getByRole('button', { name: 'Remover filtro Busca: SF-0001', exact: true }).click()
  await choose(page, 'Desenvolvedor', 'Alex')
  await choose(page, 'Versão', '2.0.0')
  await page.getByRole('tab', { name: 'Devolvidas 0', exact: true }).click()
  await expect(
    panel(page).getByRole('heading', {
      name: 'Nenhuma demanda corresponde aos filtros',
      exact: true,
    }),
  ).toBeVisible()
  await page
    .getByRole('button', { name: 'Remover filtro Desenvolvedor: Alex', exact: true })
    .click()
  await expect(panel(page)).toContainText('Revisar retorno da API')
  await page.getByRole('tab', { name: 'Concluídas 1', exact: true }).click()
  await expect(panel(page)).toContainText('Concluir validação do XML')
  await page.screenshot({
    path: 'test-results/qa-filters-desktop-dark.png',
    fullPage: true,
    animations: 'disabled',
  })
  await page
    .getByRole('region', { name: 'Filtrar demandas', exact: false })
    .getByRole('button', { name: 'Limpar filtros', exact: true })
    .click()
  await expect(page.getByRole('tab', { name: 'Concluídas 1', exact: true })).toHaveAttribute(
    'aria-selected',
    'true',
  )
  await expect(page.getByRole('button', { name: 'Minhas revisões', exact: true })).toHaveAttribute(
    'aria-pressed',
    'false',
  )
  expect(errors).toEqual([])
})

test('filtra por IDs mesmo com nomes iguais e consulta arquivadas ou itens sem atribuição no celular', async ({
  page,
}) => {
  await page.setViewportSize({ width: 390, height: 844 })
  const errors = await mockWorkspace(page)
  const controls = page.getByRole('region', { name: 'Filtrar demandas', exact: false })
  await page.getByRole('button', { name: 'Minhas revisões', exact: true }).click()
  await expect(panel(page).locator('.work-card')).toHaveCount(1)
  await page.getByRole('button', { name: 'Minhas revisões', exact: true }).click()
  await expect(panel(page).locator('.work-card')).toHaveCount(3)
  await choose(page, 'Responsável QA', 'Marina · marina2@example.test')
  await expect(panel(page)).toContainText('Corrigir integração do evento')
  await expect(panel(page).locator('.work-card')).toHaveCount(1)
  await choose(page, 'Versão', '1.0.0 · Arquivada')
  await expect(panel(page).locator('.work-card')).toHaveCount(1)
  await controls.getByRole('button', { name: 'Limpar filtros', exact: true }).click()
  await choose(page, 'Responsável QA', 'Sem responsável QA')
  await choose(page, 'Desenvolvedor', 'Sem desenvolvedor')
  await choose(page, 'Versão', 'Sem versão')
  await expect(panel(page)).toContainText('Entrega sem responsáveis')
  await expect(panel(page).locator('.work-card')).toHaveCount(1)
  await expect
    .poll(() => page.evaluate(() => document.documentElement.scrollWidth))
    .toBeLessThanOrEqual(390)
  await page.screenshot({
    path: 'test-results/qa-filters-mobile-dark.png',
    fullPage: true,
    animations: 'disabled',
  })
  await page.getByRole('textbox', { name: 'Buscar demandas de QA', exact: true }).fill('nao existe')
  await expect(
    panel(page).getByRole('heading', {
      name: 'Nenhuma demanda corresponde aos filtros',
      exact: true,
    }),
  ).toBeVisible()
  await panel(page).getByRole('button', { name: 'Limpar filtros', exact: true }).click()
  await expect(panel(page).locator('.work-card')).toHaveCount(3)
  await expect(controls.getByRole('button', { name: 'Limpar filtros', exact: true })).toBeDisabled()
  await page.getByRole('button', { name: 'Tema claro', exact: true }).click()
  await page.screenshot({
    path: 'test-results/qa-filters-mobile-light.png',
    fullPage: true,
    animations: 'disabled',
  })
  expect(errors).toEqual([])
})
