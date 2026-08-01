<template>
  <div class="stack" style="margin-top: var(--space-4)">
    <Button v-if="enabled" :label="link ? 'Vincular mi cuenta de Google' : 'Continuar con Google'" icon="pi pi-google" severity="secondary" outlined fluid :loading="busy" @click="start" />
    <small v-else-if="loaded" class="form-help">El ingreso con Google estará disponible próximamente. Puedes usar tu correo y contraseña.</small>
    <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>
  </div>
</template>

<script setup lang="ts">
import Button from 'primevue/button'
import Message from 'primevue/message'
import type { ApiEnvelope } from '~/types/api'
import { apiError } from '~/utils/api-error'
const props = defineProps<{ link?: boolean }>()
const config = useRuntimeConfig()
const auth = useAuthStore()
const enabled = ref(false)
const loaded = ref(false)
const busy = ref(false)
const error = ref('')
onMounted(async () => {
  try {
    const result = await $fetch<ApiEnvelope<{ google_enabled: boolean }>>(`${config.public.apiBase}/auth/options`, { credentials: 'include' })
    enabled.value = Boolean(result.data?.google_enabled)
    loaded.value = true
  } catch { error.value = 'No se pudo consultar la disponibilidad de Google.' }
})
const start = async () => {
  if (busy.value) return
  busy.value = true; error.value = ''
  try {
    const result = props.link
      ? await auth.request<{ url: string }>('/me/google/start', { method: 'POST' })
      : await $fetch<ApiEnvelope<{ url: string }>>(`${config.public.apiBase}/auth/google/start`, { method: 'POST', credentials: 'include' })
    const destination = new URL(result.data?.url || '')
    if (destination.origin !== 'https://accounts.google.com') throw new Error('Invalid OAuth destination')
    window.location.assign(destination.href)
  } catch (cause) { error.value = apiError(cause, 'No se pudo iniciar el acceso con Google').message; busy.value = false }
}
</script>
