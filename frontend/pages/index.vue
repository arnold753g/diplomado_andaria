<template>
  <div class="home-shell">
    <section class="home-hero" aria-labelledby="home-title">
      <div class="hero-copy">
        <span class="auth-eyebrow">Explora Bolivia, comenzando por Tarija</span>
        <h1 id="home-title">Tu próxima experiencia empieza aquí</h1>
        <p>Descubre atracciones auténticas y paquetes con itinerarios, fechas, precios y cupos claros.</p>
        <div class="hero-actions">
          <NuxtLink class="primary-cta" to="/packages">Explorar paquetes <i class="pi pi-arrow-right" aria-hidden="true" /></NuxtLink>
          <NuxtLink class="secondary-cta" to="/attractions">Ver atracciones</NuxtLink>
        </div>
        <div class="trust-row" aria-label="Características de Andaria">
          <span><i class="pi pi-calendar" aria-hidden="true" /> Fechas reales</span>
          <span><i class="pi pi-users" aria-hidden="true" /> Cupos visibles</span>
          <span><i class="pi pi-shield" aria-hidden="true" /> Compras con seguimiento</span>
        </div>
      </div>
      <div class="hero-visual" aria-label="Experiencia destacada">
        <img v-if="heroPackage?.photos.length" :src="packagePhotoURL(heroPackage)" :alt="heroPackage.name">
        <div v-else class="hero-placeholder"><i class="pi pi-compass" aria-hidden="true" /><span>Nuevas rutas por descubrir</span></div>
        <div v-if="heroPackage" class="hero-card">
          <small>Próxima experiencia</small>
          <strong>{{ heroPackage.name }}</strong>
          <span v-if="heroPackage.next_departure">{{ packageDepartureShortDate(heroPackage.next_departure.starts_at) }} · {{ heroPackage.next_departure.available_capacity }} cupos</span>
        </div>
      </div>
    </section>

    <div class="home-stats" aria-label="Contenido disponible">
      <div><strong>{{ packageTotal }}</strong><span>paquetes con venta abierta</span></div>
      <div><strong>{{ attractionTotal }}</strong><span>atracciones publicadas</span></div>
      <div><strong>100 %</strong><span>devolución según política vigente</span></div>
    </div>

    <Message v-if="error" severity="warn" :closable="false">{{ error }} <Button label="Reintentar" text @click="load" /></Message>
    <UiLoadingState v-if="loading" />

    <template v-else>
      <section class="home-section" aria-labelledby="packages-title">
        <header class="section-heading"><div><span class="auth-eyebrow">Salidas disponibles</span><h2 id="packages-title">Viajes listos para reservar</h2><p>Propuestas de agencias locales ordenadas por su próxima fecha.</p></div><NuxtLink to="/packages">Ver todos <i class="pi pi-arrow-right" aria-hidden="true" /></NuxtLink></header>
        <div v-if="packages.length" class="feature-grid package-grid">
          <article v-for="item in packages" :key="item.id" class="feature-card">
            <NuxtLink :to="`/packages/${item.id}`" class="feature-image" :aria-label="`Ver paquete ${item.name}`"><img v-if="item.photos.length" :src="packagePhotoURL(item)" :alt="item.name" loading="lazy"><span v-else class="image-placeholder"><i class="pi pi-images" aria-hidden="true" /></span><span class="card-chip">{{ item.next_departure ? packageDepartureShortDate(item.next_departure.starts_at) : 'Próximamente' }}</span></NuxtLink>
            <div class="feature-body"><span class="feature-location"><i class="pi pi-map-marker" aria-hidden="true" /> {{ item.agency_city }}, {{ item.agency_department }}</span><h3><NuxtLink :to="`/packages/${item.id}`">{{ item.name }}</NuxtLink></h3><p>{{ item.description }}</p><div class="feature-footer"><strong>{{ packageMoney(item.national_price_cents) }}</strong><span v-if="item.next_departure">{{ item.next_departure.available_capacity }} cupos</span></div></div>
          </article>
        </div>
        <UiEmptyState v-else title="Pronto habrá nuevas salidas" description="Las agencias están preparando sus próximas experiencias." />
      </section>

      <section class="home-section" aria-labelledby="attractions-title">
        <header class="section-heading"><div><span class="auth-eyebrow">Lugares con identidad</span><h2 id="attractions-title">Atracciones para inspirar tu ruta</h2><p>Espacios naturales, culturales y turísticos publicados por sus encargados.</p></div><NuxtLink to="/attractions">Explorar mapa y catálogo <i class="pi pi-arrow-right" aria-hidden="true" /></NuxtLink></header>
        <div v-if="attractions.length" class="feature-grid attraction-grid">
          <article v-for="item in attractions" :key="item.id" class="attraction-card">
            <NuxtLink :to="`/attractions/${item.id}`" class="attraction-image"><img v-if="item.photos.length" :src="attractionPhotoURL(item)" :alt="item.name" loading="lazy"><span v-else class="image-placeholder"><i class="pi pi-map-marker" aria-hidden="true" /></span></NuxtLink>
            <div><span>{{ item.city }}, {{ item.department }}</span><h3><NuxtLink :to="`/attractions/${item.id}`">{{ item.name }}</NuxtLink></h3><p>{{ item.description || 'Descubre este lugar y planifica tu visita.' }}</p></div>
          </article>
        </div>
        <UiEmptyState v-else title="Estamos preparando el catálogo" description="Las primeras atracciones de Tarija aparecerán aquí." />
      </section>
    </template>

    <section class="how-section" aria-labelledby="how-title">
      <div class="section-heading"><div><span class="auth-eyebrow">Un proceso claro</span><h2 id="how-title">De la inspiración a la salida</h2></div></div>
      <ol class="steps-grid"><li><span>01</span><i class="pi pi-search" aria-hidden="true" /><h3>Explora</h3><p>Compara atracciones, itinerarios, precios y próximas fechas.</p></li><li><span>02</span><i class="pi pi-receipt" aria-hidden="true" /><h3>Compra</h3><p>Selecciona una salida, indica tu grupo y adjunta el comprobante.</p></li><li><span>03</span><i class="pi pi-check-circle" aria-hidden="true" /><h3>Da seguimiento</h3><p>Consulta la validación, el cupo mínimo y cualquier devolución desde tu cuenta.</p></li></ol>
    </section>

    <section class="home-cta">
      <div><span class="auth-eyebrow">Andaria</span><h2>Guarda tus lugares favoritos y organiza tu próxima experiencia</h2></div>
      <div class="hero-actions"><NuxtLink v-if="!auth.isAuthenticated" class="primary-cta" to="/register">Crear cuenta</NuxtLink><NuxtLink v-else class="primary-cta" :to="auth.isAdmin ? '/admin' : '/app'">Ir a mi panel</NuxtLink><NuxtLink class="secondary-cta" to="/packages">Explorar ahora</NuxtLink></div>
    </section>
  </div>
</template>

<script setup lang="ts">
import Button from 'primevue/button'
import Message from 'primevue/message'
import type { ApiEnvelope } from '~/types/api'
import type { Attraction, AttractionPage } from '~/types/attraction'
import type { PackageCatalogItem, PackageCatalogPage } from '~/types/package'
import { packageDepartureShortDate, packageMoney } from '~/utils/packages'

definePageMeta({ layout: 'catalog' })
useHead({ title: 'Explora Bolivia', meta: [{ name: 'description', content: 'Atracciones y paquetes turísticos con fechas, precios y cupos claros. Explora Bolivia comenzando por Tarija.' }] })
const config = useRuntimeConfig(); const auth = useAuthStore()
const packages = ref<PackageCatalogItem[]>([]); const attractions = ref<Attraction[]>([])
const packageTotal = ref(0); const attractionTotal = ref(0); const loading = ref(true); const error = ref('')
const heroPackage = computed(() => packages.value[0])
const packagePhotoURL = (item: PackageCatalogItem) => `${config.public.apiBase}/packages/${item.id}/photos/${item.photos[0]?.id}`
const attractionPhotoURL = (item: Attraction) => `${config.public.apiBase}/attractions/${item.id}/photos/${item.photos[0]?.id}`
const load = async () => {
  loading.value = true; error.value = ''
  const results = await Promise.allSettled([
    $fetch<ApiEnvelope<PackageCatalogPage>>(`${config.public.apiBase}/packages?limit=3&bookable=true&sort=next_departure`),
    $fetch<ApiEnvelope<AttractionPage>>(`${config.public.apiBase}/attractions?limit=3`)
  ])
  if (results[0].status === 'fulfilled') { packages.value = results[0].value.data?.packages || []; packageTotal.value = results[0].value.data?.pagination.total || 0 }
  if (results[1].status === 'fulfilled') { attractions.value = results[1].value.data?.attractions || []; attractionTotal.value = results[1].value.data?.pagination.total || 0 }
  if (results.some(result => result.status === 'rejected')) error.value = 'Parte del contenido no pudo cargarse. Puedes seguir explorando los catálogos.'
  loading.value = false
}
onMounted(load)
</script>

<style scoped>
.home-shell { display:grid; gap:clamp(4rem,8vw,7rem); padding-bottom:2rem; }.home-hero { min-height:min(46rem,calc(100dvh - 8rem)); display:grid; grid-template-columns:minmax(0,1.05fr) minmax(22rem,.95fr); gap:clamp(2rem,5vw,5rem); align-items:center; }.hero-copy h1 { max-width:11ch; margin:.45rem 0 1.25rem; font-size:clamp(3rem,7vw,6.4rem); line-height:.94; }.hero-copy > p { max-width:58ch; color:var(--color-text-muted); font-size:clamp(1.05rem,2vw,1.25rem); }.hero-actions { display:flex; flex-wrap:wrap; gap:.75rem; margin-top:1.75rem; }.primary-cta,.secondary-cta { min-height:3rem; display:inline-flex; align-items:center; justify-content:center; gap:.55rem; padding:.78rem 1.2rem; border-radius:var(--radius-sm); font-weight:700; text-decoration:none; }.primary-cta { background:var(--color-primary); color:var(--color-primary-contrast); }.primary-cta:hover { background:var(--color-primary-hover); color:var(--color-primary-contrast); }.secondary-cta { border:1px solid var(--color-border); color:var(--color-text); }.secondary-cta:hover { border-color:var(--color-primary); color:var(--color-primary); }.trust-row { display:flex; flex-wrap:wrap; gap:.75rem 1.4rem; margin-top:1.75rem; color:var(--color-text-muted); font-size:.9rem; }.trust-row span { display:flex; align-items:center; gap:.4rem; }.trust-row i { color:var(--color-primary); }.hero-visual { position:relative; min-height:34rem; overflow:hidden; border:1px solid var(--color-border-soft); border-radius:var(--radius-xl); background:radial-gradient(circle at 70% 20%,var(--color-primary-soft),transparent 16rem),var(--color-surface); box-shadow:var(--shadow-md); }.hero-visual > img { width:100%; height:100%; min-height:34rem; object-fit:cover; }.hero-visual::after { content:""; position:absolute; inset:0; background:linear-gradient(to top,rgba(1,1,0,.88),transparent 58%); pointer-events:none; }.hero-placeholder { min-height:34rem; display:grid; place-items:center; align-content:center; gap:1rem; color:var(--color-text-muted); }.hero-placeholder i { color:var(--color-primary); font-size:4rem; }.hero-card { position:absolute; z-index:1; left:1.4rem; right:1.4rem; bottom:1.4rem; display:grid; gap:.2rem; padding:1.15rem; border:1px solid rgba(248,249,242,.16); border-radius:var(--radius-md); background:rgba(8,12,15,.78); backdrop-filter:blur(12px); }.hero-card small,.hero-card span { color:var(--color-text-muted); }.hero-card strong { font-size:1.25rem; }.home-stats { display:grid; grid-template-columns:repeat(3,1fr); border-block:1px solid var(--color-border); }.home-stats div { display:grid; gap:.25rem; padding:1.4rem clamp(1rem,3vw,2rem); text-align:center; }.home-stats div+div { border-left:1px solid var(--color-border); }.home-stats strong { color:var(--color-primary); font-size:clamp(1.6rem,4vw,2.8rem); font-variant-numeric:tabular-nums; }.home-stats span { color:var(--color-text-muted); }.home-section { display:grid; gap:1.7rem; }.section-heading { display:flex; align-items:end; justify-content:space-between; gap:2rem; }.section-heading h2 { max-width:20ch; margin:.3rem 0 .6rem; font-size:clamp(2rem,4vw,3.4rem); }.section-heading p { margin:0; color:var(--color-text-muted); }.section-heading > a { flex:0 0 auto; font-weight:700; text-decoration:none; }.feature-grid { display:grid; grid-template-columns:repeat(3,minmax(0,1fr)); gap:1.25rem; }.feature-card,.attraction-card { overflow:hidden; border:1px solid var(--color-border); border-radius:var(--radius-lg); background:var(--color-surface); box-shadow:var(--shadow-sm); transition:transform var(--motion-fast),border-color var(--motion-fast); }.feature-card:hover,.attraction-card:hover { transform:translateY(-3px); border-color:rgba(186,254,6,.35); }.feature-image { position:relative; display:block; aspect-ratio:16/10; overflow:hidden; background:var(--color-background-deep); }.feature-image img,.attraction-image img { width:100%; height:100%; object-fit:cover; }.image-placeholder { width:100%; height:100%; display:grid; place-items:center; color:var(--color-primary); font-size:2rem; }.card-chip { position:absolute; left:.8rem; bottom:.8rem; padding:.35rem .65rem; border-radius:999px; background:rgba(8,12,15,.84); color:var(--color-text); font-size:.78rem; font-weight:700; }.feature-body { display:grid; gap:.65rem; padding:1.15rem; }.feature-location,.attraction-card > div > span { color:var(--color-text-muted); font-size:.84rem; }.feature-body h3,.attraction-card h3 { margin:0; font-size:1.25rem; }.feature-body h3 a,.attraction-card h3 a { color:var(--color-text); text-decoration:none; }.feature-body p,.attraction-card p { margin:0; color:var(--color-text-muted); display:-webkit-box; overflow:hidden; -webkit-line-clamp:3; -webkit-box-orient:vertical; }.feature-footer { display:flex; justify-content:space-between; gap:1rem; align-items:end; padding-top:.7rem; border-top:1px solid var(--color-border-soft); }.feature-footer strong { color:var(--color-primary); }.feature-footer span { color:var(--color-text-muted); font-size:.82rem; }.attraction-card { display:grid; grid-template-columns:minmax(8rem,.85fr) minmax(0,1.15fr); min-height:15rem; }.attraction-image { min-height:100%; background:var(--color-background-deep); }.attraction-card > div { display:grid; align-content:center; gap:.65rem; padding:1.2rem; }.how-section { padding:clamp(1.5rem,4vw,3rem); border:1px solid var(--color-border); border-radius:var(--radius-xl); background:linear-gradient(145deg,var(--color-surface),var(--color-background-deep)); }.steps-grid { display:grid; grid-template-columns:repeat(3,1fr); gap:1rem; padding:0; margin:1.5rem 0 0; list-style:none; }.steps-grid li { position:relative; min-height:14rem; padding:1.3rem; border:1px solid var(--color-border-soft); border-radius:var(--radius-lg); background:var(--color-surface-elevated); }.steps-grid li > span { position:absolute; top:1rem; right:1rem; color:var(--color-text-disabled); font-size:1.3rem; font-weight:800; }.steps-grid li > i { color:var(--color-primary); font-size:1.6rem; }.steps-grid h3 { margin:2rem 0 .6rem; }.steps-grid p { margin:0; color:var(--color-text-muted); }.home-cta { display:flex; align-items:center; justify-content:space-between; gap:2rem; padding:clamp(1.5rem,5vw,3.5rem); border:1px solid rgba(186,254,6,.24); border-radius:var(--radius-xl); background:radial-gradient(circle at 90%,var(--color-primary-soft),transparent 24rem),var(--color-surface); box-shadow:var(--shadow-glow); }.home-cta h2 { max-width:22ch; margin:.3rem 0 0; font-size:clamp(1.8rem,4vw,3rem); }.home-cta .hero-actions { flex:0 0 auto; margin:0; }
@media(max-width:1000px) { .home-hero { min-height:auto; grid-template-columns:1fr; }.hero-copy h1 { max-width:13ch; }.hero-visual { min-height:26rem; }.hero-visual > img,.hero-placeholder { min-height:26rem; }.feature-grid { grid-template-columns:repeat(2,minmax(0,1fr)); }.feature-grid > :last-child { display:none; }.home-cta { align-items:flex-start; flex-direction:column; }.home-cta .hero-actions { width:100%; } }
@media(max-width:680px) { .home-shell { gap:3.5rem; }.hero-copy h1 { font-size:clamp(2.65rem,14vw,4rem); }.hero-actions > * { width:100%; }.trust-row { display:grid; }.hero-visual,.hero-visual > img,.hero-placeholder { min-height:22rem; }.home-stats { grid-template-columns:1fr; }.home-stats div+div { border-top:1px solid var(--color-border); border-left:0; }.section-heading { align-items:flex-start; flex-direction:column; gap:.75rem; }.feature-grid { grid-template-columns:1fr; }.feature-grid > :last-child { display:block; }.attraction-card { grid-template-columns:1fr; }.attraction-image { aspect-ratio:16/9; }.steps-grid { grid-template-columns:1fr; }.steps-grid li { min-height:11rem; }.home-cta .hero-actions > * { width:100%; } }
@media(prefers-reduced-motion:reduce) { .feature-card,.attraction-card { transition:none; } }
</style>
