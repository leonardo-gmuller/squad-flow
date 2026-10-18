import { test, expect } from '@playwright/test'
import { chartDateLabel } from '../../src/utils/chartDate'
import { dateLabel, dateTime, isLate, reportCellLabel } from '../../src/utils/format'
import { calendarDate, calendarDateString } from '../../src/utils/calendar'

test('Moment formata datas e horários em português e rejeita datas inválidas', () => {
  expect(dateLabel('2026-10-16')).toBe('16 de out de 2026')
  expect(dateTime('2026-10-16T09:30:00')).toBe('16/10/2026 09:30')
  expect(dateLabel('2026-02-30')).toBe('—')
  expect(dateTime('inválida')).toBe('—')
  expect(dateLabel(null)).toBe('—')
  expect(isLate('inválida')).toBe(false)
  expect(isLate('2000-01-01')).toBe(true)
  expect(isLate('2000-01-01', true)).toBe(false)
  expect(reportCellLabel('2026-10-16T09:30:00')).toBe('16/10/2026 09:30')
  expect(reportCellLabel('October 16, 2026')).toBe('16 de outubro de 2026')
  expect(reportCellLabel('OS-10')).toBe('OS-10')
})

test('rótulos e tooltips localizam datas ISO e inglesas sem alterar categorias', () => {
  expect(chartDateLabel('2026-09-01')).toBe('01 de set')
  expect(chartDateLabel('2026-09-01', true)).toBe('1 de setembro de 2026')
  expect(chartDateLabel('2026-01-01T00:00:00Z', true)).toBe('1 de janeiro de 2026')
  expect(chartDateLabel('October 16, 2026')).toBe('16 de out')
  expect(chartDateLabel('Oct 16, 2026', true)).toBe('16 de outubro de 2026')
  expect(chartDateLabel('November 2026')).toBe('nov de 2026')
  expect(chartDateLabel('2026-11', true)).toBe('novembro de 2026')
  expect(chartDateLabel('February', true)).toBe('fevereiro')
  expect(chartDateLabel('Wed', true)).toBe('quarta-feira')
  for (const label of ['Melhorias', 'OS-10', 'Squad', '2026', '2026-02-30', '16/10/2026']) {
    expect(chartDateLabel(label)).toBe(label)
  }
})

test('calendários preservam o dia civil e o formato ISO enviado à API', () => {
  const date = calendarDate('2026-10-16T12:00:00Z')!
  expect(calendarDateString(date)).toBe('2026-10-16')
  expect(date.getDate()).toBe(16)
  expect(date.getMonth()).toBe(9)
  expect(calendarDate('2026-02-30')).toBeNull()
  expect(calendarDate('')).toBeNull()
  expect(calendarDateString(new Date(NaN))).toBe('')
})
