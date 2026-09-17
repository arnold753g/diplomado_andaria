<template>
  <div class="location-widget stack">
    <p v-if="editable" class="muted">Haz clic en el mapa para ubicar la atracción o arrastra el marcador. También puedes mover el mapa y elegir su centro.</p>
    <p v-if="loading" role="status">Cargando mapa…</p>
    <Message v-if="error" severity="warn" :closable="false">{{ error }}</Message>
    <div class="map-shell">
      <div ref="container" class="location-map" role="region" :aria-label="editable ? 'Mapa para seleccionar la ubicación de la atracción' : 'Mapa de ubicación de la atracción'" />
      <span class="map-style-badge" aria-hidden="true">ANDARIA MAP</span>
    </div>
    <div v-if="editable" class="cluster"><Button label="Usar centro del mapa" icon="pi pi-map-marker" outlined :disabled="disabled || loading || !ready" @click="useCenter" /><Button label="Quitar ubicación" text severity="secondary" :disabled="disabled || latitude === null || longitude === null" @click="emit('change', { latitude: null, longitude: null })" /></div>
    <small v-if="editable" role="status">{{ latitude !== null && longitude !== null ? `Ubicación seleccionada: ${latitude.toFixed(6)}, ${longitude.toFixed(6)}` : 'Todavía no seleccionaste una ubicación.' }}</small>
  </div>
</template>
<script setup lang="ts">
import Button from 'primevue/button'
import Message from 'primevue/message'
import type { Map as LeafletMap, Marker, LatLng, Layer, TileLayer } from 'leaflet'
import type { StyleSpecification } from 'maplibre-gl'
import mapLibreWorkerUrl from 'maplibre-gl/dist/maplibre-gl-worker.mjs?worker&url'
import { buildAndariaMapStyle, type MapStyle } from '~/utils/map-style'
const props = withDefaults(defineProps<{ latitude: number | null; longitude: number | null; editable?: boolean; disabled?: boolean }>(), { editable: false, disabled: false })
const emit = defineEmits<{ change: [point: { latitude: number | null; longitude: number | null }] }>()
const config = useRuntimeConfig(); const container = ref<HTMLElement | null>(null)
const loading = ref(true); const ready = ref(false); const error = ref('')
let map: LeafletMap | undefined; let marker: Marker | undefined; let leaflet: typeof import('leaflet') | undefined; let observer: ResizeObserver | undefined; let vectorLayer: Layer | undefined; let rasterLayer: TileLayer | undefined; let disposed = false
const hasPoint = () => props.latitude !== null && props.longitude !== null && Number.isFinite(props.latitude) && Number.isFinite(props.longitude)
const choose = (point: LatLng) => {
  if (!props.editable || props.disabled) return
  const wrapped = point.wrap()
  emit('change', { latitude: Number(Math.max(-90, Math.min(90, wrapped.lat)).toFixed(7)), longitude: Number(wrapped.lng.toFixed(7)) })
}
const syncMarker = () => {
  if (!map || !leaflet) return
  if (!hasPoint()) { marker?.remove(); marker = undefined; return }
  const point: [number, number] = [props.latitude!, props.longitude!]
  if (!marker) {
    marker = leaflet.marker(point, { draggable: props.editable && !props.disabled, title: 'Ubicación de la atracción', alt: 'Ubicación de la atracción', icon: leaflet.divIcon({
      className: 'andaria-map-marker',
      html: '<svg viewBox="0 0 36 46" aria-hidden="true" focusable="false"><path d="M18 1C8.6 1 1 8.6 1 18c0 12.2 17 27 17 27s17-14.8 17-27C35 8.6 27.4 1 18 1Z"/><circle cx="18" cy="18" r="6"/></svg>',
      iconSize: [36, 46],
      iconAnchor: [18, 45]
    }) }).addTo(map)
    marker.on('dragend', () => { if (marker) choose(marker.getLatLng()) })
  } else marker.setLatLng(point)
  if (props.editable && !props.disabled) marker.dragging?.enable(); else marker.dragging?.disable()
  if (!map.getBounds().contains(point)) map.panTo(point)
}
const useCenter = () => { if (map) choose(map.getCenter()) }
watch(() => [props.latitude, props.longitude, props.disabled], syncMarker)
onMounted(async () => {
  try {
    leaflet = await import('leaflet')
    if (disposed || !container.value) return
    map = leaflet.map(container.value, { scrollWheelZoom: false, worldCopyJump: true, minZoom: 2, maxBounds: [[-85, -Infinity], [85, Infinity]], maxBoundsViscosity: 1 }).setView(hasPoint() ? [props.latitude!, props.longitude!] : [-21.5355, -64.7296], hasPoint() ? 15 : 12)
    rasterLayer = leaflet.tileLayer(config.public.mapTileUrl, { maxZoom: 19, attribution: config.public.mapAttribution, referrerPolicy: 'strict-origin-when-cross-origin' }).on('tileerror', () => { if (!vectorLayer) error.value = 'No se pudieron cargar algunas partes del mapa. Revisa tu conexión; puedes conservar o introducir las coordenadas.' }).addTo(map)
    map.on('click', event => choose(event.latlng))
    syncMarker(); ready.value = true
    observer = new ResizeObserver(() => map?.invalidateSize({ pan: false })); observer.observe(container.value)
    try {
      const response = await fetch(config.public.mapStyleUrl)
      if (!response.ok) throw new Error(`Map style ${response.status}`)
      const style = buildAndariaMapStyle(await response.json() as MapStyle)
      const [maplibre, { maplibreGL }] = await Promise.all([import('maplibre-gl'), import('@maplibre/maplibre-gl-leaflet')])
      maplibre.setWorkerUrl(mapLibreWorkerUrl)
      if (disposed || !map) return
      const layer = maplibreGL({ style: style as StyleSpecification, attributionControl: false }).addTo(map)
      vectorLayer = layer
      const vectorMap = layer.getMaplibreMap()
      const showVectorStyle = () => {
        if (disposed || !map) return
        rasterLayer?.remove(); rasterLayer = undefined
        map.attributionControl?.addAttribution(config.public.mapVectorAttribution)
      }
      if (vectorMap.loaded()) showVectorStyle(); else vectorMap.once('load', showVectorStyle)
    } catch {
      error.value = 'El estilo visual no pudo cargarse. Se muestra el mapa de calles de respaldo.'
    }
  } catch { error.value = 'No se pudo abrir el mapa. Puedes introducir las coordenadas manualmente.' }
  finally { loading.value = false }
})
onBeforeUnmount(() => { disposed = true; observer?.disconnect(); map?.remove(); map = undefined })
</script>
<style scoped>
.map-shell { position:relative; overflow:hidden; border:1px solid #3e4814; border-radius:var(--radius-md); background:var(--color-surface-elevated); box-shadow:inset 0 0 0 1px rgba(186,254,6,.06), 0 18px 40px rgba(0,0,0,.24); }
.location-map { height:24rem; width:100%; z-index:0; isolation:isolate; background:#080c0f; }
.map-style-badge { position:absolute; z-index:500; top:.75rem; right:.75rem; padding:.35rem .55rem; border:1px solid rgba(186,254,6,.45); border-radius:999px; background:rgba(8,12,15,.84); color:#bafe06; font-size:.65rem; font-weight:800; letter-spacing:.12em; pointer-events:none; backdrop-filter:blur(6px); }
.location-map :deep(.andaria-map-marker) { background:transparent; border:0; filter:drop-shadow(0 4px 5px rgba(0,0,0,.5)); }
.location-map :deep(.andaria-map-marker svg) { display:block; width:100%; height:100%; overflow:visible; }
.location-map :deep(.andaria-map-marker path) { fill:#bafe06; stroke:#263d00; stroke-width:2; }
.location-map :deep(.andaria-map-marker circle) { fill:#102000; }
.location-map :deep(.leaflet-control-attribution) { background:rgba(8,12,15,.84); color:#b9baba; font-size:11px; backdrop-filter:blur(6px); }
.location-map :deep(.leaflet-control-attribution a) { color:#bafe06; }
.location-map :deep(.leaflet-control-zoom) { overflow:hidden; border:1px solid #303432; box-shadow:0 6px 16px #0006; }
.location-map :deep(.leaflet-control-zoom a) { border-color:#303432; color:#f8f9f2; background:#111517; }
.location-map :deep(.leaflet-control-zoom a:hover) { color:#080c0f; background:#bafe06; }
</style>
