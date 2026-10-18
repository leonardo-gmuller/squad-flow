import { createApp } from 'vue'
import { createPinia } from 'pinia'
import PrimeVue from 'primevue/config'
import ToastService from 'primevue/toastservice'
import Tooltip from 'primevue/tooltip'
import { primeVueConfig } from './config/primevue'
import router from './router'
import App from './App.vue'
import './style.css'
import './styles/workspace.css'
import './styles/atmosphere.css'
import './styles/buttons.css'
import './styles/overlays.css'

document.documentElement.classList.toggle(
  'app-dark',
  localStorage.getItem('squadflow:dark') !== 'false',
)

createApp(App)
  .use(createPinia())
  .use(router)
  .use(PrimeVue, primeVueConfig)
  .use(ToastService)
  .directive('tooltip', Tooltip)
  .mount('#app')
