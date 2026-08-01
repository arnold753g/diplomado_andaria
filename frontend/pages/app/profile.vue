<template>
  <section>
    <div class="page-heading"><h1>Mi perfil</h1><p>Mantén tus datos actualizados para tus próximas compras.</p></div>
    <div class="detail-grid">
      <Card>
        <template #title>Datos personales</template>
        <template #content>
          <Message v-if="profileError" severity="error" :closable="false">{{ profileError }}</Message>
          <form class="stack" novalidate @submit.prevent="saveProfile">
            <div class="form-field"><label for="first-name">Nombre</label><InputText id="first-name" v-model.trim="profile.first_name" autocomplete="given-name" /></div>
            <div class="form-field"><label for="last-name">Apellido</label><InputText id="last-name" v-model.trim="profile.last_name" autocomplete="family-name" /></div>
            <div class="form-field"><label for="email">Correo</label><InputText id="email" :model-value="auth.user?.email" disabled /></div>
            <div class="form-field"><label for="phone">Teléfono</label><InputText id="phone" v-model.trim="profile.phone" type="tel" autocomplete="tel" maxlength="30" /></div>
            <div class="form-field"><label for="document">Documento de identidad o pasaporte</label><InputText id="document" v-model.trim="profile.document_number" maxlength="40" /></div>
            <div class="form-field"><label for="nationality">Nacionalidad</label><InputText id="nationality" v-model.trim="profile.nationality" maxlength="80" /></div>
            <Button type="submit" label="Guardar perfil" icon="pi pi-save" :loading="savingProfile" />
          </form>
        </template>
      </Card>
      <Card>
        <template #title>{{ auth.user?.has_password ? 'Cambiar contraseña' : 'Crear contraseña' }}</template>
        <template #content>
          <Message v-if="passwordError" severity="error" :closable="false">{{ passwordError }}</Message>
          <form class="stack" novalidate @submit.prevent="changePassword">
            <div v-if="auth.user?.has_password" class="form-field"><label for="current-password">Contraseña actual</label><Password input-id="current-password" v-model="password.current" autocomplete="current-password" :feedback="false" toggle-mask /></div>
            <div class="form-field"><label for="new-password">Nueva contraseña</label><Password input-id="new-password" v-model="password.next" autocomplete="new-password" toggle-mask /><small class="form-help">Mínimo 12 caracteres. Al cambiarla se cerrarán todas tus sesiones.</small></div>
            <div class="form-field"><label for="confirm-password">Confirmar nueva contraseña</label><Password input-id="confirm-password" v-model="password.confirm" autocomplete="new-password" :feedback="false" toggle-mask /></div>
            <Button type="submit" :label="auth.user?.has_password ? 'Cambiar contraseña' : 'Crear contraseña'" icon="pi pi-key" :loading="savingPassword" />
          </form>
        </template>
      </Card>
      <Card><template #title>Cuenta de Google</template><template #content>
        <Message v-if="auth.user?.google_linked" severity="success" :closable="false">Tu cuenta está vinculada a Google. Puedes ingresar con ese método.</Message>
        <template v-else><p>Vincula el mismo correo de tu perfil para ingresar también con Google.</p><AuthGoogleAccess link /></template>
      </template></Card>
    </div>
  </section>
</template>

<script setup lang="ts">
import Button from 'primevue/button'
import Card from 'primevue/card'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Password from 'primevue/password'
import { useToast } from 'primevue/usetoast'
import type { User } from '~/types/auth'
import { apiError } from '~/utils/api-error'
import { validPassword } from '~/utils/password'

definePageMeta({ layout: 'app', middleware: 'auth' })
useHead({ title: 'Mi perfil' })
const auth = useAuthStore()
const toast = useToast()
const profile = reactive({ first_name: auth.user?.first_name || '', last_name: auth.user?.last_name || '', phone: auth.user?.phone || '', document_number: auth.user?.document_number || '', nationality: auth.user?.nationality || '' })
const password = reactive({ current: '', next: '', confirm: '' })
onMounted(async () => {
  if (await auth.initialize()) Object.assign(profile, { first_name: auth.user?.first_name || '', last_name: auth.user?.last_name || '', phone: auth.user?.phone || '', document_number: auth.user?.document_number || '', nationality: auth.user?.nationality || '' })
})
const profileError = ref('')
const passwordError = ref('')
const savingProfile = ref(false)
const savingPassword = ref(false)

const saveProfile = async () => {
  if (savingProfile.value) return
  profileError.value = ''
  if (profile.first_name.trim().length < 2 || profile.last_name.trim().length < 2) { profileError.value = 'Nombre y apellido deben tener al menos 2 caracteres.'; return }
  savingProfile.value = true
  try {
    const response = await auth.request<User>('/me', { method: 'PATCH', body: profile })
    if (response.data) auth.user = response.data
    toast.add({ severity: 'success', summary: 'Perfil actualizado', life: 3500 })
  } catch (error) { profileError.value = apiError(error).message }
  finally { savingProfile.value = false }
}

const changePassword = async () => {
  if (savingPassword.value) return
  passwordError.value = ''
  if (password.next !== password.confirm) { passwordError.value = 'Las contraseñas no coinciden.'; return }
  if (!validPassword(password.next)) { passwordError.value = 'Usa entre 12 y 72 bytes, sin espacios al inicio o al final.'; return }
  savingPassword.value = true
  try {
    await auth.request('/me/password', { method: 'POST', body: { current_password: password.current, new_password: password.next } })
    auth.clear()
    await navigateTo('/login?password_changed=true')
  } catch (error) { passwordError.value = apiError(error).message }
  finally { savingPassword.value = false }
}
</script>
