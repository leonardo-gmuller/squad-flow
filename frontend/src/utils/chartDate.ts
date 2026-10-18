import moment from '../config/moment'

const englishDates = [
  'MMM D, YYYY',
  'MMMM D, YYYY',
  'D MMM YYYY',
  'D MMMM YYYY',
  'MMM D YYYY',
  'MMMM D YYYY',
  'ddd, MMM D, YYYY',
  'dddd, MMMM D, YYYY',
]

// Only complete dates and exact month/weekday names are localized. Category names stay intact.
// Chart dates use UTC because daily metric buckets are calculated in UTC by the API.
export function chartDateLabel(value: string, detailed = false): string {
  const label = value.trim()
  const date = moment.utc(label, [moment.ISO_8601, ...englishDates], 'en', true)
  if (date.isValid() && date.parsingFlags().parsedDateParts.length >= 3) {
    return date.locale('pt-br').format(detailed ? 'LL' : 'DD [de] MMM')
  }

  const month = moment.utc(label, ['YYYY-MM', 'MMM YYYY', 'MMMM YYYY'], 'en', true)
  if (month.isValid())
    return month.locale('pt-br').format(detailed ? 'MMMM [de] YYYY' : 'MMM [de] YYYY')

  const english = moment.localeData('en')
  const portuguese = moment.localeData('pt-br')
  const name = label.toLowerCase()
  const monthIndex = english
    .months()
    .findIndex((item, index) =>
      [item.toLowerCase(), english.monthsShort()[index].toLowerCase()].includes(name),
    )
  if (monthIndex !== -1)
    return (detailed ? portuguese.months() : portuguese.monthsShort())[monthIndex]
  const dayIndex = english
    .weekdays()
    .findIndex((item, index) =>
      [item.toLowerCase(), english.weekdaysShort()[index].toLowerCase()].includes(name),
    )
  if (dayIndex !== -1)
    return (detailed ? portuguese.weekdays() : portuguese.weekdaysShort())[dayIndex]

  return value
}
