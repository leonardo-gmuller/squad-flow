import PrimeDialog from 'primevue/dialog'
import { close, withIcons } from './withIcons'

export default withIcons(PrimeDialog, () => ({ closeicon: close }))
