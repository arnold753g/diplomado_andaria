<template>
  <section class="payments-page stack" :aria-busy="overviewLoading || detailLoading || purchasesLoading">
    <header class="payments-heading">
      <div><span class="auth-eyebrow">Administración de agencia</span><h1>Gestionar pagos</h1><p>Organiza cada viaje, revisa sus pagos y confirma la salida.</p></div>
      <Button label="Actualizar" icon="pi pi-refresh" severity="secondary" outlined :loading="overviewLoading || detailLoading" @click="refresh" />
    </header>
    <Message v-if="notice" severity="success" :closable="true" @close="notice = ''">{{ notice }}</Message>
    <template v-if="!selectedId">
      <nav class="payment-views" aria-label="Vistas de salidas">
        <button v-for="view in paymentViews" :key="view.value" type="button" :aria-current="filters.state === view.value ? 'page' : undefined" :class="{ active: filters.state === view.value }" @click="setView(view.value)"><i :class="view.icon" aria-hidden="true" />{{ view.label }}<span>{{ overview?.summary?.[view.counter] || 0 }}</span></button>
      </nav>
      <div v-if="overview && metrics" class="payments-metrics">
        <UiDashboardMetric label="Pagos por revisar" :value="metrics.payments_pending" icon="pi pi-receipt" hint="Comprobantes recibidos" :tone="metrics.payments_pending ? 'warning' : 'success'" />
        <UiDashboardMetric label="Listas para confirmar" :value="metrics.ready_departures" icon="pi pi-check-circle" hint="Mínimo de cupos alcanzado" tone="info" />
        <UiDashboardMetric label="Reembolsos pendientes" :value="metrics.refunds_pending" icon="pi pi-wallet" :hint="`${metrics.refunds_overdue} vencidos`" :tone="metrics.refunds_overdue ? 'danger' : 'warning'" />
        <UiDashboardMetric label="Importe confirmado" :value="packageMoney(overview.totals.confirmed_amount_cents)" icon="pi pi-credit-card" hint="Compras confirmadas de esta vista" tone="success" />
      </div>
      <form class="payment-filters" @submit.prevent="applyFilters">
        <div class="form-field"><label for="payment-search">Buscar paquete</label><InputText id="payment-search" v-model.trim="filters.search" placeholder="Nombre del paquete" /></div>
        <div class="form-field"><label for="payment-package">Paquete</label><Select input-id="payment-package" v-model="filters.package_id" :options="packageOptions" option-label="name" option-value="id" /></div>
        <div class="form-field"><label for="payment-from">Desde</label><input id="payment-from" v-model="filters.from" type="date"></div>
        <div class="form-field"><label for="payment-to">Hasta</label><input id="payment-to" v-model="filters.to" type="date" :min="filters.from || undefined"></div>
        <div class="filter-actions"><Button type="submit" label="Filtrar" icon="pi pi-filter" :loading="overviewLoading" /><Button label="Limpiar" severity="secondary" text @click="clearFilters" /></div>
      </form>
      <p class="muted filter-note">Habilitadas muestra solo salidas futuras autorizadas. Consulta las demás en Pendientes, Reembolsos o Historial. Los avisos de pagos y reembolsos consideran todas las salidas del filtro de paquete y fechas. Horarios de Bolivia.</p>
      <Message v-if="overviewError" severity="error" :closable="false">{{ overviewError }} <Button label="Reintentar" text @click="loadOverview" /></Message>
      <UiLoadingState v-else-if="overviewLoading" />
      <UiEmptyState v-else-if="!groups.length" icon="pi pi-calendar" :title="filters.state === 'authorized' ? 'No hay salidas habilitadas' : 'No hay salidas para estos filtros'" :description="filters.state === 'authorized' ? 'Las salidas futuras pendientes de autorización están en Pendientes.' : 'Prueba otras fechas, otro paquete o una vista diferente.'"><Button v-if="filters.state === 'authorized'" label="Ver pendientes" icon="pi pi-clock" @click="setView('pending')" /></UiEmptyState>
      <div v-else class="package-groups">
        <section v-for="group in groups" :key="group.id" class="package-group" :aria-labelledby="`package-heading-${group.id}`">
          <header class="package-group-heading"><div class="package-icon"><i class="pi pi-briefcase" aria-hidden="true" /></div><div><small>Paquete turístico</small><h2 :id="`package-heading-${group.id}`">{{ group.name }}</h2></div></header>
          <div class="departure-list">
            <article v-for="departure in group.departures" :key="departure.id" class="departure-row">
              <div class="departure-date"><i class="pi pi-calendar" aria-hidden="true" /><div><strong>{{ shortDate(departure.starts_at) }}</strong><span>{{ clock(departure.starts_at) }} · {{ departure.total_purchases }} compras</span></div></div>
              <div class="departure-capacity"><span><strong>{{ departure.confirmed_capacity }}</strong> / {{ departure.max_capacity }} cupos confirmados</span><progress :value="departure.confirmed_capacity" :max="departure.max_capacity" :aria-label="`${departure.confirmed_capacity} cupos confirmados, mínimo ${departure.min_capacity}, máximo ${departure.max_capacity}`" /><small>Mínimo {{ departure.min_capacity }} · {{ departure.held_capacity }} retenidos</small></div>
              <div class="departure-alerts"><Tag :value="authorizationLabel(departure.authorization_state)" :severity="authorizationSeverity(departure.authorization_state)" /><small v-if="departure.payments_pending">{{ departure.payments_pending }} pagos por revisar</small><small v-if="departure.corrections_pending">{{ departure.corrections_pending }} correcciones solicitadas</small><small v-if="departure.refunds_pending" :class="{ overdue: departure.refunds_overdue }">{{ departure.refunds_pending }} {{ departure.refunds_pending === 1 ? 'reembolso pendiente' : 'reembolsos pendientes' }}<span v-if="departure.refunds_overdue"> · {{ departure.refunds_overdue }} vencidos</span></small></div>
              <Button label="Gestionar salida" icon="pi pi-arrow-right" icon-pos="right" severity="secondary" outlined @click="openDeparture(departure.id)" />
            </article>
          </div>
        </section>
      </div>
      <nav v-if="overview && overview.pagination.total > overview.pagination.limit" class="payments-pagination" aria-label="Páginas de salidas"><Button label="Anterior" icon="pi pi-chevron-left" severity="secondary" outlined :disabled="page <= 1 || overviewLoading" @click="changeOverviewPage(-1)" /><span>Página {{ page }} de {{ Math.ceil(overview.pagination.total / overview.pagination.limit) }} · {{ overview.pagination.total }} salidas</span><Button label="Siguiente" icon="pi pi-chevron-right" icon-pos="right" severity="secondary" outlined :disabled="page * overview.pagination.limit >= overview.pagination.total || overviewLoading" @click="changeOverviewPage(1)" /></nav>
    </template>
    <template v-else>
      <Button class="back-to-departures" label="Volver a las salidas" icon="pi pi-arrow-left" severity="secondary" text @click="closeDeparture" />
      <Message v-if="detailError" severity="error" :closable="false">{{ detailError }} <Button label="Reintentar" text @click="loadDetail" /></Message>
      <UiLoadingState v-else-if="detailLoading || !selected" />
      <template v-else>
        <section class="departure-overview">
          <div class="departure-overview-title"><div><span class="auth-eyebrow">Salida seleccionada</span><h2>{{ selected.package_name }}</h2><p><i class="pi pi-calendar" aria-hidden="true" /> {{ dateTime(selected.starts_at) }} · Bolivia</p><p v-if="selected.meeting_point"><i class="pi pi-map-marker" aria-hidden="true" /> {{ selected.meeting_point }}</p></div><Tag :value="authorizationLabel(selected.authorization_state)" :severity="authorizationSeverity(selected.authorization_state)" /></div>
          <div class="departure-numbers"><div><small>Cupos confirmados</small><strong>{{ selected.confirmed_capacity }} / {{ selected.max_capacity }}</strong><span>Mínimo requerido: {{ selected.min_capacity }}</span></div><div><small>Cupos retenidos</small><strong>{{ selected.held_capacity }}</strong><span>Pagos pendientes de validar</span></div><div><small>Turistas confirmados</small><strong>{{ selected.confirmed_travelers }}</strong><span>{{ selected.confirmed_adults }} adultos · {{ selected.confirmed_minors }} {{ selected.confirmed_minors === 1 ? 'menor' : 'menores' }}</span></div><div><small>Importe confirmado</small><strong>{{ packageMoney(selected.confirmed_amount_cents) }}</strong><span>{{ selected.confirmed_purchases }} compras confirmadas</span></div></div>
          <p class="muted">Cierre de compras: {{ dateTime(selected.booking_closes_at) }}. Los menores exentos se incluyen en turistas, pero no consumen cupo.</p>
          <Message v-if="selected.authorization_revoked_at && selected.status !== 'cancelled'" severity="warn" :closable="false">La autorización anterior se retiró el {{ dateTime(selected.authorization_revoked_at) }}: {{ selected.authorization_revocation_reason }}. Revisa los cupos antes de confirmar de nuevo.</Message>
          <Message v-if="selected.authorization_state === 'authorized'" severity="success" :closable="false">Autorizada por {{ selected.authorized_by_name }} el {{ dateTime(selected.authorized_at!) }}.</Message>
          <Message v-else-if="selected.can_authorize" severity="info" :closable="false">El mínimo está alcanzado. La salida requiere tu confirmación manual para quedar autorizada.</Message>
          <Message v-else-if="selected.authorization_state === 'minimum_review'" severity="warn" :closable="false">Cerraron las compras sin alcanzar el mínimo. Revisa los pagos pendientes antes de decidir la cancelación y los reembolsos.</Message>
          <div class="departure-controls"><Button v-if="selected.can_authorize" label="Confirmar salida" icon="pi pi-check-circle" :disabled="actionSaving" @click="askAction('authorize')" /><Button v-if="selected.can_cancel" label="Cancelar salida" icon="pi pi-ban" severity="danger" outlined :disabled="actionSaving" @click="askAction('cancel')" /></div>
          <Message v-if="selected.status === 'cancelled'" severity="warn" :closable="false">Salida cancelada<span v-if="selected.cancelled_at"> el {{ dateTime(selected.cancelled_at) }}</span>. {{ selected.cancellation_reason }}<br>Consulta las devoluciones en Reembolsos. Cada turista debe indicar su QR o cuenta bancaria; la agencia registra la devolución con su comprobante.</Message>
        </section>
        <section class="payment-detail">
          <div class="payment-tabs" role="tablist" aria-label="Información de la salida">
            <button v-for="(item, index) in tabs" :id="`tab-${item.id}`" :key="item.id" type="button" role="tab" :aria-selected="tab === item.id" aria-controls="departure-tabpanel" :tabindex="tab === item.id ? 0 : -1" :class="{ active: tab === item.id }" @click="selectTab(item.id)" @keydown="tabKey($event, index)"><i :class="item.icon" aria-hidden="true" />{{ item.label }}<span>{{ tabCount(item.id) }}</span></button>
          </div>
          <div id="departure-tabpanel" role="tabpanel" :aria-labelledby="`tab-${tab}`" tabindex="0">
            <div class="detail-tools"><div class="form-field"><label for="purchase-status">{{ tab === 'tourists' ? 'Compras en la lista' : 'Estado del pago' }}</label><Select input-id="purchase-status" v-model="purchaseStatusFilter" :options="purchaseStatusOptions" option-label="label" option-value="value" @change="applyPurchaseStatus" /></div><a v-if="tab === 'tourists' && selected.confirmed_purchases" class="print-roster" :href="rosterURL" target="_blank" rel="noopener noreferrer"><i class="pi pi-print" aria-hidden="true" /> Lista confirmada para imprimir</a></div>
            <Message v-if="tab === 'refunds'" severity="info" :closable="false">1. El turista indica su QR o cuenta bancaria. 2. La agencia realiza la devolución. 3. Abre Ver detalle, adjunta el comprobante y marca la devolución completada. Revisa el plazo de cada compra.</Message>
            <p v-if="tab === 'tourists'" class="muted roster-note">Cada fila corresponde al comprador responsable y su grupo. La lista imprimible incluye todas las compras con pago confirmado, aunque haya varias páginas.</p>
            <Message v-if="purchasesError" severity="error" :closable="false">{{ purchasesError }} <Button label="Reintentar" text @click="loadPurchases" /></Message>
            <UiLoadingState v-else-if="purchasesLoading" />
            <UiEmptyState v-else-if="!purchases.length" icon="pi pi-receipt" title="No hay compras en este estado" description="Selecciona otro filtro para consultar las compras de esta salida." />
            <PaymentsPurchaseTable v-else :purchases="purchases" :tab="tab" @open="openPurchase" />
            <nav v-if="purchaseTotal > purchaseLimit" class="payments-pagination" aria-label="Páginas de compras"><Button label="Anterior" icon="pi pi-chevron-left" severity="secondary" outlined :disabled="purchasePage <= 1 || purchasesLoading" @click="changePurchasePage(-1)" /><span>Página {{ purchasePage }} de {{ Math.ceil(purchaseTotal / purchaseLimit) }} · {{ purchaseTotal }} compras</span><Button label="Siguiente" icon="pi pi-chevron-right" icon-pos="right" severity="secondary" outlined :disabled="purchasePage * purchaseLimit >= purchaseTotal || purchasesLoading" @click="changePurchasePage(1)" /></nav>
          </div>
        </section>
      </template>
    </template>
    <Dialog v-model:visible="purchaseDialog" modal header="Detalle de compra" :style="{ width: 'min(58rem, 96vw)' }" :draggable="false"><PaymentsPurchaseReview v-if="activePurchase" :key="activePurchase.id" :purchase="activePurchase" @updated="purchaseUpdated" /></Dialog>
    <Dialog v-model:visible="actionDialog" modal :header="actionKind === 'authorize' ? 'Confirmar salida' : 'Cancelar salida y preparar reembolsos'" :style="{ width: 'min(32rem, 94vw)' }" :closable="!actionSaving" :draggable="false"><p v-if="actionKind === 'authorize'">Confirmarás que la agencia realizará esta salida. Se registrarán tu usuario y la fecha de autorización. Los pagos se revisan de forma independiente.</p><template v-else><p>Se cancelará esta salida, aunque ya esté autorizada. Las compras confirmadas, en revisión o en corrección pasarán a reembolso pendiente. La agencia realiza la devolución y adjunta el comprobante de cada compra.</p><div class="form-field"><label for="cancel-cause">Causa de cancelación</label><Select input-id="cancel-cause" v-model="cancellationCause" :options="cancellationOptions" option-label="label" option-value="value" /></div><div class="form-field"><label for="cancel-detail">Motivo para el turista *</label><Textarea id="cancel-detail" v-model.trim="cancellationDetail" rows="4" maxlength="900" placeholder="Explica el motivo de la cancelación" /><small>Entre 5 y 900 caracteres. Se mostrará en las compras afectadas.</small></div></template><p v-if="selected"><strong>{{ selected.package_name }}</strong><br>{{ dateTime(selected.starts_at) }}</p><Message v-if="actionError" severity="error" :closable="false">{{ actionError }}</Message><template #footer><Button label="Volver" severity="secondary" outlined :disabled="actionSaving" @click="actionDialog = false" /><Button :label="actionKind === 'authorize' ? 'Autorizar salida' : 'Confirmar cancelación'" :severity="actionKind === 'authorize' ? 'primary' : 'danger'" :loading="actionSaving" @click="performAction" /></template></Dialog>
  </section>
</template>
<script setup lang="ts">
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import Message from 'primevue/message'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import type { PaymentView, CancellationCause } from '~/utils/payments'
import type { PaymentDeparture, PaymentsOverview } from '~/types/payments'
import type { Purchase, PurchasePage } from '~/types/purchase'
import { apiError } from '~/utils/api-error'
import { packageMoney } from '~/utils/packages'
import { authorizationLabel, authorizationSeverity, groupPaymentDepartures, paymentViews, normalizePaymentView, departureCancellationReason, paymentDateTime as dateTime } from '~/utils/payments'

definePageMeta({ layout: 'app', middleware: 'agency' })
useHead({ title: 'Gestionar pagos', meta: [{ name: 'robots', content: 'noindex, nofollow' }] })
const auth = useAuthStore(); const config = useRuntimeConfig(); const route = useRoute(); const router = useRouter()
const overview = ref<PaymentsOverview | null>(null); const overviewLoading = ref(true); const overviewError = ref(''); const page = ref(1)
const filters = reactive({ search: '', package_id: 0, state: normalizePaymentView(route.query.view), from: '', to: '' })
const groups = computed(() => groupPaymentDepartures(overview.value?.departures || []))
const packageOptions = computed(() => [{ id: 0, name: 'Todos los paquetes' }, ...(overview.value?.packages || [])])
const metrics = computed(() => overview.value?.summary || overview.value?.totals)
const cancellationCause = ref<CancellationCause>('external'); const cancellationDetail = ref('')
const cancellationOptions = [{ label: 'Falta de cupos', value: 'minimum' }, { label: 'Motivos externos', value: 'external' }]

const selectedId = computed(() => { const value = String(route.query.departure || ''); return /^[1-9]\d*$/.test(value) ? Number(value) : 0 })
const selected = ref<PaymentDeparture | null>(null); const detailLoading = ref(false); const detailError = ref('')
type DetailTab = 'payments' | 'tourists' | 'refunds'
const tabs: { id: DetailTab; label: string; icon: string }[] = [{ id: 'payments', label: 'Pagos', icon: 'pi pi-credit-card' }, { id: 'tourists', label: 'Turistas', icon: 'pi pi-users' }, { id: 'refunds', label: 'Reembolsos', icon: 'pi pi-wallet' }]
const tab = ref<DetailTab>('payments'); const purchaseStatusFilter = ref('all'); const purchasePage = ref(1); const purchaseLimit = 20
const purchases = ref<Purchase[]>([]); const purchaseTotal = ref(0); const purchasesLoading = ref(false); const purchasesError = ref('')
const purchaseDialog = ref(false); const activePurchase = ref<Purchase | null>(null)
const actionDialog = ref(false); const actionKind = ref<'authorize' | 'cancel'>('authorize'); const actionSaving = ref(false); const actionError = ref(''); const notice = ref('')
const purchaseStatusOptions = computed(() => tab.value === 'refunds' ? [{ label: 'Pendientes de devolución', value: 'refund_pending' }, { label: 'Devoluciones realizadas', value: 'refunded' }] : [{ label: 'Todas las compras', value: 'all' }, { label: 'Pago confirmado', value: 'confirmed' }, { label: 'Pago en revisión', value: 'payment_review' }, { label: 'Corrección solicitada', value: 'correction_requested' }, { label: 'Reembolso pendiente', value: 'refund_pending' }, { label: 'Reembolsadas', value: 'refunded' }, { label: 'Canceladas', value: 'cancelled' }])
const shortDate = (value: string) => new Intl.DateTimeFormat('es-BO', { day: 'numeric', month: 'long', year: 'numeric', timeZone: 'America/La_Paz' }).format(new Date(value))
const clock = (value: string) => new Intl.DateTimeFormat('es-BO', { hour: '2-digit', minute: '2-digit', hourCycle: 'h23', timeZone: 'America/La_Paz' }).format(new Date(value))
const rosterURL = computed(() => `${config.public.apiBase}/agency/payments/departures/${selectedId.value}/roster`)
let overviewSequence = 0; let detailSequence = 0; let purchasesSequence = 0; let mounted = false
const loadOverview = async () => {
  const request = ++overviewSequence; overviewLoading.value = true; overviewError.value = ''
  const query = new URLSearchParams({ page: String(page.value), limit: '10', state: filters.state })
  for (const key of ['search', 'from', 'to'] as const) if (filters[key]) query.set(key, filters[key])
  if (filters.package_id) query.set('package_id', String(filters.package_id))
  try { const response = await auth.request<PaymentsOverview>(`/agency/payments/departures?${query}`); if (request === overviewSequence) { overview.value = response.data || null; const lastPage = Math.max(1, Math.ceil((overview.value?.pagination.total || 0) / (overview.value?.pagination.limit || 10))); if (page.value > lastPage) { page.value = lastPage; await loadOverview() } } }
  catch (cause) { if (request === overviewSequence) overviewError.value = apiError(cause, 'No se pudieron cargar las salidas').message }
  finally { if (request === overviewSequence) overviewLoading.value = false }
}
const loadPurchases = async () => {
  const id = selectedId.value; if (!id) return
  const request = ++purchasesSequence; purchasesLoading.value = true; purchasesError.value = ''
  const query = new URLSearchParams({ departure_id: String(id), page: String(purchasePage.value), limit: String(purchaseLimit) })
  if (purchaseStatusFilter.value !== 'all') query.set('status', purchaseStatusFilter.value)
  try { const response = await auth.request<PurchasePage>(`/agency/purchases?${query}`); if (request === purchasesSequence && id === selectedId.value) { purchases.value = response.data?.purchases || []; purchaseTotal.value = response.data?.pagination.total || 0; const lastPage = Math.max(1, Math.ceil(purchaseTotal.value / purchaseLimit)); if (purchasePage.value > lastPage) { purchasePage.value = lastPage; await loadPurchases() } } }
  catch (cause) { if (request === purchasesSequence && id === selectedId.value) purchasesError.value = apiError(cause, 'No se pudieron cargar las compras').message }
  finally { if (request === purchasesSequence) purchasesLoading.value = false }
}
const loadDetail = async () => {
  const id = selectedId.value; const request = ++detailSequence; if (!id) return
  detailLoading.value = true; detailError.value = ''; selected.value = null
  try { const response = await auth.request<PaymentDeparture>(`/agency/payments/departures/${id}`); if (request === detailSequence && id === selectedId.value) { selected.value = response.data || null; if (!selected.value) detailError.value = 'La salida no está disponible.' } }
  catch (cause) { if (request === detailSequence && id === selectedId.value) detailError.value = apiError(cause, 'No se pudo cargar la salida').message }
  finally { if (request === detailSequence) detailLoading.value = false }
}
const setView = async (view: PaymentView) => {
  filters.state = view; page.value = 1
  await router.replace({ path: route.path, query: { ...route.query, view, departure: undefined } })
  await loadOverview()
}
watch(() => route.query.view, (value) => {
  const view = normalizePaymentView(value)
  if (view !== filters.state) { filters.state = view; page.value = 1; if (mounted) void loadOverview() }
})
const applyFilters = () => { page.value = 1; return loadOverview() }
const clearFilters = () => { Object.assign(filters, { search: '', package_id: 0, from: '', to: '' }); return setView('authorized') }
const changeOverviewPage = (delta: number) => { page.value += delta; void loadOverview() }
const changePurchasePage = (delta: number) => { purchasePage.value += delta; void loadPurchases() }
const applyPurchaseStatus = () => { purchasePage.value = 1; void loadPurchases() }
const openDeparture = (id: number) => router.push({ path: route.path, query: { ...route.query, departure: String(id) } })
const closeDeparture = () => router.push({ path: route.path, query: { ...route.query, departure: undefined } })
const selectTab = (id: DetailTab) => { if (tab.value === id) return; tab.value = id; purchaseStatusFilter.value = id === 'tourists' ? 'confirmed' : id === 'refunds' ? 'refund_pending' : 'all'; applyPurchaseStatus() }
const tabKey = (event: KeyboardEvent, index: number) => {
  let next = index
  if (event.key === 'ArrowRight') next = (index + 1) % tabs.length
  else if (event.key === 'ArrowLeft') next = (index + tabs.length - 1) % tabs.length
  else if (event.key === 'Home') next = 0
  else if (event.key === 'End') next = tabs.length - 1
  else return
  event.preventDefault(); selectTab(tabs[next]!.id); (event.currentTarget as HTMLElement).parentElement?.querySelectorAll<HTMLButtonElement>('[role="tab"]')[next]?.focus()
}
const tabCount = (id: DetailTab) => !selected.value ? 0 : id === 'payments' ? selected.value.total_purchases : id === 'tourists' ? selected.value.confirmed_travelers : selected.value.refunds_pending
const openPurchase = (purchase: Purchase) => { activePurchase.value = purchase; purchaseDialog.value = true }
const purchaseUpdated = async () => { purchaseDialog.value = false; activePurchase.value = null; notice.value = 'La compra se actualizó correctamente.'; await refresh() }
const askAction = (kind: 'authorize' | 'cancel') => {
  actionKind.value = kind; actionError.value = ''; actionDialog.value = true
  cancellationCause.value = selected.value?.authorization_state === 'minimum_review' ? 'minimum' : 'external'
  cancellationDetail.value = ''
}
const performAction = async () => {
  const departure = selected.value; if (!departure || actionSaving.value) return
  if (actionKind.value === 'cancel' && cancellationDetail.value.trim().length < 5) { actionError.value = 'Explica el motivo con al menos cinco caracteres.'; return }
  actionSaving.value = true; actionError.value = ''
  try {
    const path = actionKind.value === 'authorize' ? `/agency/payments/departures/${departure.id}/authorize` : `/agency/packages/${departure.package_id}/departures/${departure.id}/cancel`
    await auth.request(path, { method: 'POST', body: { version: departure.version, ...(actionKind.value === 'cancel' ? { reason: departureCancellationReason(cancellationCause.value, cancellationDetail.value) } : {}) } })
    actionDialog.value = false; notice.value = actionKind.value === 'authorize' ? 'Salida autorizada. Se registraron el responsable y la fecha.' : 'Salida cancelada. Los reembolsos quedaron pendientes de gestionar.'
    await refresh()
    if (actionKind.value === 'cancel' && selectedId.value === departure.id) { tab.value = 'refunds'; purchaseStatusFilter.value = 'refund_pending'; applyPurchaseStatus() }
  } catch (cause) { actionError.value = apiError(cause, 'No se pudo registrar la decisión. Recarga para consultar el estado actual.').message }
  finally { actionSaving.value = false }
}
const refresh = async () => { await Promise.all([loadOverview(), ...(selectedId.value ? [loadDetail(), loadPurchases()] : [])]) }
watch(selectedId, () => {
  ++detailSequence; ++purchasesSequence; selected.value = null; purchases.value = []; purchaseDialog.value = false; actionDialog.value = false; activePurchase.value = null; notice.value = ''; tab.value = 'payments'; purchaseStatusFilter.value = 'all'; purchasePage.value = 1
  if (mounted && selectedId.value) void Promise.all([loadDetail(), loadPurchases()])
})
onMounted(() => { mounted = true; void refresh() })
onBeforeUnmount(() => { ++overviewSequence; ++detailSequence; ++purchasesSequence })
</script>
<style scoped>
.payments-page{min-width:0;width:100%}
.payment-filters>.form-field{min-width:0}
.payment-filters input,.payment-filters :deep(.p-select){min-width:0;max-width:100%;width:100%;box-sizing:border-box}
.payment-views{display:flex;flex-wrap:wrap;gap:.6rem}
.payment-views button{display:flex;align-items:center;gap:.6rem;min-height:46px;padding:.7rem 1rem;border:1px solid var(--color-border);border-radius:var(--radius-sm);background:var(--color-surface);color:var(--color-text-muted);font:inherit;cursor:pointer}
.payment-views button.active{border-color:var(--color-primary);background:var(--color-primary-soft);color:var(--color-primary)}
.payment-views button:focus-visible{outline:2px solid var(--color-primary);outline-offset:3px}
.payment-views button span{font-size:.8rem;padding:.1rem .4rem;border-radius:var(--radius-sm);background:var(--color-background)}
@media(max-width:760px){.payment-views{display:grid;grid-template-columns:repeat(2,minmax(0,1fr))}.payment-views button{justify-content:center;font-size:.85rem}}

.payments-metrics :deep(.dashboard-metric){flex-direction:column;min-height:10rem;padding:1rem;gap:.7rem}
.payments-metrics :deep(.metric-icon){width:2rem;height:2rem;font-size:1rem}
.payments-metrics :deep(.metric-copy){width:100%;gap:.4rem}
.payments-metrics :deep(.metric-copy strong){font-size:clamp(1.4rem,2vw,1.7rem);line-height:1.2;overflow-wrap:anywhere}

.payments-page{gap:1.4rem}.payments-heading,.package-group-heading,.departure-overview-title,.departure-controls,.detail-tools,.payments-pagination{display:flex;align-items:center;justify-content:space-between;gap:1rem}.payments-heading h1{margin:.3rem 0;font-size:clamp(2rem,4vw,3.2rem)}.payments-heading p,.departure-overview p{margin:.4rem 0;color:var(--color-text-muted)}.payments-metrics{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:1rem}.payment-filters{display:grid;grid-template-columns:1.2fr 1.2fr .8fr .8fr;gap:1rem;padding:1.25rem;background:var(--color-surface);border:1px solid var(--color-border);border-radius:var(--radius-md)}.filter-actions{grid-column:1/-1;display:flex;gap:.5rem}.payment-filters input[type=date]{width:100%;min-height:44px;padding:.7rem;border:1px solid var(--color-border);border-radius:var(--radius-sm);background:var(--color-background);color:var(--color-text);font:inherit}.muted{color:var(--color-text-muted);font-size:.9rem;line-height:1.6}.filter-note{margin:0}.package-groups{display:grid;gap:1.35rem}.package-group,.departure-overview,.payment-detail{background:var(--color-surface);border:1px solid var(--color-border);border-radius:var(--radius-md);overflow:hidden}.package-group-heading{padding:1.2rem 1.4rem;background:var(--color-surface-elevated)}.package-group-heading h2{margin:.2rem 0;font-size:1.2rem}.package-group-heading small{color:var(--color-text-muted)}.package-group-heading>div:nth-child(2){flex:1}.package-group-heading>a{font-size:.85rem;white-space:nowrap}.package-icon{display:grid;place-items:center;width:42px;height:42px;border-radius:var(--radius-sm);background:var(--color-primary-soft);color:var(--color-primary);flex-shrink:0}.departure-row{display:grid;grid-template-columns:1.15fr 1.1fr 1.05fr auto;gap:1.3rem;align-items:center;padding:1.2rem 1.4rem}.departure-row+.departure-row{border-top:1px solid var(--color-border)}.departure-date{display:flex;align-items:center;gap:.8rem}.departure-date>i{color:var(--color-text-muted)}.departure-date strong,.departure-date span,.departure-capacity span,.departure-capacity small,.departure-alerts small{display:block}.departure-date span,.departure-capacity small,.departure-alerts small{font-size:.8rem;color:var(--color-text-muted);margin-top:.3rem}.departure-capacity span{font-size:.85rem}.departure-capacity strong{font-size:1.1rem}.departure-capacity progress{width:100%;height:6px;margin:.65rem 0 .15rem;accent-color:var(--color-primary)}.departure-alerts{display:grid;justify-items:start;gap:.2rem}.overdue{color:var(--color-danger)!important}.payments-pagination{justify-content:center;flex-wrap:wrap;font-size:.85rem;color:var(--color-text-muted);margin:1rem 0}.back-to-departures{align-self:start}.departure-overview{padding:1.4rem;display:grid;gap:1rem}.departure-overview h2{font-size:clamp(1.4rem,3vw,2rem);margin:.35rem 0}.departure-overview-title{align-items:start}.departure-numbers{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:1rem;padding:1rem 0;border-block:1px solid var(--color-border)}.departure-numbers>div{display:grid;gap:.4rem}.departure-numbers strong{font-size:1.5rem}.departure-numbers small,.departure-numbers span{font-size:.8rem;color:var(--color-text-muted)}.departure-controls{justify-content:flex-start;flex-wrap:wrap}.departure-controls .panel-action{margin-left:auto}.payment-tabs{display:flex;gap:.5rem;padding:1rem;border-bottom:1px solid var(--color-border);background:var(--color-surface-elevated)}.payment-tabs button{display:flex;align-items:center;justify-content:center;gap:.6rem;min-height:46px;padding:.7rem 1rem;border:1px solid transparent;border-radius:var(--radius-sm);background:transparent;color:var(--color-text-muted);font:inherit;font-weight:650;cursor:pointer}.payment-tabs button.active{background:var(--color-primary-soft);color:var(--color-primary);border-color:var(--color-primary-muted)}.payment-tabs button span{font-size:.75rem;padding:.1rem .4rem;border-radius:var(--radius-sm);background:var(--color-background)}.payment-tabs button:focus-visible,.print-roster:focus-visible{outline:2px solid var(--color-primary);outline-offset:3px}#departure-tabpanel{padding:1.3rem}.detail-tools{align-items:end;flex-wrap:wrap;margin-bottom:1.2rem}.detail-tools>.form-field{min-width:15rem}.print-roster{display:inline-flex;align-items:center;gap:.6rem;min-height:44px;padding:.7rem .9rem;border:1px solid var(--color-border);border-radius:var(--radius-sm);text-decoration:none}.roster-note{max-width:65rem}
@media(min-width:1400px){.payments-page{max-width:1500px;margin-inline:auto;width:100%}}
@media(max-width:1150px){.payment-filters{grid-template-columns:repeat(3,minmax(0,1fr))}.departure-row{grid-template-columns:1fr 1fr}.departure-alerts{align-self:start}.departure-row>.p-button{justify-self:end}.payments-metrics{grid-template-columns:repeat(2,minmax(0,1fr))}}
@media(max-width:760px){.payments-heading{align-items:start;flex-direction:column}.payments-heading>.p-button{font-size:.8rem}.payment-filters{grid-template-columns:repeat(2,minmax(0,1fr));padding:1rem}.payment-filters>.form-field:first-child{grid-column:1/-1}.payments-metrics{gap:.7rem}.package-group-heading{padding:1rem;flex-wrap:wrap}.package-group-heading>a{margin-left:3.2rem}.departure-row{grid-template-columns:1fr;padding:1rem;gap:1rem}.departure-row>.p-button{justify-self:stretch}.departure-overview{padding:1rem}.departure-overview-title{flex-direction:column}.departure-numbers{grid-template-columns:repeat(2,minmax(0,1fr))}.departure-controls .panel-action{margin-left:0}.payment-tabs{padding:.7rem;gap:.2rem}.payment-tabs button{flex:1;padding:.6rem .4rem;font-size:.8rem;gap:.35rem}.payment-tabs button>i{display:none}#departure-tabpanel{padding:1rem}.payment-tabs button span{font-size:.65rem}.detail-tools,.detail-tools>.form-field,.print-roster{width:100%}.print-roster{justify-content:center}.payments-pagination span{font-size:.8rem}}
@media(max-width:380px){.payments-metrics{grid-template-columns:1fr}.payment-filters{grid-template-columns:1fr}.payments-heading{flex-direction:column}}
</style>
