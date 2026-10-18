import PrimeToast from 'primevue/toast'
import { h } from 'vue'
import { close, severityIcon, withIcons } from './withIcons'

export default withIcons(PrimeToast, () => ({
  closeicon: close,
  message: ({ message }) => {
    const value = message as { severity?: string; summary?: string; detail?: string }
    return [
      h(severityIcon(value.severity), { class: 'app-icon p-toast-message-icon' }),
      h('div', { class: 'p-toast-message-text' }, [
        h('span', { class: 'p-toast-summary' }, value.summary),
        value.detail ? h('div', { class: 'p-toast-detail' }, value.detail) : null,
      ]),
    ]
  },
}))
