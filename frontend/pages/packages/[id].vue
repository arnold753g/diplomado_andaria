<template>
  <section class="package-detail stack">
    <NuxtLink to="/packages" class="back-link"><i class="pi pi-arrow-left" aria-hidden="true" /> Explorar paquetes</NuxtLink>
    <UiLoadingState v-if="loading" />
    <Message v-else-if="error" severity="error" :closable="false">{{ error }} <Button label="Reintentar" text @click="load" /></Message>
    <template v-else-if="item">
      <header class="detail-header">
        <div>
          <span class="auth-eyebrow">{{ item.agency_name }}</span>
          <h1>{{ item.name }}</h1>
          <div class="header-meta"><span><i class="pi pi-map-marker" aria-hidden="true" /> {{ item.agency_city }}, {{ item.agency_department }}</span><span><i class="pi pi-calendar" aria-hidden="true" /> {{ packageDuration(item) }}</span><span><i class="pi pi-chart-bar" aria-hidden="true" /> {{ packageDifficulty(item.difficulty) }}</span></div>
        </div>
        <div class="price-block"><span>Precio nacional desde</span><strong>{{ packageMoney(item.national_price_cents) }}</strong><small>por persona</small></div>
      </header>

      <div v-if="item.photos.length" class="gallery">
        <img class="hero-photo" :src="photoURL(selectedPhoto)" :alt="item.name">
        <div v-if="item.photos.length > 1" class="thumbnails" aria-label="Fotografías del paquete">
          <button v-for="(photo, index) in item.photos" :key="photo.id" type="button" :aria-label="`Ver fotografía ${index + 1}`" :aria-pressed="selectedPhoto === photo.id" @click="selectedPhoto = photo.id"><img :src="photoURL(photo.id)" :alt="`Fotografía ${index + 1} de ${item.name}`" loading="lazy"></button>
        </div>
      </div>

      <div class="detail-layout">
        <main class="detail-content stack">
          <Card><template #title>La experiencia</template><template #content><p class="preserve-text description">{{ item.description }}</p><div class="feature-row"><div><i class="pi pi-clock" aria-hidden="true" /><span>Duración</span><strong>{{ packageDuration(item) }}</strong></div><div><i class="pi pi-directions" aria-hidden="true" /><span>Frecuencia</span><strong>{{ packageFrequency(item.frequency_type) }}</strong></div><div><i class="pi pi-users" aria-hidden="true" /><span>Menores</span><strong>Pagan desde {{ item.minimum_paying_age }} años</strong></div></div></template></Card>

          <section class="stack" aria-labelledby="itinerary-title"><div class="section-heading"><span>Plan de viaje</span><h2 id="itinerary-title">Itinerario por días</h2></div><div class="itinerary-list"><article v-for="day in item.itinerary" :key="day.id" class="itinerary-item"><div class="day-number"><span>Día</span><strong>{{ day.day_number }}</strong></div><div><h3>{{ day.title }}</h3><p v-if="day.description" class="preserve-text">{{ day.description }}</p><ul v-if="day.activities.length" class="activity-list"><li v-for="activity in day.activities" :key="activity">{{ activity }}</li></ul><div v-if="day.attractions.length" class="cluster attractions"><NuxtLink v-for="link in day.attractions" :key="link.attraction_id" :to="`/attractions/${link.attraction_id}`"><i class="pi pi-map-marker" aria-hidden="true" /> {{ link.attraction?.name || 'Ver atracción' }}</NuxtLink></div></div></article></div></section>

          <div class="included-grid">
            <Card><template #title><span class="list-title"><i class="pi pi-check-circle" aria-hidden="true" /> Incluye</span></template><template #content><ul v-if="item.includes.length" class="plain-list positive"><li v-for="value in item.includes" :key="value">{{ value }}</li></ul><p v-else class="muted">La agencia no detalló elementos incluidos.</p></template></Card>
            <Card><template #title><span class="list-title"><i class="pi pi-times-circle" aria-hidden="true" /> No incluye</span></template><template #content><ul v-if="item.excludes.length" class="plain-list"><li v-for="value in item.excludes" :key="value">{{ value }}</li></ul><p v-else class="muted">No se indicaron exclusiones.</p></template></Card>
            <Card><template #title><span class="list-title"><i class="pi pi-shopping-bag" aria-hidden="true" /> Qué llevar</span></template><template #content><ul v-if="item.bring.length" class="plain-list"><li v-for="value in item.bring" :key="value">{{ value }}</li></ul><p v-else class="muted">No hay recomendaciones adicionales.</p></template></Card>
          </div>

          <Card><template #title>Tarifas y cancelación</template><template #content><dl class="policy-list"><dt>Tarifa nacional</dt><dd>{{ packageMoney(item.national_price_cents) }} por persona</dd><dt>Viajero extranjero</dt><dd>{{ item.foreign_surcharge_cents ? `${packageMoney(item.foreign_surcharge_cents)} adicionales por persona` : 'Sin recargo adicional' }}</dd><dt>Menores</dt><dd>Los menores de {{ item.minimum_paying_age }} años se registran, pero no pagan ni descuentan cupo. Desde esa edad se aplica la tarifa correspondiente.</dd><dt>Cancelación del turista</dt><dd>{{ item.cancellation_allowed ? `Permitida hasta ${item.cancellation_notice_hours} horas antes de la salida, con devolución del 100 % en esta versión.` : 'Este paquete no admite cancelación por parte del turista.' }}</dd></dl></template></Card>
        </main>

        <aside class="departures-panel stack" aria-labelledby="departures-title">
          <div><span class="auth-eyebrow">Disponibilidad real</span><h2 id="departures-title">Próximas salidas</h2><p>Los cupos y plazos se actualizan para cada fecha.</p></div>
          <Message v-if="!item.departures.length" severity="secondary" :closable="false">Esta agencia todavía no tiene próximas salidas disponibles.</Message>
          <article v-for="departure in item.departures" :key="departure.id" class="departure-card" :class="{ 'is-bookable': departure.bookable }">
            <div class="departure-heading"><strong>{{ packageDepartureDate(departure.starts_at) }}</strong><Tag :value="packageDepartureStatus(departure)" :severity="departureSeverity(departure)" /></div>
            <dl><dt>Cupos disponibles</dt><dd>{{ departure.available_capacity }} de {{ departure.max_capacity }}</dd><dt>Hora de encuentro</dt><dd>{{ timeLabel(departure.meeting_at) }}</dd><template v-if="departure.meeting_point"><dt>Punto de encuentro</dt><dd>{{ departure.meeting_point }}</dd></template><dt>Compra</dt><dd>{{ purchaseWindow(departure) }}</dd></dl>
            <Message v-if="departure.bookable && !departure.minimum_reached" severity="warn" :closable="false">Faltan {{ departure.remaining_for_minimum }} {{ departure.remaining_for_minimum === 1 ? 'persona' : 'personas' }} para confirmar el mínimo. Las compras se acumulan.</Message>
            <Message v-else-if="departure.minimum_reached" severity="success" :closable="false">La salida ya alcanzó el mínimo de participantes.</Message>
            <p v-if="departure.instructions" class="departure-instructions"><strong>Indicaciones:</strong> {{ departure.instructions }}</p>
            <Button v-if="departure.bookable && auth.user?.role === 'turista'" label="Comprar esta salida" icon="pi pi-shopping-cart" fluid @click="buy(departure.id)" />
            <NuxtLink v-else-if="departure.bookable && !auth.isAuthenticated" class="purchase-action" :to="loginForPurchase(departure.id)"><i class="pi pi-sign-in" aria-hidden="true" /> Ingresar para comprar</NuxtLink>
          </article>
          <Message v-if="item.departures.some(value => value.bookable)" severity="info" :closable="false">La compra utilizará una de estas salidas y validará nuevamente el cupo antes de aceptar el pago.</Message>
          <NuxtLink v-if="!auth.isAuthenticated" class="account-link" :to="`/login?redirect=${encodeURIComponent(route.fullPath)}`"><i class="pi pi-user" aria-hidden="true" /> Ingresar como turista</NuxtLink>
          <NuxtLink v-else-if="auth.user?.role === 'turista'" class="account-link" to="/app/profile"><i class="pi pi-id-card" aria-hidden="true" /> Revisar mis datos de viajero</NuxtLink>
        </aside>
      </div>
    </template>
  </section>
</template>

<script setup lang="ts">
import Button from 'primevue/button'
import Card from 'primevue/card'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import type { ApiEnvelope } from '~/types/api'
import type { PublicPackageDeparture, PublicPackageDetail } from '~/types/package'
import { apiError } from '~/utils/api-error'
import { packageDepartureDate, packageDepartureStatus, packageDifficulty, packageDuration, packageFrequency, packageMoney } from '~/utils/packages'

definePageMeta({ layout: 'catalog', key: route => route.fullPath, validate: route => /^[1-9]\d*$/.test(String(route.params.id)) })
const route = useRoute(); const config = useRuntimeConfig(); const auth = useAuthStore()
const item = ref<PublicPackageDetail | null>(null); const selectedPhoto = ref(0); const loading = ref(true); const error = ref('')
useHead(() => ({ title: item.value?.name || 'Paquete turístico' }))
const photoURL = (photoID: number) => `${config.public.apiBase}/packages/${route.params.id}/photos/${photoID}`
const timeLabel = (value: string) => new Intl.DateTimeFormat('es-BO', { hour: '2-digit', minute: '2-digit', timeZone: 'America/La_Paz' }).format(new Date(value))
const shortDateTime = (value: string) => new Intl.DateTimeFormat('es-BO', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit', timeZone: 'America/La_Paz' }).format(new Date(value))
const purchaseWindow = (departure: PublicPackageDeparture) => {
  const now = Date.now(); const opens = new Date(departure.booking_opens_at).getTime(); const closes = new Date(departure.booking_closes_at).getTime()
  if (now < opens) return `Abre el ${shortDateTime(departure.booking_opens_at)}`
  if (now >= closes) return 'Plazo de compra finalizado'
  return `Disponible hasta ${shortDateTime(departure.booking_closes_at)}`
}
const departureSeverity = (departure: PublicPackageDeparture) => departure.bookable ? departure.minimum_reached ? 'success' : 'warn' : 'secondary'
const purchasePath = (departureID: number) => `/packages/${route.params.id}/purchase?departure=${departureID}`
const loginForPurchase = (departureID: number) => `/login?redirect=${encodeURIComponent(purchasePath(departureID))}`
const buy = (departureID: number) => navigateTo(purchasePath(departureID))
const load = async () => {
  loading.value = true; error.value = ''; item.value = null
  try {
    const result = await $fetch<ApiEnvelope<PublicPackageDetail>>(`${config.public.apiBase}/packages/${route.params.id}`)
    if (!result.data) throw new Error('Missing package')
    item.value = result.data; selectedPhoto.value = result.data.photos[0]?.id || 0
    await auth.initialize()
  } catch (cause) { error.value = apiError(cause).message }
  finally { loading.value = false }
}
onMounted(load)
</script>

<style scoped>
.back-link { display:inline-flex; align-items:center; gap:.5rem; width:max-content; text-decoration:none; }
.detail-header { display:flex; align-items:end; justify-content:space-between; gap:2rem; margin:1rem 0; }.detail-header h1 { max-width:20ch; margin:.3rem 0 1rem; font-size:clamp(2.3rem,5vw,4.4rem); }.header-meta { display:flex; flex-wrap:wrap; gap:.75rem 1.4rem; color:var(--color-text-muted); }.header-meta span { display:flex; align-items:center; gap:.45rem; }
.price-block { flex:0 0 auto; display:grid; min-width:14rem; padding:1.3rem; border:1px solid rgba(186,254,6,.24); border-radius:var(--radius-lg); background:var(--color-primary-soft); }.price-block span,.price-block small { color:var(--color-text-muted); }.price-block strong { color:var(--color-primary); font-size:2rem; }
.gallery { margin-bottom:1rem; }.hero-photo { width:100%; max-height:38rem; aspect-ratio:16/8; object-fit:cover; border-radius:var(--radius-xl); }.thumbnails { display:flex; gap:.75rem; overflow-x:auto; padding:.75rem .2rem .2rem; }.thumbnails button { flex:0 0 auto; padding:3px; border:2px solid transparent; border-radius:var(--radius-sm); background:transparent; cursor:pointer; }.thumbnails button[aria-pressed=true] { border-color:var(--color-primary); }.thumbnails img { width:7rem; height:4.5rem; object-fit:cover; border-radius:.45rem; }
.detail-layout { display:grid; grid-template-columns:minmax(0,1.7fr) minmax(20rem,.8fr); gap:1.5rem; align-items:start; }.description { margin:0; font-size:1.05rem; line-height:1.85; }.preserve-text { white-space:pre-wrap; overflow-wrap:anywhere; }
.feature-row { display:grid; grid-template-columns:repeat(3,minmax(0,1fr)); gap:1rem; margin-top:1.5rem; }.feature-row > div { display:grid; gap:.15rem; padding:1rem; border:1px solid var(--color-border-soft); border-radius:var(--radius-md); background:var(--color-background-deep); }.feature-row i { margin-bottom:.35rem; color:var(--color-primary); font-size:1.25rem; }.feature-row span { color:var(--color-text-muted); font-size:.82rem; }
.section-heading span { color:var(--color-primary); font-size:.8rem; font-weight:700; letter-spacing:.1em; text-transform:uppercase; }.section-heading h2 { margin:.25rem 0 0; font-size:2rem; }.itinerary-list { position:relative; display:grid; gap:1rem; }.itinerary-item { display:grid; grid-template-columns:4rem 1fr; gap:1rem; padding:1.25rem; border:1px solid var(--color-border); border-radius:var(--radius-lg); background:var(--color-surface); }.day-number { display:grid; align-content:start; justify-items:center; padding:.7rem; border-radius:var(--radius-md); background:var(--color-primary-soft); color:var(--color-primary); }.day-number span { font-size:.72rem; font-weight:700; text-transform:uppercase; }.day-number strong { font-size:1.8rem; line-height:1; }.itinerary-item h3 { margin:.2rem 0 .75rem; font-size:1.35rem; }.activity-list { padding-left:1.25rem; color:var(--color-text-muted); }.attractions a { display:inline-flex; align-items:center; gap:.35rem; padding:.35rem .6rem; border:1px solid var(--color-border); border-radius:999px; text-decoration:none; font-size:.85rem; }
.included-grid { display:grid; grid-template-columns:repeat(3,minmax(0,1fr)); gap:1rem; }.list-title { display:flex; align-items:center; gap:.5rem; }.list-title i { color:var(--color-primary); }.plain-list { display:grid; gap:.6rem; margin:0; padding-left:1.2rem; color:var(--color-text-muted); }.plain-list.positive li::marker { color:var(--color-success); }
.policy-list { display:grid; grid-template-columns:minmax(10rem,.6fr) 1.4fr; gap:.8rem 1.5rem; margin:0; }.policy-list dt { color:var(--color-text-muted); }.policy-list dd { margin:0; }
.departures-panel { position:sticky; top:1rem; padding:1.25rem; border:1px solid var(--color-border); border-radius:var(--radius-xl); background:var(--color-surface); box-shadow:var(--shadow-sm); }.departures-panel h2 { margin:.2rem 0 .6rem; }.departures-panel > div > p { margin:0; color:var(--color-text-muted); }.departure-card { display:grid; gap:.85rem; padding:1rem; border:1px solid var(--color-border); border-radius:var(--radius-md); background:var(--color-background-deep); }.departure-card.is-bookable { border-color:rgba(186,254,6,.28); }.departure-heading { display:grid; gap:.6rem; }.departure-heading strong { text-transform:capitalize; }.departure-heading :deep(.p-tag) { width:max-content; }.departure-card dl { display:grid; grid-template-columns:1fr 1fr; gap:.45rem .75rem; margin:0; font-size:.9rem; }.departure-card dt { color:var(--color-text-muted); }.departure-card dd { margin:0; text-align:right; overflow-wrap:anywhere; }.departure-instructions { margin:0; color:var(--color-text-muted); font-size:.9rem; }.account-link,.purchase-action { display:flex; align-items:center; justify-content:center; gap:.5rem; padding:.8rem 1rem; border:1px solid var(--color-primary); border-radius:var(--radius-sm); font-weight:700; text-decoration:none; }.purchase-action { background:var(--color-primary); color:#111; }
@media(max-width:1000px) { .detail-layout { grid-template-columns:1fr; }.departures-panel { position:static; }.included-grid { grid-template-columns:1fr 1fr; } }
@media(max-width:700px) { .detail-header { align-items:start; flex-direction:column; }.price-block { width:100%; }.feature-row,.included-grid { grid-template-columns:1fr; }.policy-list { grid-template-columns:1fr; gap:.25rem; }.policy-list dd { margin-bottom:.8rem; }.itinerary-item { grid-template-columns:3.3rem 1fr; padding:1rem; }.hero-photo { aspect-ratio:4/3; }.departure-card dl { grid-template-columns:1fr; }.departure-card dd { margin-bottom:.5rem; text-align:left; } }
</style>
