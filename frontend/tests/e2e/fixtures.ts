import { test as base, expect, type Cookie } from '@playwright/test'
// Reuse the worker's admin session; authenticate again if a logout revoked it.
let adminCookies: Cookie[] = []
export const test = base.extend({
  page: async ({ page, context }, use) => {
    const errors: string[] = []
    page.on('pageerror', (error) => errors.push(error.message))
    const email = process.env.E2E_EMAIL,
      password = process.env.E2E_PASSWORD
    if (!email || !password)
      throw new Error('Configure E2E_EMAIL e E2E_PASSWORD para um usuário de teste.')
    if (adminCookies.length) await context.addCookies(adminCookies)
    const current = adminCookies.length ? await page.request.get('/api/v1/auth/me') : undefined
    if (!current?.ok()) {
      const response = await page.request.post('/api/v1/auth/login', {
        headers: { 'X-SquadFlow-Client': 'web' },
        data: { email, password },
      })
      expect(response.status()).toBe(200)
      adminCookies = await context.cookies()
    }
    await use(page)
    expect(errors).toEqual([])
  },
})
export { expect }
