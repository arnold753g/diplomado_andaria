<template>
  <section>
    <div class="page-heading"><span class="auth-eyebrow">Administración</span><h1>Agencias</h1><p>Organiza las agencias y sus responsables.</p></div>
    <form class="toolbar" @submit.prevent="searchAgencies">
      <div class="form-field"><label for="agency-search">Buscar agencia o ciudad</label><InputText id="agency-search" v-model.trim="search" placeholder="Nombre o ciudad" maxlength="160" /></div>
      <div class="form-field"><label for="agency-status">Estado</label><Select input-id="agency-status" v-model="status" :options="statuses" option-label="label" option-value="value" placeholder="Todos" show-clear @change="searchAgencies" /></div>
      <Button type="submit" label="Buscar" icon="pi pi-search" :loading="loading" />
      <NuxtLink class="agency-create-link" to="/admin/agencies/new"><i class="pi pi-plus" aria-hidden="true" /> Crear agencia</NuxtLink>
    </form>
    <Message v-if="error" severity="error" :closable="false">{{ error }} <Button label="Reintentar" text @click="load" /></Message>
    <UiLoadingState v-if="loading" />
    <UiEmptyState v-else-if="!error && !agencies.length" title="Todavía no hay agencias" description="Crea una agencia o ajusta los filtros de búsqueda." />
    <div v-else-if="agencies.length" class="agency-grid">
      <Card v-for="agency in agencies" :key="agency.id">
        <template #title><div class="cluster"><i class="pi pi-building" aria-hidden="true" /><span>{{ agency.name }}</span></div></template>
        <template #subtitle>{{ agency.city }} · {{ agency.department }}</template>
        <template #content>
          <div class="cluster"><Tag :value="agency.status === 'active' ? 'Activa' : 'Inactiva'" :severity="agency.status === 'active' ? 'success' : 'danger'" /><Tag :value="agency.published ? 'Visible' : 'Oculta'" severity="secondary" /></div>
          <dl class="agency-summary"><dt>Encargado</dt><dd>{{ agency.manager_name }}</dd><dt>Edad mínima de pago</dt><dd>{{ agency.minimum_paying_age }} años</dd><dt>Medios de pago</dt><dd>{{ paymentLabel(agency) }}</dd></dl>
          <Message v-if="agency.manager_status === 'inactive'" severity="warn" :closable="false">El encargado está desactivado.</Message>
          <NuxtLink :to="`/admin/agencies/${agency.id}`">Gestionar agencia <i class="pi pi-arrow-right" aria-hidden="true" /></NuxtLink>
        </template>
      </Card>
    </div>
    <Paginator v-if="total > limit" :rows="limit" :total-records="total" :first="(page - 1) * limit" @page="paginate" />
  </section>
</template>
<script setup lang="ts">
import Button from 'primevue/button'
import Card from 'primevue/card'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import Paginator from 'primevue/paginator'
import type { Agency } from '~/types/agency'
import { apiError } from '~/utils/api-error'
definePageMeta({ layout: 'admin', middleware: 'admin' })
useHead({ title: 'Agencias' })
const auth = useAuthStore()
const agencies = ref<Agency[]>([])
const loading = ref(true)
const error = ref('')
const search = ref('')
const status = ref<string | null>(null)
const statuses = [{ label: 'Activa', value: 'active' }, { label: 'Inactiva', value: 'inactive' }]
const page = ref(1)
const total = ref(0)
const limit = 12
let requestNumber = 0
const load = async () => {
  const current = ++requestNumber
  loading.value = true; error.value = ''
  try {
    const query = new URLSearchParams({ page: String(page.value), limit: String(limit), search: search.value })
    if (status.value) query.set('status', status.value)
    const response = await auth.request<{ agencies: Agency[], pagination: { total: number } }>(`/admin/agencies?${query}`)
    if (current !== requestNumber) return
    agencies.value = response.data?.agencies || []; total.value = response.data?.pagination.total || 0
  } catch (cause) { if (current === requestNumber) error.value = apiError(cause).message }
  finally { if (current === requestNumber) loading.value = false }
}
const searchAgencies = () => { page.value = 1; void load() }
const paginate = (event: { page: number }) => { page.value = event.page + 1; void load() }
const paymentLabel = (agency: Agency) => [agency.accepts_qr ? 'QR' : '', agency.accepts_transfer ? 'Transferencia' : ''].filter(Boolean).join(' · ') || 'Por configurar'
onMounted(load)
</script>
<style scoped>
.agency-grid { display: grid; grid-template-columns: repeat(auto-fit,minmax(min(100%,20rem),1fr)); gap: var(--space-5); }
.agency-summary { display: grid; gap: .4rem; margin: var(--space-5) 0; }
.agency-summary dt { color: var(--color-text-muted); font-size: .85rem; }
.agency-summary dd { margin: 0 0 .6rem; overflow-wrap: anywhere; }
.agency-create-link { display: inline-flex; align-items: center; gap: .5rem; padding: .8rem 1rem; border: 1px solid var(--color-primary); border-radius: var(--radius-sm); }
</style>
