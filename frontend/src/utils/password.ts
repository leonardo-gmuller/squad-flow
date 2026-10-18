export function passwordError(password: string): string {
  if (Array.from(password).length < 12) return 'Use pelo menos 12 caracteres na senha.'
  if (new TextEncoder().encode(password).length > 72)
    return 'Essa senha ficou muito longa. Reduza a quantidade de caracteres.'
  return ''
}

export function resetToken(value: string): string {
  const text = value.trim()
  try {
    return new URLSearchParams(new URL(text).hash.slice(1)).get('token') || ''
  } catch {
    return text
  }
}
