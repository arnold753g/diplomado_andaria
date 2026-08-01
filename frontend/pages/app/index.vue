<template>
  <section>
    <div class="page-heading"><span class="auth-eyebrow">{{ roleLabel(auth.user?.role) }}</span><h1>Hola, {{ auth.user?.first_name }}</h1><p>Bienvenido a tu espacio en Andaria.</p></div>
    <div class="detail-grid">
      <Card><template #title>Tu perfil</template><template #content><p class="muted">Completa tu teléfono, documento y nacionalidad para tener tus datos listos.</p><NuxtLink to="/app/profile">Revisar mis datos <i class="pi pi-arrow-right" aria-hidden="true" /></NuxtLink></template></Card>
      <Card><template #title>Tu cuenta</template><template #content><dl><dt>Correo electrónico</dt><dd>{{ auth.user?.email }}</dd><dt>Acceso</dt><dd>{{ auth.user?.google_linked ? 'Google vinculado' : 'Correo y contraseña' }}</dd></dl></template></Card>
    </div>
  </section>
</template>
<script setup lang="ts">
import Card from 'primevue/card'
import { roleLabel } from '~/utils/roles'
definePageMeta({ layout: 'app', middleware: 'user' })
useHead({ title: 'Mi cuenta' })
const auth = useAuthStore()
</script>
