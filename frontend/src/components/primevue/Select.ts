import PrimeSelect from 'primevue/select'
import { ChevronDownIcon, MagnifyingGlassIcon, XMarkIcon } from '@heroicons/vue/24/outline'
import { icon, loading, withIcons } from './withIcons'

export default withIcons(PrimeSelect, () => ({
  dropdownicon: icon(ChevronDownIcon),
  filtericon: icon(MagnifyingGlassIcon),
  loadingicon: loading,
  clearicon: (scope) => icon(XMarkIcon, { onClick: scope.clearCallback })(scope),
}))
