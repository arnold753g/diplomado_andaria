import { existsSync } from 'node:fs'
import { resolve } from 'node:path'
import { loadEnvFile } from 'node:process'

const rootEnvPath = resolve(process.cwd(), '../.env')
if (existsSync(rootEnvPath)) loadEnvFile(rootEnvPath)

const isProduction = process.env.NODE_ENV === 'production'

export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: !isProduction },
  modules: ['@pinia/nuxt'],
  css: ['primeicons/primeicons.css', '~/assets/theme.css'],
  build: { transpile: ['primevue'] },
  runtimeConfig: {
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8080/api/v1',
      appName: process.env.NUXT_PUBLIC_APP_NAME || process.env.APP_NAME || 'Andaria',
      appShortName: process.env.NUXT_PUBLIC_APP_SHORT_NAME || process.env.APP_SHORT_NAME || 'Andaria',
      registrationEnabled: process.env.NUXT_PUBLIC_REGISTRATION_ENABLED !== 'false'
    }
  },
  app: {
    head: {
      htmlAttrs: { lang: 'es', class: 'app-dark' },
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: 'Andaria: experiencias y turismo en Tarija' }
      ],
      link: [
        { rel: 'icon', type: 'image/svg+xml', href: '/branding/favicon.svg' },
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' },
        { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=Outfit:wght@400;500;600;700;800&display=swap' }
      ]
    }
  },
  typescript: { strict: true, typeCheck: false }
})
