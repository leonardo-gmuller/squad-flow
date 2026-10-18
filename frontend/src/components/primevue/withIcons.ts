import { defineComponent, h, type Component, type Slots } from 'vue'
import {
  ArrowPathIcon,
  CheckCircleIcon,
  ExclamationTriangleIcon,
  InformationCircleIcon,
  XCircleIcon,
  XMarkIcon,
} from '@heroicons/vue/24/outline'

type Scope = Record<string, unknown>
export const icon =
  (component: Component, extra: Scope = {}) =>
  (scope: Scope = {}) => [h(component, { ...extra, class: ['app-icon', scope.class, extra.class] })]
export const loading = icon(ArrowPathIcon, { class: 'app-icon-spin' })
export const close = icon(XMarkIcon)
export const severityIcon = (severity: unknown) =>
  severity === 'success'
    ? CheckCircleIcon
    : severity === 'error'
      ? XCircleIcon
      : severity === 'warn'
        ? ExclamationTriangleIcon
        : InformationCircleIcon

// Keep PrimeVue behavior, props, events and caller slots; only supply defaults
// through its public icon slots. Native Column components retain their registry.
export function withIcons<T extends Component>(component: T, defaults: (attrs: Scope) => Slots): T {
  return defineComponent({
    name: `Hero${(component as { name?: string }).name || 'Component'}`,
    inheritAttrs: false,
    setup(_, { attrs, slots }) {
      return () => h(component, attrs, { ...defaults(attrs), ...slots })
    },
  }) as T
}
