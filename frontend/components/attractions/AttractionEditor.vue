<template>
  <UiLoadingState v-if="loading" />
  <div v-else class="stack">
    <Message v-if="error" severity="error" :closable="false">{{ error }}<ul v-if="Object.keys(fields).length"><li v-for="(message, key) in fields" :key="key">{{ message }}</li></ul></Message>
    <Button v-if="!ready" label="Reintentar" @click="load" />
    <form v-if="ready" class="stack" @submit.prevent="save">
      <Message v-if="!admin && form.status === 'inactive'" severity="warn" :closable="false">Esta atracción está desactivada. Puedes consultarla; contacta al administrador para modificarla.</Message>
      <Message v-if="record?.manager_status === 'inactive'" severity="warn" :closable="false">El encargado está desactivado. Puedes reasignar la atracción.</Message>
      <fieldset :disabled="locked">
        <div class="editor-grid">
          <Card><template #title>Información del lugar</template><template #content><div class="stack">
            <div class="form-field"><label for="place-name">Nombre *</label><InputText id="place-name" v-model.trim="form.name" required minlength="2" maxlength="160" /></div>
            <div class="form-field"><h3>Clasificación *</h3><ClassificationSelect v-model="form.subcategory_ids" :categories="options.categories" :disabled="locked" /><Message v-if="record && !record.subcategories?.length" severity="info" :closable="false">Clasificación anterior: {{ record.category }}. Selecciona sus subcategorías para actualizarla al modelo de Andaria.</Message></div>
            <div class="form-field"><label for="place-description">Descripción</label><Textarea id="place-description" v-model.trim="form.description" rows="6" maxlength="6000" auto-resize /><small>Para publicar, escribe al menos 20 caracteres y agrega una fotografía.</small></div>
            <div class="stack visit-settings"><h3>Horarios y días de apertura</h3>
              <Message v-if="record?.opening_hours && form.schedule_mode === 'unspecified'" severity="info" :closable="false">Horario anterior: {{ record.opening_hours }}. Selecciona los días y horas para actualizarlo.</Message>
              <div class="form-field"><label for="place-schedule-mode">Horario de visita</label><Select input-id="place-schedule-mode" v-model="form.schedule_mode" :options="scheduleModes" option-label="label" option-value="value" :disabled="locked" @change="changeSchedule" /></div>
              <div v-if="form.schedule_mode !== 'unspecified'" class="stack"><div class="cluster"><Button label="Todos los días" text size="small" :disabled="locked" @click="form.opening_days = weekDays.map(d => d.id)" /><Button label="Limpiar días" text size="small" :disabled="locked" @click="form.opening_days = []" /></div><div class="day-options" role="group" aria-label="Días de apertura"><div v-for="day in weekDays" :key="day.id" class="cluster"><Checkbox :input-id="`place-day-${day.id}`" v-model="form.opening_days" :value="day.id" :disabled="locked" /><label :for="`place-day-${day.id}`">{{ day.name }}</label></div></div></div>
              <div v-if="form.schedule_mode === 'scheduled'" class="detail-grid"><div class="form-field"><label for="place-opening-hour">Apertura</label><TimeSelect id="place-opening" label="Apertura" v-model="form.opening_time" :disabled="locked" /></div><div class="form-field"><label for="place-closing-hour">Cierre</label><TimeSelect id="place-closing" label="Cierre" v-model="form.closing_time" :disabled="locked" /></div></div>
              <small v-if="form.schedule_mode !== 'unspecified'">Horario local de Bolivia. El mismo horario se aplica a los días seleccionados.</small><small v-if="form.schedule_mode === 'scheduled' && form.closing_time && form.opening_time && form.closing_time < form.opening_time">El cierre corresponde al día siguiente.</small>
            </div>
            <div class="stack visit-settings"><h3>Mejor época de visita</h3><div class="form-field"><label for="place-season-mode">Temporada</label><Select input-id="place-season-mode" v-model="form.season_mode" :options="seasonModes" option-label="label" option-value="value" :disabled="locked" @change="changeSeason" /></div>
              <div v-if="form.season_mode === 'months'" class="detail-grid"><div class="form-field"><label for="place-season-start">Mes de inicio</label><Select input-id="place-season-start" v-model="form.season_start_month" :options="months" option-label="name" option-value="id" placeholder="Selecciona un mes" :disabled="locked" /></div><div class="form-field"><label for="place-season-end">Mes de fin</label><Select input-id="place-season-end" v-model="form.season_end_month" :options="months" option-label="name" option-value="id" placeholder="Selecciona un mes" :disabled="locked" /></div></div><small>Es una recomendación de visita. No cambia automáticamente la publicación ni los días de apertura.</small><small v-if="form.season_mode === 'months' && form.season_start_month && form.season_end_month && form.season_end_month < form.season_start_month">La temporada termina en el año siguiente.</small>
            </div>
            <div class="form-field"><label for="place-price">Precio de entrada (Bs) *</label><InputNumber input-id="place-price" v-model="price" :min="0" :max="1000000" :max-fraction-digits="2" :min-fraction-digits="2" :disabled="locked" /><small>Usa 0 si la entrada es gratuita. Es un importe informativo; aquí no se venden entradas.</small></div>
            <div class="form-field"><label for="place-recommendations">Recomendaciones para la visita</label><Textarea id="place-recommendations" v-model.trim="form.recommendations" rows="4" maxlength="3000" /></div>
            <div class="form-field"><label for="place-phone">Teléfono de contacto</label><InputText id="place-phone" v-model.trim="form.phone" type="tel" maxlength="30" /></div>
          </div></template></Card>
          <div class="stack">
            <Card><template #title>Ubicación</template><template #content><div class="stack">
              <div class="form-field"><label for="place-department">Departamento *</label><Select input-id="place-department" v-model="form.department" :options="options.departments" :disabled="locked" /></div>
              <div class="form-field"><label for="place-city">Ciudad o municipio *</label><InputText id="place-city" v-model.trim="form.city" required minlength="2" maxlength="100" /></div>
              <div class="form-field"><label for="place-address">Dirección o referencia de acceso *</label><InputText id="place-address" v-model.trim="form.address" required minlength="3" maxlength="250" /></div>
              <ClientOnly><LocationMap :latitude="form.latitude" :longitude="form.longitude" editable :disabled="locked" @change="setLocation" /><template #fallback><p>Cargando mapa…</p></template></ClientOnly>
              <details><summary>Coordenadas (opcional)</summary><div class="detail-grid"><div class="form-field"><label for="place-latitude">Latitud</label><InputNumber input-id="place-latitude" v-model="form.latitude" :min="-90" :max="90" :max-fraction-digits="7" :use-grouping="false" :disabled="locked" /></div><div class="form-field"><label for="place-longitude">Longitud</label><InputNumber input-id="place-longitude" v-model="form.longitude" :min="-180" :max="180" :max-fraction-digits="7" :use-grouping="false" :disabled="locked" /></div></div>
              </details><small>La ubicación elegida se mostrará en la ficha pública.</small>
            </div></template></Card>
            <Card><template #title>Responsable y publicación</template><template #content><div class="stack">
              <template v-if="admin"><div class="form-field"><label for="place-manager-search">Buscar encargado</label><div class="cluster"><InputText id="place-manager-search" v-model.trim="managerSearch" placeholder="Nombre o correo" @keydown.enter.prevent="searchManagers" /><Button label="Buscar" :loading="managerLoading" @click="searchManagers" /></div></div><Message v-if="managerError" severity="error" :closable="false">{{ managerError }}</Message><div class="form-field"><label for="place-manager">Encargado de atracción *</label><Select input-id="place-manager" v-model="form.manager_id" :options="managerChoices" option-label="label" option-value="id" placeholder="Selecciona un encargado" :disabled="locked" /></div><small>Un encargado puede gestionar varias atracciones. Se muestran hasta 50 resultados.</small><NuxtLink to="/admin/users">Crear o gestionar encargados</NuxtLink><div class="form-field"><label for="place-state">Estado</label><Select input-id="place-state" v-model="form.status" :options="states" option-label="label" option-value="value" :disabled="locked" @change="hideInactive" /></div></template>
              <p v-else class="muted">Encargado: {{ record?.manager_name }}</p>
              <div class="cluster"><Checkbox input-id="place-published" v-model="form.published" binary :disabled="locked || form.status === 'inactive'" /><label for="place-published">Publicar en el catálogo</label></div><small>Una atracción inactiva queda oculta. Puedes guardar un borrador y completarlo después.</small>
            </div></template></Card>
          </div>
        </div>
        <Card class="photos-card"><template #title>Fotografías</template><template #subtitle>Hasta 6 imágenes PNG o JPG. Máximo 5 MB y 4096 × 4096 píxeles por imagen. La primera será la portada.</template><template #content>
          <div class="form-field"><label for="place-photos">Agregar fotografías</label><input id="place-photos" type="file" accept="image/png,image/jpeg" multiple :disabled="locked || reading || photos.length >= 6" @change="readPhotos"></div>
          <Message v-if="photoError" severity="error" :closable="false">{{ photoError }}</Message>
          <p v-if="reading" role="status">Procesando fotografías…</p>
          <div class="photo-grid"><div v-for="(photo, i) in photos" :key="photo.key" class="photo-item"><img :src="photo.preview" :alt="`Fotografía ${i + 1} de la atracción`"><span>{{ i === 0 ? 'Portada' : `Fotografía ${i + 1}` }}</span><div class="cluster"><Button v-if="i > 0" label="Usar de portada" size="small" text :disabled="locked || reading" @click="makeCover(i)" /><Button label="Quitar" size="small" severity="danger" text :disabled="locked || reading" @click="photos.splice(i, 1)" /></div></div></div>
        </template></Card>
      </fieldset>
      <div class="editor-actions"><span class="muted">{{ dirty ? 'Tienes cambios sin guardar' : 'Datos actualizados' }}</span><div class="cluster"><Button v-if="!create" label="Recargar datos" outlined severity="secondary" :disabled="saving || reading" @click="reload" /><Button type="submit" :label="create ? 'Crear atracción' : 'Guardar cambios'" icon="pi pi-save" :loading="saving" :disabled="locked || reading || (!create && !dirty)" /></div></div>
    </form>
  </div>
</template>
<script setup lang="ts">
import LocationMap from './LocationMap.client.vue'
import ClassificationSelect from './ClassificationSelect.vue'
import TimeSelect from './TimeSelect.vue'
import { weekDays, months } from '~/utils/attractions'
import Button from 'primevue/button'
import Card from 'primevue/card'
import Checkbox from 'primevue/checkbox'
import InputText from 'primevue/inputtext'
import InputNumber from 'primevue/inputnumber'
import Textarea from 'primevue/textarea'
import Select from 'primevue/select'
import Message from 'primevue/message'
import { useToast } from 'primevue/usetoast'
import { useConfirm } from 'primevue/useconfirm'
import type { Attraction, AttractionOptions } from '~/types/attraction'
import type { AgencyManager } from '~/types/agency'
import type { ApiEnvelope } from '~/types/api'
import { apiError } from '~/utils/api-error'
const props = defineProps<{ admin?: boolean; create?: boolean; attractionId?: number }>()
const emit = defineEmits<{ saved: [attraction: Attraction] }>()
const auth = useAuthStore(); const config = useRuntimeConfig(); const toast = useToast(); const confirm = useConfirm()
const loading = ref(true); const ready = ref(false); const saving = ref(false); const error = ref(''); const fields = ref<Record<string, string>>({})
const options = ref<AttractionOptions>({ departments: [], categories: [] }); const record = ref<Attraction | null>(null)
const managers = ref<AgencyManager[]>([]); const managerSearch = ref(''); const managerLoading = ref(false); const managerError = ref('')
const reading = ref(false); const photoError = ref(''); const price = ref<number | null>(0)
interface EditorPhoto { key: string; id?: number; data?: string; preview: string }
const photos = ref<EditorPhoto[]>([])
const form = reactive({ name: '', description: '', subcategory_ids: [] as number[], department: 'Tarija', city: '', address: '', latitude: null as number | null, longitude: null as number | null, schedule_mode: 'unspecified', opening_time: '', closing_time: '', opening_days: [] as number[], season_mode: 'unspecified', season_start_month: null as number | null, season_end_month: null as number | null, recommendations: '', phone: '', manager_id: null as number | null, status: 'active', published: false, version: 0 })
const scheduleModes = [{ label: 'Por confirmar', value: 'unspecified' }, { label: 'Seleccionar horas', value: 'scheduled' }, { label: '24 horas', value: 'all_day' }]
const seasonModes = [{ label: 'Por confirmar', value: 'unspecified' }, { label: 'Todo el año', value: 'all_year' }, { label: 'Seleccionar meses', value: 'months' }]
const changeSchedule = () => { if (form.schedule_mode !== 'scheduled') { form.opening_time = ''; form.closing_time = '' }; if (form.schedule_mode === 'unspecified') form.opening_days = [] }
const changeSeason = () => { if (form.season_mode !== 'months') { form.season_start_month = null; form.season_end_month = null } }
const states = [{ label: 'Activa', value: 'active' }, { label: 'Inactiva', value: 'inactive' }]
const initial = ref(''); const snapshot = () => JSON.stringify({ form, price: price.value, photos: photos.value })
const dirty = computed(() => ready.value && snapshot() !== initial.value)
const locked = computed(() => saving.value || (!props.admin && form.status === 'inactive'))
const prefix = computed(() => props.admin ? '/admin/attractions' : '/managed-attractions')
const endpoint = computed(() => `${prefix.value}${props.create ? '' : `/${props.attractionId}`}`)
const managerChoices = computed(() => {
  const result = managers.value.map(m => ({ id: m.id, label: `${m.first_name} ${m.last_name} · ${m.email}` }))
  if (record.value?.manager_id && !result.some(m => m.id === record.value?.manager_id)) result.unshift({ id: record.value.manager_id, label: `${record.value.manager_name} · ${record.value.manager_email}${record.value.manager_status === 'inactive' ? ' (inactivo)' : ''}` })
  return result
})
const apply = (a: Attraction) => {
  record.value = a
  Object.assign(form, { name: a.name, description: a.description, subcategory_ids: (a.subcategories || []).map(s => s.subcategory_id), department: a.department, city: a.city, address: a.address, latitude: a.latitude, longitude: a.longitude, schedule_mode: a.schedule_mode, opening_time: a.opening_time, closing_time: a.closing_time, opening_days: [...(a.opening_days || [])], season_mode: a.season_mode, season_start_month: a.season_start_month, season_end_month: a.season_end_month, recommendations: a.recommendations, phone: a.phone, manager_id: a.manager_id || null, status: a.status || 'active', published: a.published || false, version: a.version || 0 })
  price.value = a.admission_cents / 100
  photos.value = (a.photos || []).map(p => ({ key: String(p.id), id: p.id, preview: `${config.public.apiBase}${prefix.value}/${a.id}/photos/${p.id}` }))
  initial.value = snapshot()
}
let managerSerial = 0
const searchManagers = async () => {
  const current = ++managerSerial; managerLoading.value = true; managerError.value = ''
  try { const result = await auth.request<AgencyManager[]>(`/admin/attractions/managers?search=${encodeURIComponent(managerSearch.value)}`); if (current === managerSerial) managers.value = result.data || [] }
  catch (cause) { if (current === managerSerial) managerError.value = apiError(cause).message }
  finally { if (current === managerSerial) managerLoading.value = false }
}
const load = async () => {
  loading.value = true; ready.value = false; error.value = ''; fields.value = {}; photoError.value = ''
  try {
    const result = await $fetch<ApiEnvelope<AttractionOptions>>(`${config.public.apiBase}/attractions/options`); if (result.data) options.value = result.data
    if (!props.create) { const result = await auth.request<Attraction>(endpoint.value); if (!result.data) throw new Error('Missing attraction'); apply(result.data) }
    if (props.admin) await searchManagers()
    initial.value = snapshot(); ready.value = true
  } catch (cause) { error.value = apiError(cause).message }
  finally { loading.value = false }
}
const reload = () => { if (!dirty.value) { void load(); return }; confirm.require({ header: 'Recargar atracción', message: 'Se descartarán los cambios sin guardar.', acceptLabel: 'Recargar', rejectLabel: 'Seguir editando', accept: () => { void load() } }) }
const setLocation = (point: { latitude: number | null; longitude: number | null }) => { form.latitude = point.latitude; form.longitude = point.longitude }
const hideInactive = () => { if (form.status === 'inactive') form.published = false }
const makeCover = (i: number) => { const photo = photos.value.splice(i, 1)[0]; if (photo) photos.value.unshift(photo) }
const readPhotos = async (event: Event) => {
  const input = event.target as HTMLInputElement; const files = Array.from(input.files || []); photoError.value = ''
  if (files.length + photos.value.length > 6) { photoError.value = 'Puedes guardar hasta seis fotografías.'; input.value = ''; return }
  reading.value = true
  try {
    const additions: EditorPhoto[] = []
    for (const file of files) {
      if (!['image/png', 'image/jpeg'].includes(file.type) || file.size > 5 * 1024 * 1024) throw new Error('Selecciona fotografías PNG o JPG de hasta 5 MB cada una.')
      const data = await new Promise<string>((resolve, reject) => { const reader = new FileReader(); reader.onload = () => resolve(String(reader.result)); reader.onerror = () => reject(new Error('No se pudo leer la fotografía.')); reader.readAsDataURL(file) })
      additions.push({ key: crypto.randomUUID(), data, preview: data })
    }
    photos.value.push(...additions)
  } catch (cause) { photoError.value = cause instanceof Error ? cause.message : 'No se pudieron leer las fotografías.' }
  finally { reading.value = false; input.value = '' }
}
const save = async () => {
  if (locked.value || reading.value) return
  error.value = ''; fields.value = {}
  if (props.admin && !form.manager_id) { error.value = 'Selecciona un encargado de atracción.'; return }
  if (!form.subcategory_ids.length || form.subcategory_ids.length > 4) { error.value = 'Selecciona entre una y cuatro subcategorías.'; return }
  if (form.schedule_mode !== 'unspecified' && !form.opening_days.length) { error.value = 'Selecciona los días de apertura.'; return }
  if (form.schedule_mode === 'scheduled' && (!form.opening_time || !form.closing_time || form.opening_time === form.closing_time)) { error.value = 'Selecciona apertura y cierre distintos, o elige 24 horas.'; return }
  if (form.season_mode === 'months' && (!form.season_start_month || !form.season_end_month)) { error.value = 'Selecciona los meses de inicio y fin.'; return }
  if (price.value === null || price.value < 0) { error.value = 'Indica el precio de entrada o 0 si es gratuita.'; return }
  if (form.published && (form.description.length < 20 || !photos.value.length)) { error.value = 'Para publicar, completa una descripción de al menos 20 caracteres y una fotografía.'; return }
  saving.value = true
  try {
    const { manager_id, status, ...content } = form
    const body = { ...content, admission_cents: Math.round(price.value * 100), photos: photos.value.map(p => p.id ? { id: p.id } : { data: p.data }), ...(props.admin ? { manager_id, status } : {}) }
    const result = await auth.request<Attraction>(endpoint.value, { method: props.create ? 'POST' : 'PUT', body })
    if (!result.data) throw new Error('Missing attraction')
    apply(result.data); toast.add({ severity: 'success', summary: props.create ? 'Atracción creada' : 'Atracción actualizada', life: 4000 }); emit('saved', result.data)
  } catch (cause) { const failure = apiError(cause); error.value = failure.message; fields.value = failure.fields || {} }
  finally { saving.value = false }
}
onBeforeRouteLeave(() => {
  if (!dirty.value && !reading.value) return true
  return new Promise<boolean>(resolve => confirm.require({ header: 'Cambios sin guardar', message: '¿Salir sin guardar los cambios de la atracción?', acceptLabel: 'Salir sin guardar', rejectLabel: 'Seguir editando', accept: () => resolve(true), reject: () => resolve(false), onHide: () => resolve(false) }))
})
onMounted(load)
</script>
<style scoped>
fieldset { border:0; margin:0; padding:0; min-width:0; }
.editor-grid { display:grid; grid-template-columns:minmax(0,1.2fr) minmax(0,1fr); gap:1.5rem; }
.visit-settings { border-top:1px solid var(--color-border); padding-top:1rem; }
.day-options { display:flex; flex-wrap:wrap; gap:1rem; }
.photos-card { margin-top:1.5rem; }
.photo-grid { display:grid; grid-template-columns:repeat(auto-fill,minmax(min(100%,15rem),1fr)); gap:1rem; margin-top:1rem; }
.photo-item { border:1px solid var(--color-border); border-radius:var(--radius-md); padding:.75rem; }
.photo-item img { width:100%; aspect-ratio:4/3; object-fit:cover; border-radius:var(--radius-sm); margin-bottom:.5rem; }
.editor-actions { display:flex; flex-wrap:wrap; justify-content:space-between; align-items:center; gap:1rem; padding:1.25rem; background:var(--color-surface); border:1px solid var(--color-border); border-radius:var(--radius-md); }
input[type=file] { max-width:100%; padding:.75rem; border:1px solid var(--color-border); }
@media(max-width:900px) { .editor-grid { grid-template-columns:1fr; } }
</style>
