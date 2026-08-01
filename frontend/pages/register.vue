<template>
  <main class="auth-page">
    <div class="auth-shell">
      <aside class="auth-showcase" aria-label="Características de la plataforma">
        <div class="auth-showcase__content">
          <span class="auth-eyebrow">Acceso simple</span>
          <p class="auth-showcase-title">Tu espacio empieza <span>aquí.</span></p>
          <p>Crea tu cuenta de turista y prepara tus próximas experiencias en Tarija.</p>
        </div>
        <div class="auth-features">
          <div class="auth-feature"><i class="pi pi-lock" aria-hidden="true" /><span>Tu información personal protegida</span></div>
          <div class="auth-feature"><i class="pi pi-check-circle" aria-hidden="true" /><span>Todos tus datos en un solo lugar</span></div>
          <div class="auth-feature"><i class="pi pi-mobile" aria-hidden="true" /><span>Accede desde tu celular o computadora</span></div>
        </div>
      </aside>
      <div class="auth-panel">
        <Card class="auth-card">
          <template #content>
            <NuxtLink class="auth-brand" to="/login"><img :src="branding.logo" :alt="branding.name"><span class="brand-name">{{ branding.name }}</span></NuxtLink>
            <div class="page-heading"><h1>Crear cuenta</h1><p>Regístrate como turista. Puedes completar tus datos personales desde tu perfil.</p></div>
            <Message v-if="errorMessage" severity="error" class="form-error" :closable="false">{{ errorMessage }}</Message>
            <form class="stack" novalidate @submit.prevent="submit">
              <div class="form-field"><label for="first-name">Nombre</label><InputText id="first-name" v-model.trim="form.first_name" autocomplete="given-name" :invalid="Boolean(errors.first_name)" /><small v-if="errors.first_name" class="field-error">{{ errors.first_name }}</small></div>
              <div class="form-field"><label for="last-name">Apellido</label><InputText id="last-name" v-model.trim="form.last_name" autocomplete="family-name" :invalid="Boolean(errors.last_name)" /><small v-if="errors.last_name" class="field-error">{{ errors.last_name }}</small></div>
              <div class="form-field"><label for="email">Correo electrónico</label><InputText id="email" v-model.trim="form.email" type="email" autocomplete="email" :invalid="Boolean(errors.email)" /><small v-if="errors.email" class="field-error">{{ errors.email }}</small></div>
              <div class="form-field"><label for="phone">Teléfono (opcional)</label><InputText id="phone" v-model.trim="form.phone" type="tel" autocomplete="tel" maxlength="30" /></div>
              <div class="form-field"><label for="document">Documento o pasaporte (opcional)</label><InputText id="document" v-model.trim="form.document_number" maxlength="40" /></div>
              <div class="form-field"><label for="nationality">Nacionalidad (opcional)</label><InputText id="nationality" v-model.trim="form.nationality" maxlength="80" /></div>
              <div class="form-field"><label for="password">Contraseña</label><Password input-id="password" v-model="form.password" autocomplete="new-password" toggle-mask :invalid="Boolean(errors.password)" /><small class="form-help">Usa entre 12 y 72 bytes, sin espacios al inicio o final.</small><small v-if="errors.password" class="field-error">{{ errors.password }}</small></div>
              <div class="form-field"><label for="confirm-password">Confirmar contraseña</label><Password input-id="confirm-password" v-model="confirmation" autocomplete="new-password" :feedback="false" toggle-mask /></div>
              <div class="auth-actions"><Button type="submit" label="Crear cuenta" icon="pi pi-user-plus" fluid :loading="submitting" /><NuxtLink to="/login">Ya tengo una cuenta</NuxtLink></div>
            </form>
            <AuthGoogleAccess />
          </template>
        </Card>
      </div>
    </div>
  </main>
</template>

<script setup lang="ts">
import Button from 'primevue/button'
import Card from 'primevue/card'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Password from 'primevue/password'
import { z } from 'zod'
import { validPassword } from '~/utils/password'

definePageMeta({ layout: false, middleware: 'guest' })
useHead({ title: 'Crear cuenta' })
const config = useRuntimeConfig()
if (!config.public.registrationEnabled) await navigateTo('/login')
const auth = useAuthStore()
const branding = useBranding()
const submitting = ref(false)
const confirmation = ref('')
const errorMessage = ref('')
const errors = reactive<Record<string, string>>({})
const form = reactive({ first_name: '', last_name: '', email: '', password: '', phone: '', document_number: '', nationality: '' })
const schema = z.object({
  first_name: z.string().trim().min(2, 'Ingresa al menos 2 caracteres').max(100),
  last_name: z.string().trim().min(2, 'Ingresa al menos 2 caracteres').max(100),
  email: z.string().email('Ingresa un correo válido').max(320),
  phone: z.string().trim().max(30),
  document_number: z.string().trim().max(40),
  nationality: z.string().trim().max(80),
  password: z.string().refine(validPassword, 'Usa entre 12 y 72 bytes, sin espacios en los extremos')
})

const submit = async () => {
  if (submitting.value) return
  errorMessage.value = ''
  if (form.password !== confirmation.value) { errorMessage.value = 'Las contraseñas no coinciden.'; return }
  Object.keys(errors).forEach(key => delete errors[key])
  const parsed = schema.safeParse(form)
  if (!parsed.success) {
    for (const issue of parsed.error.issues) errors[String(issue.path[0])] = issue.message
    return
  }
  submitting.value = true
  const result = await auth.register(parsed.data)
  submitting.value = false
  if (!result.ok) {
    errorMessage.value = result.error?.message || 'No se pudo crear la cuenta'
    Object.assign(errors, result.error?.fields || {})
    return
  }
  await navigateTo('/login?registered=true')
}
</script>
