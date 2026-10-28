<template>
  <main class="auth-page">
    <section class="auth-card stack" aria-labelledby="verification-title" aria-live="polite">
      <h1 id="verification-title">Confirmar correo</h1>
      <p>{{ message }}</p>
      <Button v-if="token && !verified" label="Confirmar mi cuenta" :loading="busy" @click="verify" />
      <NuxtLink to="/login">Ir a iniciar sesión</NuxtLink>
    </section>
  </main>
</template>

<script setup lang="ts">
import Button from 'primevue/button'
import type { ApiEnvelope } from '~/types/api'
import { apiError } from '~/utils/api-error'

definePageMeta({ layout: false })
useHead({ title: 'Confirmar correo', meta: [{ name: 'referrer', content: 'no-referrer' }] })
const token = ref('')
const busy = ref(false)
const verified = ref(false)
const message = ref('Abre el enlace recibido por correo para confirmar tu cuenta.')
const config = useRuntimeConfig()
onMounted(() => {
  token.value = new URLSearchParams(window.location.hash.slice(1)).get('token') || ''
  // The token never reaches SSR, proxy access logs, history or referrer headers.
  window.history.replaceState(null, '', window.location.pathname)
  if (token.value) message.value = 'Confirma que solicitaste esta cuenta. El enlace vence en 30 minutos.'
})
const verify = async () => {
  if (busy.value || !token.value) return
  busy.value = true
  try {
    const result = await $fetch<ApiEnvelope<{ verified: boolean }>>(`${config.public.apiBase}/auth/verify-email`, {
      method: 'POST', credentials: 'include', body: { token: token.value }
    })
    verified.value = result.data?.verified === true
    token.value = ''
    message.value = result.message || 'Correo confirmado. Ya puedes iniciar sesión en la web o la app.'
  } catch (error) {
    message.value = apiError(error, 'No se pudo confirmar el correo').message
  } finally { busy.value = false }
}
</script>
