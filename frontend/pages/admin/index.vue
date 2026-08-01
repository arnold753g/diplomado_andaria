<template>
  <section>
    <div class="page-heading"><h1>Administración</h1><p>Resumen de usuarios y acceso a Andaria.</p></div>
    <UiLoadingState v-if="loading" />
    <Message v-else-if="errorMessage" severity="error" :closable="false">{{ errorMessage }}</Message>
    <div v-else class="stat-grid">
      <Card><template #content><i class="pi pi-users" aria-hidden="true" /><span class="muted">Usuarios totales</span><div class="stat-value">{{ stats.users }}</div></template></Card>
      <Card><template #content><i class="pi pi-user-plus" aria-hidden="true" /><span class="muted">Usuarios activos</span><div class="stat-value">{{ stats.active_users }}</div></template></Card>
      <Card><template #content><i class="pi pi-shield" aria-hidden="true" /><span class="muted">Administradores activos</span><div class="stat-value">{{ stats.active_admins }}</div></template></Card>
    </div>
  </section>
</template>

<script setup lang="ts">
import Card from 'primevue/card'
import Message from 'primevue/message'
import { apiError } from '~/utils/api-error'

definePageMeta({ layout: 'admin', middleware: 'admin' })
useHead({ title: 'Administración' })
const auth = useAuthStore()
const loading = ref(true)
const errorMessage = ref('')
const stats = reactive({ users: 0, active_users: 0, active_admins: 0 })
onMounted(async () => {
  try {
    const response = await auth.request<typeof stats>('/admin/dashboard')
    Object.assign(stats, response.data)
  } catch (error) { errorMessage.value = apiError(error).message }
  finally { loading.value = false }
})
</script>
