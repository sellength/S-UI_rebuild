/**
 * plugins/vuetify.ts
 *
 * Framework documentation: https://vuetifyjs.com`
 */

// Styles
import '@mdi/font/css/materialdesignicons.css'
import 'vuetify/styles'

import colors from 'vuetify/util/colors'
import { fa, en, vi, zhHans, zhHant } from 'vuetify/locale'

// Composables
import { createVuetify } from 'vuetify'

// https://vuetifyjs.com/en/introduction/why-vuetify/#feature-guides
export default createVuetify({
  defaults: {
    VRow: { dense: true }, // Apply dense to v-row as default
    VTextField: { variant: 'outlined', density: 'comfortable' },
    VSelect: { variant: 'outlined', density: 'comfortable' }
  },
  theme: {
    defaultTheme: localStorage.getItem('theme') ?? 'light',
    themes: {
      light: {
        colors: {
          primary: '#1867C0',
          secondary: '#5CBBF6',
          tertiary: '#E57373',
          accent: '#005CAF',
          error: colors.red.accent3,
          warning: colors.amber.base,
          info: colors.teal.darken1,
          success: colors.green.base,
          background: colors.grey.lighten4,
        },
      },
      dark: {
        colors: {
          primary: '#06b6d4',
          secondary: '#8b5cf6',
          accent: '#8b5cf6',
          error: '#ef4444',
          warning: '#f59e0b',
          info: '#3b82f6',
          success: '#10b981',
          surface: '#111827',
          background: '#0b0f19',
        },
      },
    },
  },
  locale: {
    locale: localStorage.getItem("locale") ?? 'en',
    fallback: 'en',
    messages: { en, fa, vi, zhHans, zhHant },
  },
})
