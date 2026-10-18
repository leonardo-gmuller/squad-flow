import { color } from 'chart.js/helpers'

export function chartFill(value: string, opacity = 0.22): string {
  return color(value).alpha(opacity).rgbString()
}
