/**
 * main.ts
 *
 * Bootstraps Vuetify and other plugins then mounts the App`
 */

// Composables
import { createApp, ref } from 'vue'

// Components
import App from './App.vue'

// Use router
import router from './router'

// Store
import store from './store'

// Plugins
import { registerPlugins } from '@/plugins'

// Locale
import { i18n } from '@/locales'
import Vue3PersianDatetimePicker from 'vue3-persian-datetime-picker'

// Notivue
import { createNotivue } from 'notivue'
import 'notivue/notification.css'
import 'notivue/animations.css'
const notivue = createNotivue({
  position: 'bottom-center',
  limit: 4,
  enqueue: false,
  avoidDuplicates: true,
  notifications: {
    global: {
      duration: 3000
    }
  },
})

const loading = ref(false)

const app = createApp(App)
app.provide('loading', loading)

registerPlugins(app)

app.mixin({
  watch: {
    menu(this: any, val: any) {
      if (typeof val === 'boolean') {
        if (val) {
          document.addEventListener('click', this._handleOutsideClickCapture, true)
        } else {
          document.removeEventListener('click', this._handleOutsideClickCapture, true)
        }
      }
    }
  },
  beforeUnmount(this: any) {
    if (typeof this.menu === 'boolean') {
      document.removeEventListener('click', this._handleOutsideClickCapture, true)
    }
  },
  methods: {
    _handleOutsideClickCapture(this: any, e: MouseEvent) {
      const target = e.target as HTMLElement
      if (target && (target.closest('.v-menu-custom-options') || target.closest('.v-menu-custom-activator'))) {
        return
      }
      this.menu = false
    }
  }
})

app
  .use(router)
  .use(store)
  .use(i18n)
  .use(notivue)
  .component('DatePicker', Vue3PersianDatetimePicker)
  .mount('#app')
