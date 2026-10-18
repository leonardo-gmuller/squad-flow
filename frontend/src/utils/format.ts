import moment from '../config/moment'
import { chartDateLabel } from './chartDate'

export function dateLabel(value?: string | null) {
  if (!value) return '—'
  const date = moment(value, moment.ISO_8601, true)
  return date.isValid() ? date.format('DD [de] MMM [de] YYYY') : '—'
}
export function dateTime(value?: string | null) {
  if (!value) return '—'
  const date = moment(value, moment.ISO_8601, true)
  return date.isValid() ? date.format('L LT') : '—'
}
export function dateTimestamp(value: string) {
  const date = moment(value, moment.ISO_8601, true)
  return date.isValid() ? date.valueOf() : 0
}
export function reportCellLabel(value: string) {
  const date = moment(value, moment.ISO_8601, true)
  if (date.isValid() && /[T ]\d{2}:\d{2}/.test(value)) return date.format('L LT')
  return chartDateLabel(value, true)
}
export function initials(value?: string) {
  if (!value?.trim()) return '—'
  const parts = value.trim().split(/\s+/)
  return (
    parts.length > 1 ? parts[0][0] + parts[parts.length - 1][0] : parts[0].slice(0, 2)
  ).toUpperCase()
}
export function demandCode(number: number) {
  return `SF-${String(number).padStart(4, '0')}`
}
export function percent(n: number, total: number) {
  return total ? Math.round((n / total) * 100) : 0
}
export function isLate(date?: string | null, done = false) {
  if (!date || done) return false
  return moment.utc(date, moment.ISO_8601, true).isBefore(moment())
}
