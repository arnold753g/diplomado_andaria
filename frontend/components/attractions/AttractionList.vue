<template>
  <section>
    <div class="page-heading"><span class="auth-eyebrow">{{ admin ? 'Administración' : managed ? 'Tu espacio' : favorites ? 'Tu colección' : 'Explora Bolivia' }}</span><h1>{{ admin ? 'Atracciones turísticas' : managed ? 'Mis atracciones' : favorites ? 'Mis favoritos' : 'Lugares que invitan a descubrir' }}</h1><p>{{ admin ? 'Crea atracciones y asigna a sus encargados.' : managed ? 'Completa y publica las atracciones que tienes asignadas.' : favorites ? 'Aquí aparecen tus atracciones favoritas que siguen publicadas.' : 'Encuentra naturaleza, cultura y nuevas experiencias. Empezamos en Tarija.' }}</p></div>
    <form class="toolbar" @submit.prevent="searchItems">
      <div class="form-field"><label for="attraction-search">Nombre o localidad</label><InputText id="attraction-search" v-model.trim="search" maxlength="160" placeholder="¿Qué lugar buscas?" /></div>
      <div class="form-field"><label for="attraction-department">Departamento</label><Select input-id="attraction-department" v-model="department" :options="options.departments" placeholder="Todos" show-clear /></div>
      <div class="form-field"><label for="attraction-category">Categoría</label><Select input-id="attraction-category" v-model="category" :options="options.categories" option-label="name" option-value="name" placeholder="Todas" show-clear @change="subcategory = null" /></div>
      <div class="form-field"><label for="attraction-subcategory">Subcategoría</label><Select input-id="attraction-subcategory" v-model="subcategory" :options="filteredSubcategories" option-label="label" option-value="id" placeholder="Todas" show-clear /></div>
      <div v-if="admin || managed" class="form-field"><label for="attraction-status">Estado</label><Select input-id="attraction-status" v-model="status" :options="statuses" option-label="label" option-value="value" placeholder="Todos" show-clear /></div>
      <Button type="submit" label="Buscar" icon="pi pi-search" :loading="loading" />
      <NuxtLink v-if="admin" class="create-link" to="/admin/attractions/new">Crear atracción <i class="pi pi-plus" aria-hidden="true" /></NuxtLink>
    </form>
    <Message v-if="error" severity="error" :closable="false">{{ error }} <Button label="Reintentar" text @click="load" /></Message>
    <UiLoadingState v-if="loading" />
    <UiEmptyState v-else-if="!error && !items.length" :title="favorites ? 'Todavía no hay favoritos disponibles' : 'No hay atracciones para mostrar'" :description="managed ? 'El administrador debe asignarte atracciones. También puedes revisar los filtros.' : favorites ? 'Explora el catálogo y guarda los lugares que te interesan.' : 'Prueba otros filtros o vuelve cuando se publiquen nuevos lugares.'" />
    <NuxtLink v-if="favorites && !loading && !items.length" to="/attractions">Explorar atracciones →</NuxtLink>
    <div v-if="!loading && !error && items.length" class="attraction-grid">
      <article v-for="item in items" :key="item.id" class="attraction-card">
        <NuxtLink :to="detailPath(item.id)" class="cover-link" :aria-label="`Ver ${item.name}`"><img v-if="item.photos?.length" :src="photoURL(item)" :alt="item.name" loading="lazy"><div v-else class="cover-placeholder"><i class="pi pi-image" aria-hidden="true" /><span>Fotografías por agregar</span></div></NuxtLink>
        <div class="attraction-card-body"><div class="cluster"><Tag :value="item.category" severity="secondary" /><Tag v-if="admin || managed" :value="item.status === 'inactive' ? 'Inactiva' : item.published ? 'Publicada' : 'Borrador'" :severity="item.status === 'inactive' ? 'danger' : item.published ? 'success' : 'warn'" /></div><h2><NuxtLink :to="detailPath(item.id)">{{ item.name }}</NuxtLink></h2><p class="muted"><i class="pi pi-map-marker" aria-hidden="true" /> {{ item.city }} · {{ item.department }}</p><div class="cluster subcategory-tags"><Tag v-for="classification in item.subcategories" :key="classification.subcategory_id" :value="classification.subcategory.name" severity="secondary" /></div><p class="excerpt">{{ item.description || 'Completa la información de esta atracción.' }}</p><p v-if="admin" class="muted">Encargado: {{ item.manager_name }}</p><p v-if="(admin || managed) && item.manager_status === 'inactive'" class="muted">Encargado desactivado</p><div class="card-bottom"><strong>{{ admissionLabel(item.admission_cents) }}</strong><NuxtLink :to="detailPath(item.id)">{{ admin || managed ? 'Gestionar' : 'Descubrir' }} →</NuxtLink></div><Button v-if="favorites" label="Quitar de favoritos" icon="pi pi-heart-fill" text :loading="removing === item.id" :disabled="removing !== null" @click="removeFavorite(item.id)" /></div>
      </article>
    </div>
    <Paginator v-if="total > limit" :rows="limit" :total-records="total" :first="(page - 1) * limit" @page="paginate" />
  </section>
</template>
<script setup lang="ts">
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import Paginator from 'primevue/paginator'
import type { Attraction, AttractionOptions, AttractionPage } from '~/types/attraction'
import type { ApiEnvelope } from '~/types/api'
import { admissionLabel } from '~/utils/attractions'
import { apiError } from '~/utils/api-error'
const props = defineProps<{ admin?: boolean; managed?: boolean; favorites?: boolean }>()
const auth = useAuthStore(); const config = useRuntimeConfig()
const items = ref<Attraction[]>([]); const options = ref<AttractionOptions>({ departments: [], categories: [] })
const loading = ref(true); const error = ref(''); const search = ref(''); const category = ref<string | null>(null); const department = ref<string | null>(null); const status = ref<string | null>(null)
const subcategory = ref<number | null>(null)
const filteredSubcategories = computed(() => options.value.categories.filter(c => !category.value || c.name === category.value).flatMap(c => c.subcategories.map(s => ({ id: s.id, label: `${s.name} · ${c.name}` }))))
const statuses = [{ label: 'Activa', value: 'active' }, { label: 'Inactiva', value: 'inactive' }]
const page = ref(1); const total = ref(0); const limit = 12; const removing = ref<number | null>(null)
const prefix = computed(() => props.admin ? '/admin/attractions' : props.managed ? '/managed-attractions' : '/attractions')
const detailPath = (id: number) => `${prefix.value}/${id}`
const photoURL = (a: Attraction) => `${config.public.apiBase}${prefix.value}/${a.id}/photos/${a.photos[0]?.id}`
let serial = 0
const load = async () => {
  const current = ++serial; loading.value = true; error.value = ''
  try {
    if (!options.value.categories.length) {
      const result = await $fetch<ApiEnvelope<AttractionOptions>>(`${config.public.apiBase}/attractions/options`)
      if (result.data) options.value = result.data
    }
    const query = new URLSearchParams({ page: String(page.value), limit: String(limit), search: search.value })
    if (category.value) query.set('category', category.value)
    if (department.value) query.set('department', department.value)
    if (subcategory.value) query.set('subcategory_id', String(subcategory.value))
    if (status.value && (props.admin || props.managed)) query.set('status', status.value)
    const path = `${props.favorites ? '/me/favorites' : prefix.value}?${query}`
    const result = props.admin || props.managed || props.favorites ? await auth.request<AttractionPage>(path) : await $fetch<ApiEnvelope<AttractionPage>>(`${config.public.apiBase}${path}`)
    if (current !== serial) return
    items.value = result.data?.attractions || []; total.value = result.data?.pagination.total || 0
  } catch (cause) { if (current === serial) error.value = apiError(cause).message }
  finally { if (current === serial) loading.value = false }
}
const searchItems = () => { page.value = 1; void load() }
const paginate = (e: { page: number }) => { page.value = e.page + 1; void load() }
const removeFavorite = async (id: number) => {
  removing.value = id; error.value = ''
  try { await auth.request(`/me/favorites/${id}`, { method: 'DELETE' }); if (items.value.length === 1 && page.value > 1) page.value--; await load() }
  catch (cause) { error.value = apiError(cause).message }
  finally { removing.value = null }
}
onMounted(load)
</script>
<style scoped>
.attraction-grid { display:grid; grid-template-columns:repeat(auto-fill,minmax(min(100%,19rem),1fr)); gap:1.5rem; }
.attraction-card { min-width:0; background:var(--color-surface); border:1px solid var(--color-border); border-radius:var(--radius-lg); overflow:hidden; }
.cover-link { display:block; aspect-ratio:16/10; background:var(--color-surface-elevated); }
.cover-link img { width:100%; height:100%; object-fit:cover; }
.cover-placeholder { height:100%; display:flex; flex-direction:column; align-items:center; justify-content:center; gap:.75rem; color:var(--color-text-muted); }
.cover-placeholder i { font-size:2rem; }
.attraction-card-body { padding:1.4rem; }
h2 { font-size:1.35rem; margin:1rem 0 .5rem; overflow-wrap:anywhere; }
h2 a { color:var(--color-text); text-decoration:none; }
.excerpt { color:var(--color-text-muted); display:-webkit-box; -webkit-line-clamp:3; -webkit-box-orient:vertical; overflow:hidden; }
.card-bottom { display:flex; flex-wrap:wrap; justify-content:space-between; gap:.75rem; margin-top:1.5rem; }
.create-link { padding:.8rem 1rem; border:1px solid var(--color-primary); border-radius:var(--radius-sm); }
</style>
