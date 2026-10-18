import { test, expect } from './fixtures'
import { readFile } from 'node:fs/promises'

const reportID = '77777777-7777-4777-8777-777777777701'
const html = `<!doctype html><html lang="pt-BR"><head><meta charset="utf-8"><meta http-equiv="Content-Security-Policy" content="default-src 'none'; script-src 'none'; style-src 'unsafe-inline'; img-src data:; base-uri 'none'; form-action 'none'"><style>body{font:16px system-ui;background:#f5f4fa;padding:24px;color:#24223a}section{background:white;border:1px solid #c7bfea;border-radius:18px;padding:24px;margin:16px 0}svg{width:100%;max-width:600px;height:auto}@media print{section{break-after:page}}</style></head><body><h1>Análise de entregas da squad</h1><section><h2>Qualidade e entregas</h2><p>Três demandas concluídas com evidências de QA.</p><svg viewBox="0 0 600 160" role="img" aria-label="Entregas por esteira"><rect x="0" y="40" width="420" height="50" fill="#8b76dc" fill-opacity="0.3" stroke="#8b76dc" stroke-width="2"/><text x="10" y="72">Melhorias · 3 entregas</text></svg></section><section><h2>Próximos passos</h2><p>Revisar os critérios de aceite antes de iniciar QA.</p></section></body></html>`
const content = {
  title: 'Análise de entregas da squad',
  summary: 'Três entregas concluídas com QA.',
  html,
  metrics: [],
  sections: [],
  charts: [],
  tables: [],
  recommendations: [],
  missing_data: [],
  source_ids: [],
}
const report = {
  id: reportID,
  demand_id: '',
  prompt: 'Analise as entregas',
  created_at: '2026-10-17T12:00:00Z',
  content,
}
function job(status: 'queued' | 'processing' | 'completed' | 'failed', id = reportID) {
  return {
    id,
    request: { prompt: 'Analise as entregas' },
    status,
    attempts: status === 'queued' ? 0 : status === 'failed' ? 3 : 1,
    max_attempts: 3,
    available_at: '2026-10-17T12:00:00Z',
    progress_agent: status === 'processing' ? 'análise' : 'fila',
    progress_message:
      status === 'processing'
        ? 'Conferindo indicadores e evidências'
        : status === 'failed'
          ? 'Não foi possível concluir o relatório'
          : 'Aguardando processamento',
    error_message: status === 'failed' ? 'O provedor não concluiu as três tentativas.' : '',
    report_id: status === 'completed' ? reportID : '',
    created_at: '2026-10-17T12:00:00Z',
    updated_at: '2026-10-17T12:00:01Z',
    completed_at: status === 'completed' ? '2026-10-17T12:00:02Z' : null,
  }
}

test('fila persiste ao navegar e relatório HTML abre, apresenta e exporta', async ({
  page,
}, testInfo) => {
  let state: 'none' | 'queued' | 'processing' | 'completed' = 'none',
    submissions = 0,
    externalRequests = 0
  await page.route('**/api/v1/reports/status', (route) =>
    route.fulfill({ json: { enabled: true, asynchronous: true, format: 'html' } }),
  )
  await page.route('**/api/v1/reports/jobs', (route) =>
    route.fulfill({ json: state === 'none' ? [] : [job(state)] }),
  )
  await page.route('**/api/v1/reports', (route) =>
    route.fulfill({
      json: state === 'completed' ? [{ ...report, content: { ...content, html: undefined } }] : [],
    }),
  )
  await page.route(`**/api/v1/reports/${reportID}`, (route) => route.fulfill({ json: report }))
  await page.route(`**/api/v1/reports/${reportID}/html?download=1`, (route) =>
    route.fulfill({ contentType: 'text/html', body: html }),
  )
  await page.route('https://report-exfiltration.invalid/**', (route) => {
    externalRequests++
    return route.abort()
  })
  await page.route('**/api/v1/reports/generate', async (route) => {
    submissions++
    expect(route.request().headers()['idempotency-key']).toMatch(/^[0-9a-f-]{36}$/)
    expect(route.request().postDataJSON().prompt).toBe('Analise as entregas')
    state = 'queued'
    await route.fulfill({ status: 202, json: job('queued') })
  })
  await page.goto('/reports?tab=ai')
  await page.getByLabel('Seu pedido').fill('Analise as entregas')
  await page.getByRole('button', { name: 'Gerar relatório', exact: true }).click()
  await expect(page.getByTestId(`report-job-${reportID}`)).toContainText('Na fila')
  await expect(page.getByText('Você pode sair desta tela; o preparo continua.')).toBeVisible()
  expect(submissions).toBe(1)
  await page.goto('/board')
  await page.goto('/reports?tab=ai')
  await page.reload()
  await expect(page.getByTestId(`report-job-${reportID}`)).toContainText('Na fila')
  state = 'processing'
  await expect(page.getByTestId(`report-job-${reportID}`)).toContainText(
    'Conferindo indicadores e evidências',
  )
  state = 'completed'
  const rendered = page.getByRole('article', { name: 'Relatório gerado' })
  await expect(rendered.getByTestId('report-html')).toBeVisible()
  await expect(
    page
      .frameLocator('iframe[data-testid="report-html"]')
      .getByRole('heading', { name: 'Qualidade e entregas' }),
  ).toBeVisible()
  await expect(rendered.getByTestId('report-html')).toHaveAttribute('sandbox', '')
  await rendered.screenshot({ path: testInfo.outputPath('report-renderer-desktop.png') })
  const downloadEvent = page.waitForEvent('download')
  await rendered.getByRole('button', { name: 'Baixar HTML', exact: true }).click()
  const download = await downloadEvent
  expect(download.suggestedFilename()).toBe(`relatorio-${reportID}.html`)
  const exported = testInfo.outputPath('relatorio.html')
  await download.saveAs(exported)
  const exportedHTML = await readFile(exported, 'utf8')
  expect(exportedHTML).toContain('Content-Security-Policy')
  expect(exportedHTML).toContain('@media print')
  expect(exportedHTML).toContain('<svg')
  await rendered.getByRole('button', { name: 'Apresentar', exact: true }).click()
  await expect(
    page
      .frameLocator('iframe[title^="Apresentação:"]')
      .getByRole('heading', { name: 'Análise de entregas da squad' }),
  ).toBeVisible()
  await expect(page.getByRole('dialog')).toHaveCSS('opacity', '1')
  await page.screenshot({ path: testInfo.outputPath('report-presentation.png'), fullPage: false })
  await page.getByRole('button', { name: 'Fechar apresentação' }).click()
  await expect(page.getByRole('dialog')).toHaveCount(0)
  await page.setViewportSize({ width: 390, height: 844 })
  await expect(
    page
      .frameLocator('iframe[data-testid="report-html"]')
      .getByRole('heading', { name: 'Qualidade e entregas' }),
  ).toBeVisible()
  await rendered.getByTestId('report-html').scrollIntoViewIfNeeded()
  await page.evaluate(
    () =>
      new Promise<void>((resolve) =>
        requestAnimationFrame(() => requestAnimationFrame(() => resolve())),
      ),
  )
  await page.screenshot({ path: testInfo.outputPath('report-html-mobile.png'), fullPage: false })
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(391)
  await page.goto('/dashboard')
  await page.getByRole('button', { name: 'Abrir relatório no dashboard' }).click()
  await expect(page.getByRole('dialog').getByTestId('report-html')).toBeVisible()
  // Even a malformed server document cannot execute scripts in the application origin.
  await page.locator('iframe[data-testid="report-html"]').evaluate((frame) => {
    ;(frame as HTMLIFrameElement).srcdoc +=
      '<script>parent.document.body.dataset.reportExecuted="yes"</script><img src="https://report-exfiltration.invalid/tracker">'
  })
  expect(await page.locator('body').getAttribute('data-report-executed')).toBeNull()
  expect(externalRequests).toBe(0)
  expect(submissions).toBe(1)
})

test('histórico de falha permite novo pedido preservando a solicitação anterior', async ({
  page,
}) => {
  const failed = job('failed'),
    retryID = '77777777-7777-4777-8777-777777777702'
  let retried = false
  await page.route('**/api/v1/reports/status', (route) =>
    route.fulfill({ json: { enabled: true } }),
  )
  await page.route('**/api/v1/reports', (route) => route.fulfill({ json: [] }))
  await page.route('**/api/v1/reports/jobs', (route) =>
    route.fulfill({ json: retried ? [job('queued', retryID), failed] : [failed] }),
  )
  await page.route(`**/api/v1/reports/jobs/${reportID}`, (route) =>
    route.fulfill({
      json: {
        ...failed,
        events: [
          {
            id: 'event-1',
            job_id: reportID,
            type: 'report.failed',
            agent: 'falha',
            message: 'Três tentativas encerradas',
            attempt: 3,
            created_at: failed.created_at,
          },
        ],
      },
    }),
  )
  await page.route(`**/api/v1/reports/jobs/${reportID}/retry`, async (route) => {
    retried = true
    await route.fulfill({ status: 202, json: job('queued', retryID) })
  })
  await page.goto('/reports?tab=ai')
  const original = page.getByTestId(`report-job-${reportID}`)
  await expect(original).toContainText('Tentativa 3 de 3')
  await original.getByRole('button', { name: 'Ver etapas' }).click()
  await expect(original).toContainText('Três tentativas encerradas')
  await original.getByRole('button', { name: 'Tentar novamente' }).click()
  await expect(page.getByTestId(`report-job-${retryID}`)).toContainText('Na fila')
  await expect(original).toContainText('Falha')
  await expect(original.getByRole('button', { name: 'Tentar novamente' })).toBeDisabled()
  await page.reload()
  await expect(page.getByTestId(`report-job-${retryID}`)).toContainText('Na fila')
})

test('nova conclusão mantém a apresentação e o download vinculados ao relatório aberto', async ({
  page,
}) => {
  const nextID = '77777777-7777-4777-8777-777777777703'
  const nextReport = {
    ...report,
    id: nextID,
    content: {
      ...content,
      title: 'Nova análise da squad',
      html: html.replace('Análise de entregas da squad', 'Nova análise da squad'),
    },
  }
  let completed = false
  await page.route('**/api/v1/reports/status', (route) =>
    route.fulfill({ json: { enabled: true } }),
  )
  await page.route('**/api/v1/reports', (route) =>
    route.fulfill({ json: completed ? [nextReport, report] : [report] }),
  )
  await page.route('**/api/v1/reports/jobs', (route) =>
    route.fulfill({
      json: [
        {
          ...job(completed ? 'completed' : 'processing', nextID),
          report_id: completed ? nextID : '',
        },
      ],
    }),
  )
  await page.route(`**/api/v1/reports/${reportID}`, (route) => route.fulfill({ json: report }))
  await page.route(`**/api/v1/reports/${nextID}`, (route) => route.fulfill({ json: nextReport }))
  await page.route(`**/api/v1/reports/${reportID}/html?download=1`, (route) =>
    route.fulfill({ contentType: 'text/html', body: html }),
  )
  await page.goto(`/reports?report_id=${reportID}`)
  await page
    .getByRole('article', { name: 'Relatório gerado' })
    .getByRole('button', { name: 'Apresentar', exact: true })
    .click()
  const dialog = page.getByRole('dialog')
  await expect(dialog).toHaveCSS('opacity', '1')
  completed = true
  await expect(
    page
      .getByRole('article', { name: 'Relatório gerado' })
      .getByRole('heading', { name: 'Nova análise da squad', exact: true }),
  ).toBeAttached()
  await expect(dialog).toHaveAccessibleName(content.title)
  await expect(
    page
      .frameLocator('iframe[title^="Apresentação:"]')
      .getByRole('heading', { name: content.title, exact: true }),
  ).toBeVisible()
  const downloadEvent = page.waitForEvent('download')
  await dialog.getByRole('button', { name: 'Baixar HTML', exact: true }).click()
  expect((await downloadEvent).suggestedFilename()).toBe(`relatorio-${reportID}.html`)
})
