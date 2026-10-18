import PrimeDataTable from 'primevue/datatable'
import {
  ChevronDoubleLeftIcon,
  ChevronDoubleRightIcon,
  ChevronDownIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
} from '@heroicons/vue/24/outline'
import { icon, loading, withIcons } from './withIcons'

export default withIcons(PrimeDataTable, () => ({
  loadingicon: loading,
  paginatorfirstpagelinkicon: icon(ChevronDoubleLeftIcon),
  paginatorprevpagelinkicon: icon(ChevronLeftIcon),
  paginatornextpagelinkicon: icon(ChevronRightIcon),
  paginatorlastpagelinkicon: icon(ChevronDoubleRightIcon),
  paginatorrowsperpagedropdownicon: icon(ChevronDownIcon),
  paginatorjumptopagedropdownicon: icon(ChevronDownIcon),
}))
