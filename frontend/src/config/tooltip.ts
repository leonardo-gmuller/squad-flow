import type { TooltipOptions } from 'primevue/tooltip'

export function tooltip(value: string | undefined, disabled = false): TooltipOptions {
  return {
    value,
    disabled: disabled || !value,
    class: 'workspace-tooltip',
    showDelay: 240,
    hideDelay: 80,
    escape: true,
    autoHide: false,
  }
}
