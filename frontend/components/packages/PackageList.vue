<template>
  <div class="stack">
    <div class="page-heading">
      <span class="auth-eyebrow">Gestión de agencia</span>
      <h1>Paquetes turísticos</h1>
      <p>Crea la oferta, organiza su itinerario y publícala cuando esté completa.</p>
    </div>
    <div class="toolbar">
      <div class="form-field package-search"><label for="package-search">Buscar</label><InputText id="package-search" v-model.trim="search" placeholder="Nombre del paquete" @keydown.enter="load(1)" /></div>
      <div class="cluster"><Button label="Buscar" icon="pi pi-search" outlined @click="load(1)" /><NuxtLink class="create-link" to="/agency/packages/new">Crear paquete <i class="pi pi-plus" aria-hidden="true" /></NuxtLink></div>
    </div>
    <UiLoadingState v-if="loading" />
    <Message v-else-if="error" severity="error" :closable="false">{{ error }} <Button label="Reintentar" text @click="load(page)" /></Message>
    <UiEmptyState v-else-if="!items.length" title="Todavía no hay paquetes" description="Crea el primer paquete turístico de tu agencia." />
    <div v-else class="package-grid">
      <article v-for="item in items" :key="item.id" class="package-card">
        <img v-if="item.photos.length" :src="photoURL(item)" :alt="item.name">
        <div v-else class="package-placeholder"><i class="pi pi-images" aria-hidden="true" /></div>
        <div class="stack package-card-content">
          <div class="cluster"><Tag :value="item.published ? 'Publicado' : 'Borrador'" :severity="item.published ? 'success' : 'warn'" /><Tag v-if="item.schedule" :value="frequencyLabel(item.schedule.frequency_type)" severity="info" /><span class="muted">{{ durationLabel(item) }}</span></div>
          <h2>{{ item.name }}</h2>
          <p class="muted">Desde {{ money(item.national_price_cents) }} por persona</p>
          <p>{{ item.description || 'Completa la descripción antes de publicar.' }}</p>
          <NuxtLink :to="`/agency/packages/${item.id}`">Gestionar paquete →</NuxtLink>
        </div>
      </article>
    </div>
    <Paginator v-if="total > limit" :first="(page - 1) * limit" :rows="limit" :total-records="total" @page="load($event.page + 1)" />
  </div>
</template>
<script setup lang="ts">
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Paginator from 'primevue/paginator'
import Tag from 'primevue/tag'
import type { PackagePage, TourPackage } from '~/types/package'
import { apiError } from '~/utils/api-error'
const auth = useAuthStore(); const config = useRuntimeConfig()
const items = ref<TourPackage[]>([]); const search = ref(''); const page = ref(1); const limit = 12; const total = ref(0); const loading = ref(true); const error = ref('')
const money = (cents: number) => `Bs ${(cents / 100).toFixed(2)}`
const durationLabel = (item: TourPackage) => `${item.duration_days} ${item.duration_days === 1 ? 'día' : 'días'}${item.duration_nights ? ` · ${item.duration_nights} ${item.duration_nights === 1 ? 'noche' : 'noches'}` : ''}`
const frequencyLabel = (frequency: string) => ({ single: 'Única', daily: 'Diaria', specific_weekdays: 'Días específicos' }[frequency] || frequency)
const photoURL = (item: TourPackage) => `${config.public.apiBase}/agency/packages/${item.id}/photos/${item.photos[0]?.id}`
const load = async (nextPage: number) => {
  loading.value = true; error.value = ''
  try {
    const result = await auth.request<PackagePage>(`/agency/packages?page=${nextPage}&limit=${limit}&search=${encodeURIComponent(search.value)}`)
    items.value = result.data?.packages || []; total.value = result.data?.pagination.total || 0; page.value = nextPage
  } catch (cause) { error.value = apiError(cause).message }
  finally { loading.value = false }
}
onMounted(() => load(1))
</script>
<style scoped>
.package-search { min-width:min(100%,22rem); }
.package-grid { display:grid; grid-template-columns:repeat(auto-fill,minmax(min(100%,18rem),1fr)); gap:1.25rem; }
.package-card { overflow:hidden; border:1px solid var(--color-border-soft); border-radius:var(--radius-lg); background:var(--color-surface); box-shadow:var(--shadow-sm); }
.package-card > img,.package-placeholder { width:100%; aspect-ratio:16/10; object-fit:cover; }
.package-placeholder { display:grid; place-items:center; background:var(--color-background-deep); color:var(--color-text-disabled); font-size:2rem; }
.package-card-content { padding:1.25rem; }
.package-card h2 { margin:0; font-size:1.25rem; }
.package-card p { margin:0; display:-webkit-box; overflow:hidden; -webkit-line-clamp:3; -webkit-box-orient:vertical; }
.create-link { padding:.8rem 1rem; border:1px solid var(--color-primary); border-radius:var(--radius-sm); }
</style>
