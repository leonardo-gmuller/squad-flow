import PrimeDrawer from 'primevue/drawer'
import { close, withIcons } from './withIcons'

export default withIcons(PrimeDrawer, () => ({ closeicon: close }))
