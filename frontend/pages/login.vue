<template>
  <main class="auth-page">
    <div class="auth-shell">
      <aside class="auth-showcase" aria-label="Características de la plataforma">
        <div class="auth-showcase__content">
          <span class="auth-eyebrow">Andaria · Tarija</span>
          <p class="auth-showcase-title">Tu próxima aventura <span>empieza aquí.</span></p>
          <p>Ingresa a tu cuenta para organizar tus próximas experiencias en Tarija.</p>
        </div>
        <div class="auth-features">
          <div class="auth-feature"><i class="pi pi-shield" aria-hidden="true" /><span>Tu cuenta, en un solo lugar</span></div>
          <div class="auth-feature"><i class="pi pi-palette" aria-hidden="true" /><span>Descubre nuevos destinos</span></div>
          <div class="auth-feature"><i class="pi pi-bolt" aria-hidden="true" /><span>Prepárate para tu próximo viaje</span></div>
        </div>
      </aside>
      <div class="auth-panel">
        <Card class="auth-card">
          <template #content>
            <NuxtLink class="auth-brand" to="/login"><img :src="branding.logo" :alt="branding.name"><span class="brand-name">{{ branding.name }}</span></NuxtLink>
            <div class="page-heading">
              <h1>Iniciar sesión</h1>
              <p>Accede a tu espacio con tu correo y contraseña.</p>
            </div>
            <Message v-if="successMessage" severity="success" class="form-error" :closable="false">{{ successMessage }}</Message>
            <Message v-if="errorMessage" severity="error" class="form-error" :closable="false">{{ errorMessage }}</Message>
            <form class="stack" novalidate @submit.prevent="submit">
              <div class="form-field">
                <label for="email">Correo electrónico</label>
                <InputText id="email" v-model.trim="form.email" type="email" autocomplete="username" :invalid="Boolean(errors.email)" />
                <small v-if="errors.email" class="field-error" role="alert">{{ errors.email }}</small>
              </div>
              <div class="form-field">
                <label for="password">Contraseña</label>
                <Password input-id="password" v-model="form.password" autocomplete="current-password" :feedback="false" toggle-mask :invalid="Boolean(errors.password)" />
                <small v-if="errors.password" class="field-error" role="alert">{{ errors.password }}</small>
              </div>
              <div class="auth-actions">
                <Button type="submit" label="Ingresar" icon="pi pi-sign-in" fluid :loading="submitting" />
                <NuxtLink v-if="registrationEnabled" to="/register">Crear una cuenta</NuxtLink>
              </div>
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
import { safeInternalRedirect } from '~/utils/auth-session'

definePageMeta({ layout: false, middleware: 'guest' })
useHead({ title: 'Iniciar sesión' })
const auth = useAuthStore()
const branding = useBranding()
const route = useRoute()
const config = useRuntimeConfig()
const registrationEnabled = Boolean(config.public.registrationEnabled)
const submitting = ref(false)
const errorMessage = ref('')
const successMessage = computed(() => route.query.registered ? 'Cuenta creada. Ya puedes ingresar.' : route.query.password_changed ? 'Contraseña actualizada. Inicia sesión nuevamente.' : '')
const errors = reactive<Record<string, string>>({})
const form = reactive({ email: '', password: '' })
const schema = z.object({
  email: z.string().email('Ingresa un correo válido'),
  password: z.string().min(1, 'Ingresa tu contraseña')
})

const submit = async () => {
  if (submitting.value) return
  errorMessage.value = ''
  Object.keys(errors).forEach(key => delete errors[key])
  const result = schema.safeParse(form)
  if (!result.success) {
    for (const issue of result.error.issues) errors[String(issue.path[0])] = issue.message
    return
  }
  submitting.value = true
  const response = await auth.login(form.email, form.password)
  submitting.value = false
  if (!response.ok) {
    errorMessage.value = response.error?.message || 'No se pudo iniciar sesión'
    return
  }
  const fallback = auth.isAdmin ? '/admin' : '/app'
  await navigateTo(safeInternalRedirect(route.query.redirect, fallback))
}
</script>
