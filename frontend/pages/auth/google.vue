<template>
  <main class="auth-page">
    <Card class="auth-card" style="max-width: 32rem; margin: auto">
      <template #title>{{ busy ? 'Comprobando tu cuenta' : 'Acceso con Google' }}</template>
      <template #content>
        <UiLoadingState v-if="busy" />
        <div v-else class="stack">
          <Message severity="warn" :closable="false">{{ message }}</Message>
          <NuxtLink to="/login">Volver a iniciar sesión</NuxtLink>
        </div>
      </template>
    </Card>
  </main>
</template>
<script setup lang="ts">
import Card from 'primevue/card'
import Message from 'primevue/message'
definePageMeta({ layout: false })
useHead({ title: 'Acceso con Google', meta: [{ name: 'referrer', content: 'no-referrer' }] })
const route = useRoute()
const auth = useAuthStore()
const busy = ref(true)
const message = ref('No se pudo completar el ingreso. Vuelve a intentarlo.')
const messages: Record<string, string> = {
  cancelled: 'Cancelaste el acceso con Google. Puedes volver a intentarlo.',
  invalid_state: 'El intento venció o se abrió desde otro navegador. Inicia el acceso nuevamente.',
  email_exists: 'Ya tienes una cuenta con este correo. Ingresa con tu contraseña y vincula Google desde Mi perfil.',
  registration_disabled: 'El registro de cuentas nuevas está deshabilitado.',
  inactive: 'Tu cuenta está desactivada. Contacta al administrador.',
  link_invalid: 'No se pudo vincular la cuenta. Ingresa de nuevo y elige en Google el mismo correo de tu perfil.'
}
onMounted(async () => {
  const result = String(route.query.result || '')
  if (result === 'success' || result === 'linked') {
    if (await auth.initialize(true)) {
      await navigateTo(result === 'linked' || !auth.user?.phone || !auth.user?.document_number ? '/app/profile' : auth.isAdmin ? '/admin' : '/app', { replace: true })
      return
    }
  }
  message.value = messages[result] || message.value
  busy.value = false
})
</script>
