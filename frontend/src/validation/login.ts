import { z } from 'zod'

export const loginSchema = z.object({
  email: z
    .string()
    .trim()
    .min(1, 'Informe seu email.')
    .max(254, 'O email excede o tamanho permitido.')
    .pipe(z.email({ pattern: z.regexes.html5Email, message: 'Informe um email válido.' })),
  password: z
    .string()
    .min(1, 'Informe sua senha.')
    .refine(
      (value) => new TextEncoder().encode(value).length <= 72,
      'A senha excede o tamanho permitido.',
    ),
})

export type LoginValues = z.infer<typeof loginSchema>
