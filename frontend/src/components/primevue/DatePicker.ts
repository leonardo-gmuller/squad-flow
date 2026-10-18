import PrimeDatePicker from 'primevue/datepicker'
import {
  CalendarDaysIcon,
  ChevronDownIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  ChevronUpIcon,
  XMarkIcon,
} from '@heroicons/vue/24/outline'
import { icon, withIcons } from './withIcons'

export default withIcons(PrimeDatePicker, () => ({
  dropdownicon: icon(CalendarDaysIcon),
  inputicon: (scope) => icon(CalendarDaysIcon, { onClick: scope.clickCallback })(scope),
  clearicon: (scope) => icon(XMarkIcon, { onClick: scope.clearCallback })(scope),
  previcon: icon(ChevronLeftIcon),
  nexticon: icon(ChevronRightIcon),
  incrementicon: icon(ChevronUpIcon),
  decrementicon: icon(ChevronDownIcon),
}))
