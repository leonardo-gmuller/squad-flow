import type { Workstream } from '../../src/types'

// API responses for tests that intentionally mock the entire workspace.
const defaults = [
  { code: 'improvement', name: 'Melhorias', color: '#8976e8' },
  { code: 'corrective', name: 'Corretivas', color: '#e77c99' },
  { code: 'testing', name: 'Testes', color: '#57bad3' },
  { code: 'rewrite', name: 'Reescritas', color: '#e9ad67' },
]
export const catalogResponses: Record<string, unknown> = {
  '/api/v1/workstreams': defaults.map(
    (item, index) =>
      ({
        ...item,
        id: `66666666-6666-4666-8666-66666666660${index + 1}`,
        active: true,
        created_at: '2026-10-16T12:00:00Z',
        updated_at: '2026-10-16T12:00:00Z',
      }) satisfies Workstream,
  ),
  '/api/v1/members': [{ id: 'test-user', name: 'Pessoa de teste', email: 'test@example.com' }],
}
