import moment from '../config/moment'

// Calendar fields represent a day, independent of the browser's timezone.
export function calendarDate(value?: string): Date | null {
  if (!value) return null
  const date = moment(value.slice(0, 10), 'YYYY-MM-DD', true)
  return date.isValid() ? date.toDate() : null
}

export function calendarDateString(date: Date): string {
  const value = moment(date)
  return value.isValid() ? value.format('YYYY-MM-DD') : ''
}
