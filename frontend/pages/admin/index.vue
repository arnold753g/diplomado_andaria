<template>
  <section class="admin-dashboard">
    <div class="dashboard-hero"><div><span class="auth-eyebrow">Administración general</span><h1>Estado de Andaria</h1><p>Supervisa cuentas, organizaciones y contenido publicado desde un solo lugar.</p></div><NuxtLink class="public-link" to="/"><i class="pi pi-external-link" aria-hidden="true" /> Ver portada pública</NuxtLink></div>
    <UiLoadingState v-if="loading" />
    <Message v-else-if="errorMessage" severity="error" :closable="false">{{ errorMessage }} <Button label="Reintentar" text @click="load" /></Message>
    <template v-else>
      <div class="metric-grid">
        <UiDashboardMetric label="Usuarios activos" :value="stats.active_users" icon="pi pi-users" :hint="`${stats.users} cuentas registradas`" to="/admin/users" />
        <UiDashboardMetric label="Agencias publicadas" :value="stats.published_agencies" icon="pi pi-building" :hint="`${stats.agencies} agencias registradas`" to="/admin/agencies" tone="info" />
        <UiDashboardMetric label="Atracciones publicadas" :value="stats.published_attractions" icon="pi pi-map-marker" :hint="`${stats.attractions} atracciones registradas`" to="/admin/attractions" tone="success" />
        <UiDashboardMetric label="Paquetes publicados" :value="stats.published_packages" icon="pi pi-briefcase" :hint="`${stats.packages} paquetes creados`" to="/packages" tone="warning" />
      </div>
      <section class="admin-actions"><div><span class="auth-eyebrow">Accesos directos</span><h2>Gestión principal</h2><p>Las altas y asignaciones continúan bajo control del administrador.</p></div><div class="action-grid"><NuxtLink to="/admin/users"><i class="pi pi-user-plus" aria-hidden="true" /><span><strong>Usuarios y roles</strong><small>{{ stats.active_admins }} administradores activos</small></span></NuxtLink><NuxtLink to="/admin/agencies"><i class="pi pi-building" aria-hidden="true" /><span><strong>Agencias</strong><small>Crear y asignar encargados</small></span></NuxtLink><NuxtLink to="/admin/attractions"><i class="pi pi-map-marker" aria-hidden="true" /><span><strong>Atracciones</strong><small>Crear y asignar responsables</small></span></NuxtLink></div></section>
      <div class="purchase-note"><i class="pi pi-chart-line" aria-hidden="true" /><div><strong>{{ stats.purchases }} compras registradas</strong><span>Este indicador incluye todos los estados y sirve como referencia operativa.</span></div></div>
    </template>
  </section>
</template>

<script setup lang="ts">
import Button from 'primevue/button'
import Message from 'primevue/message'
import type { AdminDashboard } from '~/types/dashboard'
import { apiError } from '~/utils/api-error'

definePageMeta({ layout: 'admin', middleware: 'admin' })
useHead({ title: 'Administración' })
const auth = useAuthStore()
const loading = ref(true)
const errorMessage = ref('')
const stats = reactive<AdminDashboard>({ users: 0, active_users: 0, active_admins: 0, agencies: 0, published_agencies: 0, attractions: 0, published_attractions: 0, packages: 0, published_packages: 0, purchases: 0 })
const load = async () => {
  loading.value = true; errorMessage.value = ''
  try {
    const response = await auth.request<AdminDashboard>('/admin/dashboard')
    if (response.data) Object.assign(stats, response.data)
  } catch (error) { errorMessage.value = apiError(error).message }
  finally { loading.value = false }
}
onMounted(load)
</script>

<style scoped>
.admin-dashboard { display:grid; gap:1.5rem; }.dashboard-hero { display:flex; align-items:end; justify-content:space-between; gap:2rem; padding:clamp(1.4rem,4vw,2.5rem); border:1px solid var(--color-border-soft); border-radius:var(--radius-xl); background:radial-gradient(circle at 92% 10%,var(--color-primary-soft),transparent 18rem),linear-gradient(145deg,var(--color-surface),var(--color-background-deep)); }.dashboard-hero h1 { margin:.3rem 0 .6rem; font-size:clamp(2.3rem,5vw,4rem); }.dashboard-hero p { margin:0; color:var(--color-text-muted); }.public-link { min-height:2.8rem; display:inline-flex; align-items:center; gap:.5rem; padding:.65rem 1rem; border:1px solid var(--color-primary); border-radius:var(--radius-sm); font-weight:700; text-decoration:none; }.metric-grid { display:grid; grid-template-columns:repeat(4,minmax(0,1fr)); gap:1rem; }.admin-actions { display:grid; grid-template-columns:minmax(15rem,.65fr) minmax(0,1.35fr); gap:2rem; padding:1.4rem; border:1px solid var(--color-border); border-radius:var(--radius-lg); background:var(--color-surface); }.admin-actions h2 { margin:.25rem 0 .5rem; }.admin-actions p { margin:0; color:var(--color-text-muted); }.action-grid { display:grid; grid-template-columns:repeat(3,1fr); gap:.75rem; }.action-grid a { min-height:8rem; display:grid; align-content:center; gap:.75rem; padding:1rem; border:1px solid var(--color-border-soft); border-radius:var(--radius-md); color:var(--color-text); text-decoration:none; }.action-grid a:hover { border-color:var(--color-primary); }.action-grid i { color:var(--color-primary); font-size:1.35rem; }.action-grid span { display:grid; }.action-grid small { color:var(--color-text-muted); }.purchase-note { display:flex; align-items:center; gap:1rem; padding:1rem 1.2rem; border-left:3px solid var(--color-primary); background:var(--color-primary-muted); }.purchase-note > i { color:var(--color-primary); font-size:1.4rem; }.purchase-note div { display:grid; }.purchase-note span { color:var(--color-text-muted); }
@media(max-width:1100px) { .metric-grid { grid-template-columns:repeat(2,minmax(0,1fr)); }.admin-actions { grid-template-columns:1fr; } }
@media(max-width:720px) { .dashboard-hero { align-items:flex-start; flex-direction:column; }.public-link { width:100%; justify-content:center; }.metric-grid,.action-grid { grid-template-columns:1fr; }.admin-actions { gap:1rem; } }
</style>
