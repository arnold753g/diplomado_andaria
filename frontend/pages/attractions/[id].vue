<template>
  <section class="stack">
    <NuxtLink to="/attractions">← Explorar atracciones</NuxtLink>
    <UiLoadingState v-if="loading" />
    <Message v-else-if="error" severity="error" :closable="false">{{ error }} <Button label="Reintentar" text @click="load" /></Message>
    <article v-else-if="attraction" class="stack">
      <div class="detail-heading"><div><span class="auth-eyebrow">{{ attraction.category }}</span><h1>{{ attraction.name }}</h1><div class="cluster"><Tag v-for="s in attraction.subcategories" :key="s.subcategory_id" :value="s.subcategory.name" severity="secondary" /></div><p class="muted"><i class="pi pi-map-marker" aria-hidden="true" /> {{ attraction.city }} · {{ attraction.department }}</p></div>
        <ClientOnly><Button v-if="auth.user?.role === 'turista'" :label="favorite ? 'Guardado en favoritos' : 'Guardar en favoritos'" :icon="favorite ? 'pi pi-heart-fill' : 'pi pi-heart'" :outlined="!favorite" :loading="favoriteBusy" :disabled="!favoriteReady" @click="toggleFavorite" /><NuxtLink v-else-if="!auth.isAuthenticated" :to="`/login?redirect=${encodeURIComponent(route.fullPath)}`">Ingresa para guardar este lugar</NuxtLink></ClientOnly>
      </div>
      <Message v-if="favoriteError" severity="error" :closable="false">{{ favoriteError }} <Button v-if="!favoriteReady" label="Reintentar" text @click="loadFavorite" /></Message>
      <div v-if="attraction.photos.length" class="gallery"><img class="hero-photo" :src="photoURL(selectedPhoto)" :alt="attraction.name"><div v-if="attraction.photos.length > 1" class="thumbnails"><button v-for="(photo, i) in attraction.photos" :key="photo.id" type="button" :aria-label="`Ver fotografía ${i + 1}`" :aria-pressed="selectedPhoto === photo.id" @click="selectedPhoto = photo.id"><img :src="photoURL(photo.id)" :alt="`Fotografía ${i + 1}`" loading="lazy"></button></div></div>
      <div class="information-grid"><div class="stack"><Card><template #title>Conoce este lugar</template><template #content><p class="preserve-text">{{ attraction.description }}</p></template></Card><Card v-if="attraction.recommendations"><template #title>Antes de tu visita</template><template #content><p class="preserve-text">{{ attraction.recommendations }}</p></template></Card></div>
        <Card><template #title>Planifica tu visita</template><template #content><dl class="visit-facts"><dt>Entrada</dt><dd><strong>{{ admissionLabel(attraction.admission_cents) }}</strong><small>Precio informativo. Confirma las condiciones con el lugar.</small></dd><dt>Horarios</dt><dd class="preserve-text">{{ scheduleLabel(attraction) }}</dd><dt>Mejor época de visita</dt><dd>{{ seasonLabel(attraction) }}</dd><dt>Cómo llegar</dt><dd class="preserve-text">{{ attraction.address }}</dd><template v-if="attraction.phone"><dt>Contacto</dt><dd>{{ attraction.phone }}</dd></template></dl><ClientOnly v-if="mapURL"><LocationMap :latitude="attraction.latitude" :longitude="attraction.longitude" /></ClientOnly><a v-if="mapURL" class="directions-link" :href="mapURL" target="_blank" rel="noopener noreferrer" aria-label="Cómo llegar con Google Maps (abre en una pestaña nueva)"><i class="pi pi-directions" aria-hidden="true" /> Cómo llegar con Google Maps <span aria-hidden="true">↗</span></a><small v-if="mapURL" class="map-help">En el celular se abrirá la aplicación de mapas si está disponible.</small></template></Card>
      </div>
    </article>
  </section>
</template>
<script setup lang="ts">
import LocationMap from '~/components/attractions/LocationMap.client.vue'
import Tag from 'primevue/tag'
import Button from 'primevue/button'
import Card from 'primevue/card'
import Message from 'primevue/message'
import type { Attraction } from '~/types/attraction'
import type { ApiEnvelope } from '~/types/api'
import { apiError } from '~/utils/api-error'
import { admissionLabel, googleMapsDirectionsURL, scheduleLabel, seasonLabel } from '~/utils/attractions'
definePageMeta({ layout: 'catalog', key: route => route.fullPath, validate: route => /^[1-9]\d*$/.test(String(route.params.id)) })
const route = useRoute(); const config = useRuntimeConfig(); const auth = useAuthStore()
const attraction = ref<Attraction | null>(null); const loading = ref(true); const error = ref(''); const selectedPhoto = ref(0)
const favorite = ref(false); const favoriteBusy = ref(false); const favoriteReady = ref(false); const favoriteError = ref('')
useHead(() => ({ title: attraction.value?.name || 'Atracción turística' }))
const photoURL = (id: number) => `${config.public.apiBase}/attractions/${route.params.id}/photos/${id}`
const mapURL = computed(() => attraction.value?.latitude != null && attraction.value.longitude != null ? googleMapsDirectionsURL(attraction.value.latitude, attraction.value.longitude) : '')
const loadFavorite = async () => {
  if (auth.user?.role !== 'turista') return
  favoriteBusy.value = true; favoriteError.value = ''
  try { const result = await auth.request<{ favorite: boolean }>(`/me/favorites/${route.params.id}`); favorite.value = result.data?.favorite || false; favoriteReady.value = true }
  catch (cause) { favoriteError.value = apiError(cause).message }
  finally { favoriteBusy.value = false }
}
const load = async () => {
  loading.value = true; error.value = ''; attraction.value = null
  try { const result = await $fetch<ApiEnvelope<Attraction>>(`${config.public.apiBase}/attractions/${route.params.id}`); if (!result.data) throw new Error('Missing attraction'); attraction.value = result.data; selectedPhoto.value = result.data.photos[0]?.id || 0 }
  catch (cause) { error.value = apiError(cause).message }
  finally { loading.value = false }
  if (attraction.value) { await auth.initialize(); await loadFavorite() }
}
const toggleFavorite = async () => {
  if (favoriteBusy.value || !favoriteReady.value) return
  favoriteBusy.value = true; favoriteError.value = ''
  try { const result = await auth.request<{ favorite: boolean }>(`/me/favorites/${route.params.id}`, { method: favorite.value ? 'DELETE' : 'PUT' }); favorite.value = result.data?.favorite || false }
  catch (cause) { favoriteError.value = apiError(cause).message }
  finally { favoriteBusy.value = false }
}
onMounted(load)
</script>
<style scoped>
.detail-heading { display:flex; align-items:center; justify-content:space-between; flex-wrap:wrap; gap:1.5rem; margin:1rem 0; }
h1 { font-size:clamp(2rem,4vw,3.25rem); margin:.5rem 0; overflow-wrap:anywhere; }
.hero-photo { width:100%; max-height:34rem; aspect-ratio:16/9; object-fit:cover; border-radius:var(--radius-lg); }
.thumbnails { display:flex; flex-wrap:wrap; gap:.75rem; margin-top:.75rem; }
.thumbnails button { padding:3px; background:transparent; border:2px solid transparent; border-radius:var(--radius-sm); cursor:pointer; }
.thumbnails button[aria-pressed=true] { border-color:var(--color-primary); }
.thumbnails img { width:6rem; height:4rem; object-fit:cover; border-radius:var(--radius-sm); display:block; }
.information-grid { display:grid; grid-template-columns:minmax(0,1.7fr) minmax(0,1fr); gap:1.5rem; margin-top:1rem; align-items:start; }
.preserve-text { white-space:pre-wrap; overflow-wrap:anywhere; line-height:1.8; }
.visit-facts dt { color:var(--color-text-muted); margin-top:1rem; }
.visit-facts dd { margin:.5rem 0 1.5rem; }
.visit-facts small { display:block; color:var(--color-text-muted); margin-top:.5rem; }
.directions-link { display:flex; align-items:center; justify-content:center; gap:.6rem; width:100%; margin-top:1rem; padding:.8rem 1rem; border:1px solid var(--color-primary); border-radius:var(--radius-sm); color:var(--color-primary); font-weight:700; text-decoration:none; transition:background-color .2s ease,color .2s ease; }
.directions-link:hover { background:var(--color-primary); color:#080c0f; }
.map-help { display:block; margin-top:.6rem; color:var(--color-text-muted); text-align:center; }
@media(max-width:800px) { .information-grid { grid-template-columns:1fr; } }
</style>
