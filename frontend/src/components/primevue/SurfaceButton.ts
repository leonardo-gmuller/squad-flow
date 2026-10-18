import PrimeButton from 'primevue/button'
import { loading, withIcons } from './withIcons'

export default withIcons(PrimeButton, () => ({ loadingicon: loading }))
