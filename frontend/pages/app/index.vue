<template>
  <section class="dashboard-page">
    <header class="dashboard-hero">
      <div><span class="auth-eyebrow">{{ roleLabel(auth.user?.role) }}</span><h1>Hola, {{ auth.user?.first_name }}</h1><p>{{ welcomeMessage }}</p></div>
      <NuxtLink class="public-home-link" to="/"><i class="pi pi-compass" aria-hidden="true" /> Ver portada pública</NuxtLink>
    </header>

    <UiLoadingState v-if="loading" />
    <Message v-else-if="error" severity="error" :closable="false">{{ error }} <Button label="Reintentar" text @click="load" /></Message>

    <template v-else-if="dashboard">
      <template v-if="dashboard.tourist">
        <div class="metric-grid">
          <UiDashboardMetric label="Compras activas" :value="dashboard.tourist.active_purchases" icon="pi pi-shopping-bag" hint="En revisión, corrección o confirmadas" to="/app/purchases" />
          <UiDashboardMetric label="Reembolsos pendientes" :value="dashboard.tourist.pending_refunds" icon="pi pi-wallet" hint="Devoluciones todavía en proceso" to="/app/purchases" :tone="dashboard.tourist.pending_refunds ? 'warning' : 'success'" />
          <UiDashboardMetric label="Atracciones favoritas" :value="dashboard.tourist.favorites" icon="pi pi-heart" hint="Lugares guardados en tu colección" to="/app/favorites" tone="info" />
        </div>
        <div class="dashboard-columns">
          <section class="action-panel" aria-labelledby="next-trip-title">
            <div class="panel-heading"><div><span class="auth-eyebrow">Tu actividad</span><h2 id="next-trip-title">Próxima salida</h2></div><NuxtLink to="/app/purchases">Ver compras</NuxtLink></div>
            <div v-if="dashboard.tourist.next_purchase" class="next-trip">
              <div class="date-block"><strong>{{ dayNumber(dashboard.tourist.next_purchase.departure_start) }}</strong><span>{{ monthShort(dashboard.tourist.next_purchase.departure_start) }}</span></div>
              <div><Tag :value="purchaseStatus(dashboard.tourist.next_purchase.status)" :severity="purchaseSeverity(dashboard.tourist.next_purchase.status)" /><h3>{{ dashboard.tourist.next_purchase.package_name }}</h3><p>{{ dateTime(dashboard.tourist.next_purchase.departure_start) }} · {{ dashboard.tourist.next_purchase.reference }}</p></div>
            </div>
            <UiEmptyState v-else icon="pi pi-calendar" title="Todavía no tienes una salida próxima" description="Explora paquetes con fechas y cupos disponibles."><NuxtLink class="panel-action" to="/packages">Explorar paquetes</NuxtLink></UiEmptyState>
          </section>
          <section class="action-panel" aria-labelledby="discover-title"><div class="panel-heading"><div><span class="auth-eyebrow">Descubre</span><h2 id="discover-title">Continúa explorando</h2></div></div><div class="quick-actions"><NuxtLink to="/packages"><i class="pi pi-briefcase" aria-hidden="true" /><span><strong>Paquetes turísticos</strong><small>Fechas, itinerarios y precios</small></span><i class="pi pi-arrow-right" aria-hidden="true" /></NuxtLink><NuxtLink to="/attractions"><i class="pi pi-map-marker" aria-hidden="true" /><span><strong>Atracciones</strong><small>Lugares para tu próxima ruta</small></span><i class="pi pi-arrow-right" aria-hidden="true" /></NuxtLink></div></section>
        </div>
      </template>

      <template v-else-if="dashboard.agency">
        <Message v-if="!dashboard.agency.assigned" severity="warn" :closable="false">El administrador todavía debe asignarte una agencia.</Message>
        <template v-else>
          <Message v-if="dashboard.agency.overdue_refunds" severity="error" :closable="false">Tienes {{ dashboard.agency.overdue_refunds }} {{ dashboard.agency.overdue_refunds === 1 ? 'reembolso vencido' : 'reembolsos vencidos' }}. Registra la devolución y su comprobante.</Message>
          <Message v-else-if="dashboard.agency.minimum_reviews" severity="warn" :closable="false">Hay {{ dashboard.agency.minimum_reviews }} {{ dashboard.agency.minimum_reviews === 1 ? 'salida pendiente' : 'salidas pendientes' }} de decisión por cupo mínimo.</Message>
          <div class="metric-grid">
            <UiDashboardMetric label="Pagos por revisar" :value="dashboard.agency.payments_to_review" icon="pi pi-receipt" hint="Comprobantes recibidos" to="/agency/purchases" :tone="dashboard.agency.payments_to_review ? 'warning' : 'success'" />
            <UiDashboardMetric label="Decisiones por mínimo" :value="dashboard.agency.minimum_reviews" icon="pi pi-users" hint="Cierres que requieren confirmación" to="/agency/packages" :tone="dashboard.agency.minimum_reviews ? 'warning' : 'success'" />
            <UiDashboardMetric label="Reembolsos pendientes" :value="dashboard.agency.pending_refunds" icon="pi pi-wallet" :hint="refundHint" to="/agency/purchases" :tone="dashboard.agency.overdue_refunds ? 'danger' : dashboard.agency.pending_refunds ? 'warning' : 'success'" />
            <UiDashboardMetric label="Paquetes publicados" :value="`${dashboard.agency.published_packages}/${dashboard.agency.packages}`" icon="pi pi-briefcase" hint="Publicados sobre el total" to="/agency/packages" tone="info" />
          </div>
          <section class="action-panel"><div class="panel-heading"><div><span class="auth-eyebrow">{{ dashboard.agency.agency_name }}</span><h2>Acciones frecuentes</h2></div></div><div class="quick-actions three"><NuxtLink to="/agency/purchases"><i class="pi pi-check-square" aria-hidden="true" /><span><strong>Revisar compras</strong><small>Validar pagos y devoluciones</small></span><i class="pi pi-arrow-right" aria-hidden="true" /></NuxtLink><NuxtLink to="/agency/packages"><i class="pi pi-calendar" aria-hidden="true" /><span><strong>Gestionar salidas</strong><small>Fechas, cupos y mínimos</small></span><i class="pi pi-arrow-right" aria-hidden="true" /></NuxtLink><NuxtLink to="/agency"><i class="pi pi-credit-card" aria-hidden="true" /><span><strong>Configurar agencia</strong><small>Política de menores y pagos</small></span><i class="pi pi-arrow-right" aria-hidden="true" /></NuxtLink></div></section>
        </template>
      </template>

      <template v-else-if="dashboard.attraction_manager">
        <Message v-if="!dashboard.attraction_manager.assigned_attractions" severity="info" :closable="false">El administrador todavía no te asignó atracciones.</Message>
        <div class="metric-grid">
          <UiDashboardMetric label="Atracciones asignadas" :value="dashboard.attraction_manager.assigned_attractions" icon="pi pi-map-marker" hint="Total bajo tu responsabilidad" to="/managed-attractions" />
          <UiDashboardMetric label="Publicadas" :value="dashboard.attraction_manager.published_attractions" icon="pi pi-eye" hint="Visibles en el catálogo" to="/managed-attractions" tone="success" />
          <UiDashboardMetric label="Por completar" :value="dashboard.attraction_manager.draft_attractions" icon="pi pi-pencil" hint="Activas sin publicar" to="/managed-attractions" :tone="dashboard.attraction_manager.draft_attractions ? 'warning' : 'success'" />
          <UiDashboardMetric label="Inactivas" :value="dashboard.attraction_manager.inactive_attractions" icon="pi pi-ban" hint="No disponibles para publicación" to="/managed-attractions" tone="info" />
        </div>
        <section class="action-panel"><div class="panel-heading"><div><span class="auth-eyebrow">Contenido turístico</span><h2>Mantén cada ficha lista para los visitantes</h2></div></div><p class="panel-description">Revisa fotografías, clasificación, horarios, temporada y ubicación antes de publicar.</p><NuxtLink class="panel-action" to="/managed-attractions">Gestionar mis atracciones <i class="pi pi-arrow-right" aria-hidden="true" /></NuxtLink></section>
      </template>

      <div class="account-strip"><div><span class="auth-eyebrow">Cuenta</span><h2>Perfil y acceso</h2><p>{{ auth.user?.email }} · {{ auth.user?.google_linked ? 'Google vinculado' : 'Correo y contraseña' }}</p></div><NuxtLink class="panel-action" to="/app/profile">Revisar mis datos</NuxtLink></div>
    </template>
  </section>
</template>

<script setup lang="ts">
import Button from 'primevue/button'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import type { AccountDashboard } from '~/types/dashboard'
import { apiError } from '~/utils/api-error'
import { purchaseSeverity, purchaseStatus } from '~/utils/purchases'
import { roleLabel } from '~/utils/roles'

definePageMeta({ layout: 'app', middleware: 'user' })
useHead({ title: 'Mi panel' })
const auth = useAuthStore(); const dashboard = ref<AccountDashboard | null>(null); const loading = ref(true); const error = ref('')
const welcomeMessage = computed(() => auth.user?.role === 'turista' ? 'Organiza tus próximas experiencias y consulta el estado de cada compra.' : auth.user?.role === 'encargado_agencia' ? 'Atiende las decisiones comerciales que requieren tu intervención.' : 'Mantén actualizada la información que verán los visitantes.')
const refundHint = computed(() => { const item = dashboard.value?.agency; if (!item) return ''; if (item.overdue_refunds) return `${item.overdue_refunds} vencidos`; if (item.refunds_due_soon) return `${item.refunds_due_soon} vencen en 24 horas`; return 'Sin vencimientos próximos' })
const dateTime = (value: string) => new Intl.DateTimeFormat('es-BO', { dateStyle: 'long', timeStyle: 'short', timeZone: 'America/La_Paz' }).format(new Date(value))
const dayNumber = (value: string) => new Intl.DateTimeFormat('es-BO', { day: '2-digit', timeZone: 'America/La_Paz' }).format(new Date(value))
const monthShort = (value: string) => new Intl.DateTimeFormat('es-BO', { month: 'short', timeZone: 'America/La_Paz' }).format(new Date(value)).replace('.', '')
const load = async () => { loading.value = true; error.value = ''; try { const response = await auth.request<AccountDashboard>('/me/dashboard'); dashboard.value = response.data || null } catch (cause) { error.value = apiError(cause, 'No se pudo cargar tu panel').message } finally { loading.value = false } }
onMounted(load)
</script>

<style scoped>
.dashboard-page { display:grid; gap:1.5rem; }.dashboard-hero { display:flex; align-items:end; justify-content:space-between; gap:2rem; padding:clamp(1.4rem,4vw,2.5rem); border:1px solid var(--color-border-soft); border-radius:var(--radius-xl); background:radial-gradient(circle at 92% 12%,var(--color-primary-soft),transparent 18rem),linear-gradient(145deg,var(--color-surface),var(--color-background-deep)); }.dashboard-hero h1 { margin:.3rem 0 .6rem; font-size:clamp(2.3rem,5vw,4rem); }.dashboard-hero p { max-width:58ch; margin:0; color:var(--color-text-muted); }.public-home-link,.panel-action { min-height:2.8rem; display:inline-flex; align-items:center; justify-content:center; gap:.5rem; padding:.65rem 1rem; border:1px solid var(--color-primary); border-radius:var(--radius-sm); font-weight:700; text-decoration:none; }.metric-grid { display:grid; grid-template-columns:repeat(4,minmax(0,1fr)); gap:1rem; }.dashboard-columns { display:grid; grid-template-columns:1.15fr .85fr; gap:1rem; }.action-panel,.account-strip { padding:1.3rem; border:1px solid var(--color-border); border-radius:var(--radius-lg); background:var(--color-surface); }.panel-heading { display:flex; align-items:end; justify-content:space-between; gap:1rem; margin-bottom:1rem; }.panel-heading h2 { margin:.25rem 0 0; }.panel-heading > a { font-weight:700; text-decoration:none; }.next-trip { display:grid; grid-template-columns:auto 1fr; gap:1rem; align-items:center; padding:1rem; border-radius:var(--radius-md); background:var(--color-surface-elevated); }.next-trip h3 { margin:.55rem 0 .25rem; }.next-trip p { margin:0; color:var(--color-text-muted); }.date-block { min-width:4.4rem; display:grid; place-items:center; padding:.8rem; border:1px solid rgba(186,254,6,.24); border-radius:var(--radius-md); background:var(--color-primary-soft); }.date-block strong { color:var(--color-primary); font-size:2rem; line-height:1; }.date-block span { color:var(--color-text-muted); text-transform:uppercase; }.quick-actions { display:grid; gap:.7rem; }.quick-actions.three { grid-template-columns:repeat(3,1fr); }.quick-actions a { min-height:5.2rem; display:grid; grid-template-columns:auto 1fr auto; gap:.8rem; align-items:center; padding:.9rem; border:1px solid var(--color-border-soft); border-radius:var(--radius-md); color:var(--color-text); text-decoration:none; transition:border-color var(--motion-fast),transform var(--motion-fast); }.quick-actions a:hover { border-color:var(--color-primary); transform:translateY(-1px); }.quick-actions a > i:first-child { color:var(--color-primary); font-size:1.2rem; }.quick-actions span { display:grid; }.quick-actions small { color:var(--color-text-muted); }.panel-description { max-width:65ch; color:var(--color-text-muted); }.account-strip { display:flex; align-items:center; justify-content:space-between; gap:1rem; }.account-strip h2 { margin:.25rem 0; }.account-strip p { margin:0; color:var(--color-text-muted); }
@media(max-width:1100px) { .metric-grid { grid-template-columns:repeat(2,minmax(0,1fr)); }.dashboard-columns { grid-template-columns:1fr; }.quick-actions.three { grid-template-columns:1fr; } }
@media(max-width:680px) { .dashboard-hero,.panel-heading,.account-strip { align-items:flex-start; flex-direction:column; }.public-home-link,.account-strip .panel-action { width:100%; }.metric-grid { grid-template-columns:1fr; }.next-trip { align-items:start; }.quick-actions a { min-height:4.8rem; } }
</style>
