<template>
  <section class="stack">
    <div class="catalog-hero">
      <div>
        <span class="auth-eyebrow">Experiencias organizadas</span>
        <h1>Viajes con fechas y cupos claros</h1>
        <p>Explora paquetes publicados por agencias de turismo y revisa su itinerario, precio y próximas salidas.</p>
      </div>
      <div class="hero-stat" aria-label="Cantidad de paquetes encontrados"><strong>{{ total }}</strong><span>{{ total === 1 ? 'paquete' : 'paquetes' }}</span></div>
    </div>

    <form class="catalog-filters" @submit.prevent="searchItems">
      <div class="form-field filter-search"><label for="catalog-package-search">Destino, paquete o agencia</label><InputText id="catalog-package-search" v-model.trim="search" maxlength="160" placeholder="Ej. ruta del vino" /></div>
      <div class="form-field"><label for="catalog-package-department">Departamento</label><Select input-id="catalog-package-department" v-model="department" :options="options.departments" placeholder="Todos" show-clear /></div>
      <div class="form-field"><label for="catalog-package-difficulty">Dificultad</label><Select input-id="catalog-package-difficulty" v-model="difficulty" :options="options.difficulties" option-label="label" option-value="value" placeholder="Todas" show-clear /></div>
      <div class="form-field"><label for="catalog-package-duration">Duración máxima</label><InputNumber input-id="catalog-package-duration" v-model="maxDuration" :min="1" :max="30" suffix=" días" show-buttons /></div>
      <div class="form-field"><label for="catalog-package-price">Precio máximo nacional</label><InputNumber input-id="catalog-package-price" v-model="maxPrice" :min="1" :max="1000000" prefix="Bs " :min-fraction-digits="0" :max-fraction-digits="2" /></div>
      <div class="form-field"><label for="catalog-package-date">Viajar desde</label><DatePicker input-id="catalog-package-date" v-model="dateFrom" date-format="dd/mm/yy" show-icon :min-date="today" show-button-bar /></div>
      <div class="form-field"><label for="catalog-package-sort">Ordenar por</label><Select input-id="catalog-package-sort" v-model="sort" :options="sortOptions" option-label="label" option-value="value" /></div>
      <div class="availability-filter"><Checkbox input-id="catalog-package-bookable" v-model="bookableOnly" binary /><label for="catalog-package-bookable">Solo con venta abierta</label></div>
      <div class="filter-actions"><Button type="submit" label="Buscar paquetes" icon="pi pi-search" :loading="loading" /><Button type="button" label="Limpiar" icon="pi pi-filter-slash" severity="secondary" text @click="resetFilters" /></div>
    </form>

    <Message v-if="error" severity="error" :closable="false">{{ error }} <Button label="Reintentar" text @click="load" /></Message>
    <UiLoadingState v-if="loading" />
    <UiEmptyState v-else-if="!error && !items.length" title="No encontramos paquetes" description="Prueba otros filtros o revisa nuevamente cuando las agencias publiquen nuevas fechas." />
    <div v-if="!loading && !error && items.length" class="catalog-grid">
      <article v-for="item in items" :key="item.id" class="catalog-card">
        <NuxtLink :to="`/packages/${item.id}`" class="package-cover" :aria-label="`Ver ${item.name}`">
          <img v-if="item.photos.length" :src="photoURL(item)" :alt="item.name" loading="lazy">
          <div v-else class="package-cover-placeholder"><i class="pi pi-images" aria-hidden="true" /><span>Fotografías por agregar</span></div>
          <Tag class="availability-tag" :value="item.bookable ? 'Venta abierta' : item.next_departure ? 'Próxima salida' : 'Sin fechas disponibles'" :severity="item.bookable ? 'success' : 'secondary'" />
        </NuxtLink>
        <div class="catalog-card-body">
          <div class="card-meta"><span><i class="pi pi-map-marker" aria-hidden="true" /> {{ item.agency_city }}, {{ item.agency_department }}</span><span>{{ packageDuration(item) }}</span></div>
          <h2><NuxtLink :to="`/packages/${item.id}`">{{ item.name }}</NuxtLink></h2>
          <p class="agency-name">Organiza {{ item.agency_name }}</p>
          <p class="excerpt">{{ item.description }}</p>
          <div v-if="item.next_departure" class="next-departure"><span>Próxima salida</span><strong>{{ packageDepartureShortDate(item.next_departure.starts_at) }}</strong><small>{{ item.next_departure.available_capacity }} cupos disponibles</small></div>
          <div class="card-footer"><div><small>Desde</small><strong>{{ packageMoney(item.national_price_cents) }}</strong><span>por persona</span></div><NuxtLink :to="`/packages/${item.id}`" class="detail-link">Ver paquete <i class="pi pi-arrow-right" aria-hidden="true" /></NuxtLink></div>
        </div>
      </article>
    </div>
    <Paginator v-if="total > limit" :rows="limit" :total-records="total" :first="(page - 1) * limit" @page="paginate" />
  </section>
</template>

<script setup lang="ts">
import Button from 'primevue/button'
import Checkbox from 'primevue/checkbox'
import DatePicker from 'primevue/datepicker'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Paginator from 'primevue/paginator'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import type { ApiEnvelope } from '~/types/api'
import type { PackageCatalogItem, PackageCatalogOptions, PackageCatalogPage } from '~/types/package'
import { apiError } from '~/utils/api-error'
import { packageDepartureShortDate, packageDuration, packageMoney } from '~/utils/packages'

const config = useRuntimeConfig()
const options = ref<PackageCatalogOptions>({ departments: [], difficulties: [] })
const items = ref<PackageCatalogItem[]>([])
const search = ref(''); const department = ref<string | null>(null); const difficulty = ref<string | null>(null)
const maxDuration = ref<number | null>(null); const maxPrice = ref<number | null>(null); const dateFrom = ref<Date | null>(null)
const bookableOnly = ref(false); const sort = ref('next_departure')
const loading = ref(true); const error = ref(''); const page = ref(1); const total = ref(0); const limit = 12
const today = new Date(); today.setHours(0, 0, 0, 0)
const sortOptions = [
  { label: 'Próxima salida', value: 'next_departure' },
  { label: 'Menor precio', value: 'price_asc' },
  { label: 'Mayor precio', value: 'price_desc' },
  { label: 'Más recientes', value: 'newest' }
]
const dateInput = (value: Date) => `${value.getFullYear()}-${String(value.getMonth() + 1).padStart(2, '0')}-${String(value.getDate()).padStart(2, '0')}`
const photoURL = (item: PackageCatalogItem) => `${config.public.apiBase}/packages/${item.id}/photos/${item.photos[0]?.id}`
let serial = 0
const load = async () => {
  const current = ++serial; loading.value = true; error.value = ''
  try {
    if (!options.value.departments.length) {
      const optionResult = await $fetch<ApiEnvelope<PackageCatalogOptions>>(`${config.public.apiBase}/packages/options`)
      if (optionResult.data) options.value = optionResult.data
    }
    const query = new URLSearchParams({ page: String(page.value), limit: String(limit), search: search.value, sort: sort.value })
    if (department.value) query.set('department', department.value)
    if (difficulty.value) query.set('difficulty', difficulty.value)
    if (maxDuration.value) query.set('max_duration', String(maxDuration.value))
    if (maxPrice.value) query.set('max_price_cents', String(Math.round(maxPrice.value * 100)))
    if (dateFrom.value) query.set('date_from', dateInput(dateFrom.value))
    if (bookableOnly.value) query.set('bookable', 'true')
    const result = await $fetch<ApiEnvelope<PackageCatalogPage>>(`${config.public.apiBase}/packages?${query}`)
    if (current !== serial) return
    items.value = result.data?.packages || []; total.value = result.data?.pagination.total || 0
  } catch (cause) { if (current === serial) error.value = apiError(cause).message }
  finally { if (current === serial) loading.value = false }
}
const searchItems = () => { page.value = 1; void load() }
const paginate = (event: { page: number }) => { page.value = event.page + 1; void load() }
const resetFilters = () => {
  search.value = ''; department.value = null; difficulty.value = null; maxDuration.value = null; maxPrice.value = null; dateFrom.value = null; bookableOnly.value = false; sort.value = 'next_departure'; page.value = 1; void load()
}
onMounted(load)
</script>

<style scoped>
.catalog-hero { display:flex; align-items:end; justify-content:space-between; gap:2rem; padding:clamp(1.5rem,4vw,3rem); border:1px solid var(--color-border-soft); border-radius:var(--radius-xl); background:radial-gradient(circle at 90% 10%,rgba(186,254,6,.16),transparent 18rem),linear-gradient(145deg,var(--color-surface),var(--color-background-deep)); }
.catalog-hero h1 { max-width:18ch; margin:.25rem 0 1rem; font-size:clamp(2.25rem,5vw,4.25rem); }
.catalog-hero p { max-width:60ch; margin:0; color:var(--color-text-muted); font-size:1.08rem; }
.hero-stat { flex:0 0 auto; min-width:8rem; padding:1.25rem; border:1px solid rgba(186,254,6,.2); border-radius:var(--radius-lg); background:var(--color-primary-soft); text-align:center; }
.hero-stat strong,.hero-stat span { display:block; }.hero-stat strong { color:var(--color-primary); font-size:2.25rem; line-height:1; }.hero-stat span { margin-top:.4rem; color:var(--color-text-muted); }
.catalog-filters { display:grid; grid-template-columns:minmax(16rem,2fr) repeat(3,minmax(10rem,1fr)); gap:1rem; align-items:end; padding:1.25rem; border:1px solid var(--color-border); border-radius:var(--radius-lg); background:var(--color-surface); }
.availability-filter { display:flex; align-items:center; gap:.65rem; min-height:2.75rem; font-weight:600; }
.filter-actions { display:flex; flex-wrap:wrap; justify-content:flex-end; gap:.5rem; grid-column:span 2; }
.catalog-grid { display:grid; grid-template-columns:repeat(auto-fill,minmax(min(100%,20rem),1fr)); gap:1.5rem; }
.catalog-card { min-width:0; overflow:hidden; border:1px solid var(--color-border); border-radius:var(--radius-lg); background:var(--color-surface); box-shadow:var(--shadow-sm); transition:transform var(--motion-fast),border-color var(--motion-fast); }
.catalog-card:hover { transform:translateY(-3px); border-color:rgba(186,254,6,.35); }
.package-cover { position:relative; display:block; aspect-ratio:16/10; overflow:hidden; background:var(--color-background-deep); }
.package-cover img { width:100%; height:100%; object-fit:cover; transition:transform var(--motion-standard); }.catalog-card:hover .package-cover img { transform:scale(1.025); }
.package-cover-placeholder { height:100%; display:grid; place-items:center; align-content:center; gap:.7rem; color:var(--color-text-muted); }.package-cover-placeholder i { font-size:2rem; }
.availability-tag { position:absolute; top:1rem; left:1rem; box-shadow:var(--shadow-sm); }
.catalog-card-body { display:grid; gap:.9rem; padding:1.35rem; }
.card-meta { display:flex; justify-content:space-between; gap:.75rem; color:var(--color-text-muted); font-size:.86rem; }.card-meta span { min-width:0; }
.catalog-card h2 { margin:0; font-size:1.45rem; }.catalog-card h2 a { color:var(--color-text); text-decoration:none; }
.agency-name { margin:0; color:var(--color-primary); font-size:.9rem; font-weight:600; }.excerpt { min-height:4.8em; margin:0; display:-webkit-box; overflow:hidden; -webkit-line-clamp:3; -webkit-box-orient:vertical; color:var(--color-text-muted); }
.next-departure { display:grid; gap:.15rem; padding:.85rem 1rem; border-left:3px solid var(--color-primary); border-radius:0 var(--radius-sm) var(--radius-sm) 0; background:var(--color-primary-muted); }.next-departure span,.next-departure small { color:var(--color-text-muted); }.next-departure strong { text-transform:capitalize; }
.card-footer { display:flex; align-items:end; justify-content:space-between; gap:1rem; padding-top:.5rem; border-top:1px solid var(--color-border-soft); }.card-footer > div { display:grid; }.card-footer small,.card-footer span { color:var(--color-text-muted); font-size:.78rem; }.card-footer strong { color:var(--color-primary); font-size:1.25rem; }.detail-link { font-weight:700; text-decoration:none; white-space:nowrap; }
@media(max-width:1000px) { .catalog-filters { grid-template-columns:repeat(2,minmax(0,1fr)); }.filter-actions { grid-column:auto; justify-content:flex-start; } }
@media(max-width:680px) { .catalog-hero { align-items:start; flex-direction:column; }.hero-stat { min-width:0; width:100%; }.catalog-filters { grid-template-columns:1fr; }.filter-actions { grid-column:auto; }.filter-actions > * { flex:1; }.card-meta { flex-direction:column; } }
</style>
