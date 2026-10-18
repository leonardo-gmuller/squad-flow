import PrimeCheckbox from 'primevue/checkbox'
import { CheckIcon, MinusIcon } from '@heroicons/vue/24/outline'
import { icon, withIcons } from './withIcons'

export default withIcons(PrimeCheckbox, () => ({
  icon: (scope) =>
    scope.indeterminate ? icon(MinusIcon)(scope) : scope.checked ? icon(CheckIcon)(scope) : [],
}))
