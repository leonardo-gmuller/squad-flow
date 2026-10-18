import PrimeMessage from 'primevue/message'
import { icon, close, severityIcon, withIcons } from './withIcons'

export default withIcons(PrimeMessage, (attrs) => ({
  icon: (scope) => icon(severityIcon(attrs.severity))(scope),
  closeicon: close,
}))
