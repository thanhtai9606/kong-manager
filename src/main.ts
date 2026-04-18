import { createApp } from 'vue'
import { Translation } from '@kong-ui-public/i18n'
import type english from '@/locales/en.json'
import App from '@/App.vue'
import { router } from '@/router'
import { registerGlobalComponents } from './registerGlobalComponents'
import { provideKongAxios } from '@/services/kongAxios'
import './styles/index'
import { createPinia } from 'pinia'
import { applyI18nLocale, resolveInitialLocale } from '@/locales/catalog'

// This only sets up worker initializers. They will be lazy-loaded when needed.
import '@/monaco-workers'

const i18n = applyI18nLocale(resolveInitialLocale())

const app = createApp(App)

provideKongAxios(app)

const pinia = createPinia()

app.use(Translation.install<typeof english>, { i18n })
app.use(pinia)
app.use(router)
registerGlobalComponents(app)
app.mount('#app')
