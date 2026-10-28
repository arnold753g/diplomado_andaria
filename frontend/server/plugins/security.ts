import { randomBytes } from 'node:crypto'
import { setResponseHeader } from 'h3'
import { contentSecurityPolicy, nonceScripts } from '../utils/csp'

export default defineNitroPlugin(nitro => {
  nitro.hooks.hook('render:html', (html, { event }) => {
    if (process.env.NODE_ENV !== 'production') return
    const nonce = randomBytes(32).toString('base64')
    html.head = nonceScripts(html.head, nonce)
    html.bodyPrepend = nonceScripts(html.bodyPrepend, nonce)
    html.bodyAppend = nonceScripts(html.bodyAppend, nonce)
    const header = process.env.NUXT_CSP_REPORT_ONLY === 'true'
      ? 'Content-Security-Policy-Report-Only' : 'Content-Security-Policy'
    setResponseHeader(event, header, contentSecurityPolicy(nonce, useRuntimeConfig(event).public.apiBase))
    // Prevent a shared cache from replaying a nonce or private SSR state.
    setResponseHeader(event, 'Cache-Control', 'private, no-store')
    setResponseHeader(event, 'X-Content-Type-Options', 'nosniff')
    setResponseHeader(event, 'X-Frame-Options', 'DENY')
    setResponseHeader(event, 'Referrer-Policy', 'no-referrer')
  })
})
