<template>
  <div class="stack">
    <UiLoadingState v-if="loading" />
    <Message v-else-if="loadError" severity="error" :closable="false">{{ loadError }} <Button label="Reintentar" text @click="load" /></Message>
    <form v-else class="stack" @submit.prevent="save">
      <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>
      <div class="editor-grid">
        <Card><template #title>Información del paquete</template><template #content><div class="stack">
          <div class="form-field"><label for="package-name">Nombre *</label><InputText id="package-name" v-model.trim="form.name" required maxlength="160" /></div>
          <div class="form-field"><label for="package-description">Descripción</label><Textarea id="package-description" v-model.trim="form.description" rows="6" maxlength="6000" /><small>Para publicar, escribe al menos 20 caracteres.</small></div>
          <div class="detail-grid">
            <div class="form-field"><label for="package-days">Días *</label><InputNumber input-id="package-days" v-model="form.duration_days" :min="1" :max="30" show-buttons /></div>
            <div class="form-field"><label for="package-nights">Noches</label><InputNumber input-id="package-nights" v-model="form.duration_nights" :min="0" :max="Math.max(0, form.duration_days - 1)" show-buttons /></div>
          </div>
          <div class="form-field"><label for="package-difficulty">Dificultad</label><Select input-id="package-difficulty" v-model="form.difficulty" :options="difficultyOptions" option-label="label" option-value="value" /></div>
        </div></template></Card>
        <Card><template #title>Tarifas y cancelación</template><template #content><div class="stack">
          <div class="form-field"><label for="package-price">Precio nacional por persona (Bs) *</label><InputNumber input-id="package-price" v-model="nationalPrice" :min="0" :max="1000000" :min-fraction-digits="2" :max-fraction-digits="2" /></div>
          <div class="form-field"><label for="package-foreign">Costo adicional por extranjero (Bs)</label><InputNumber input-id="package-foreign" v-model="foreignSurcharge" :min="0" :max="1000000" :min-fraction-digits="2" :max-fraction-digits="2" /></div>
          <small>La edad mínima de pago pertenece a la agencia y actualmente es {{ record?.minimum_paying_age ?? 'la configurada' }} años.</small>
          <div class="cluster"><Checkbox input-id="package-cancellation" v-model="form.cancellation_allowed" binary @change="clearCancellation" /><label for="package-cancellation">Permitir cancelación del turista</label></div>
          <div v-if="form.cancellation_allowed" class="form-field"><label for="package-cancellation-hours">Anticipación mínima (horas)</label><InputNumber input-id="package-cancellation-hours" v-model="form.cancellation_notice_hours" :min="1" :max="8760" show-buttons /><small>En esta versión, el reembolso será del 100 %.</small></div>
          <div class="cluster"><Checkbox input-id="package-published" v-model="form.published" binary /><label for="package-published">Publicar en el catálogo</label></div>
          <small>Puedes guardar un borrador y completarlo después.</small>
        </div></template></Card>
      </div>

      <Card><template #title>Programación de salidas</template><template #subtitle>Define cuándo podrá viajar y comprar el turista.</template><template #content><div class="stack">
        <div class="cluster"><Checkbox input-id="package-schedule" v-model="scheduleEnabled" binary /><label for="package-schedule">Configurar disponibilidad</label></div>
        <Message v-if="!scheduleEnabled" severity="info" :closable="false">Puedes guardar el paquete como borrador. Para publicarlo debes configurar su frecuencia.</Message>
        <template v-else>
          <div class="form-field"><label>Frecuencia *</label><SelectButton v-model="scheduleForm.frequency_type" :options="frequencyOptions" option-label="label" option-value="value" :allow-empty="false" /></div>
          <div class="schedule-grid">
            <div class="form-field"><label for="schedule-from">{{ scheduleForm.frequency_type === 'single' ? 'Fecha de salida *' : 'Desde *' }}</label><DatePicker input-id="schedule-from" v-model="scheduleStartDate" date-format="dd/mm/yy" show-icon :min-date="today" /></div>
            <div v-if="scheduleForm.frequency_type !== 'single'" class="form-field"><label for="schedule-until">Hasta *</label><DatePicker input-id="schedule-until" v-model="scheduleEndDate" date-format="dd/mm/yy" show-icon :min-date="scheduleStartDate || today" :max-date="scheduleMaxEnd" /></div>
            <div class="form-field"><label for="schedule-departure-time">Hora de salida *</label><DatePicker input-id="schedule-departure-time" v-model="scheduleDepartureTime" time-only hour-format="24" show-icon /></div>
            <div class="form-field"><label for="schedule-meeting-time">Hora de encuentro *</label><DatePicker input-id="schedule-meeting-time" v-model="scheduleMeetingTime" time-only hour-format="24" show-icon /><small>No puede ser posterior a la hora de salida.</small></div>
          </div>
          <div v-if="scheduleForm.frequency_type === 'specific_weekdays'" class="form-field"><label>Días de salida *</label><SelectButton v-model="scheduleForm.weekdays" :options="weekdayOptions" option-label="label" option-value="value" multiple /></div>
          <div class="schedule-grid">
            <div class="form-field"><label for="schedule-min-capacity">Cupo mínimo *</label><InputNumber input-id="schedule-min-capacity" v-model="scheduleForm.default_min_capacity" :min="1" :max="500" show-buttons /></div>
            <div class="form-field"><label for="schedule-max-capacity">Cupo máximo *</label><InputNumber input-id="schedule-max-capacity" v-model="scheduleForm.default_max_capacity" :min="1" :max="500" show-buttons /></div>
            <div class="form-field"><label for="schedule-cutoff">Cerrar compras antes de la salida *</label><InputNumber input-id="schedule-cutoff" v-model="scheduleForm.booking_cutoff_hours" :min="1" :max="720" suffix=" horas" show-buttons /></div>
            <div class="form-field"><label for="schedule-advance">Máxima anticipación de compra *</label><InputNumber input-id="schedule-advance" v-model="scheduleForm.maximum_advance_days" :min="1" :max="365" suffix=" días" show-buttons /><small>Hoy se podrán comprar salidas hasta el {{ formatLongDate(maximumBookableDate) }}.</small></div>
          </div>
          <div class="form-field"><label for="schedule-meeting-point">Punto de encuentro predeterminado</label><InputText id="schedule-meeting-point" v-model.trim="scheduleForm.default_meeting_point" maxlength="500" /><small>Se copiará a cada salida y podrá modificarse individualmente.</small></div>
          <div class="form-field"><label for="schedule-instructions">Instrucciones para turistas</label><Textarea id="schedule-instructions" v-model.trim="scheduleForm.default_instructions" rows="3" maxlength="3000" /></div>
          <Message severity="success" :closable="false"><strong>{{ schedulePreview.length }} {{ schedulePreview.length === 1 ? 'salida' : 'salidas' }}</strong> en la programación. <span v-if="schedulePreview.length">Primeras fechas: {{ schedulePreview.slice(0, 5).map(formatPreviewDate).join(' · ') }}<span v-if="schedulePreview.length > 5"> · …</span></span></Message>
          <Message severity="secondary" :closable="false">Solo se generarán fechas futuras. Las salidas con cupos retenidos o confirmados conservarán sus condiciones al cambiar esta programación.</Message>
        </template>

        <div v-if="record?.departures?.length" class="stack">
          <h3>Próximas salidas generadas</h3>
          <div class="departure-list">
            <div v-for="departure in record.departures.slice(0, 8)" :key="departure.id" class="departure-row">
              <div><strong>{{ formatDeparture(departure.starts_at) }}</strong><small>Compras: {{ formatWindow(departure.booking_opens_at, departure.booking_closes_at) }}</small></div>
              <div>
                <span>{{ departure.confirmed_capacity }} confirmados de {{ departure.max_capacity }}</span>
                <div class="cluster"><Tag v-if="departure.is_exception && departure.status !== 'minimum_review'" value="Ajustada" severity="info" /><Tag :value="departureStatus(departure.status)" :severity="departure.status === 'cancelled' ? 'danger' : departure.status === 'minimum_review' ? 'warn' : departure.status === 'closed' ? 'secondary' : 'success'" /></div>
                <Message v-if="departure.status === 'minimum_review'" severity="warn" :closable="false">Cerraron las compras con {{ departure.confirmed_capacity }} de {{ departure.min_capacity }} participantes mínimos. Revisa y confirma antes de iniciar los reembolsos.</Message>
                <div v-if="departure.status === 'minimum_review'" class="cluster"><Button type="button" label="Confirmar cancelación y reembolsos" icon="pi pi-wallet" size="small" severity="danger" :loading="departureBusy" :disabled="dirty" @click="confirmMinimumRefund(departure)" /></div>
                <div v-else-if="departure.status !== 'cancelled' && departure.status !== 'completed'" class="cluster"><Button type="button" label="Ajustar" icon="pi pi-pencil" size="small" outlined :disabled="dirty" @click="editDeparture(departure)" /><Button type="button" label="Cancelar" icon="pi pi-times" size="small" severity="danger" outlined :disabled="dirty" @click="openDepartureCancellation(departure)" /></div>
              </div>
            </div>
          </div>
          <small v-if="record.departures.length > 8">Se muestran las primeras 8 de {{ record.departures.length }} salidas.</small>
          <small v-if="dirty">Guarda los cambios generales antes de gestionar una salida.</small>
        </div>
      </div></template></Card>

      <Card><template #title>Información para el viajero</template><template #subtitle>Escribe un elemento por línea.</template><template #content><div class="three-columns">
        <div class="form-field"><label for="package-includes">Incluye</label><Textarea id="package-includes" v-model="includesText" rows="6" placeholder="Transporte&#10;Almuerzo&#10;Guía" /></div>
        <div class="form-field"><label for="package-excludes">No incluye</label><Textarea id="package-excludes" v-model="excludesText" rows="6" placeholder="Gastos personales" /></div>
        <div class="form-field"><label for="package-bring">Qué llevar</label><Textarea id="package-bring" v-model="bringText" rows="6" placeholder="Protector solar&#10;Agua" /></div>
      </div></template></Card>

      <Card><template #title>Itinerario por días</template><template #subtitle>Una atracción puede aparecer en diferentes días.</template><template #content><div class="stack">
        <div class="cluster"><Button type="button" label="Agregar día" icon="pi pi-plus" outlined :disabled="itinerary.length >= form.duration_days" @click="addDay" /><small>{{ itinerary.length }} de {{ form.duration_days }} días descritos</small></div>
        <Message v-if="!itinerary.length" severity="info" :closable="false">Agrega al menos un día antes de publicar.</Message>
        <article v-for="(day, index) in itinerary" :key="day.key" class="itinerary-day stack">
          <div class="day-heading"><h3>Día {{ day.day_number }}</h3><Button type="button" label="Quitar día" severity="danger" text size="small" @click="removeDay(index)" /></div>
          <div class="form-field"><label :for="`day-title-${day.key}`">Título *</label><InputText :id="`day-title-${day.key}`" v-model.trim="day.title" maxlength="160" /></div>
          <div class="form-field"><label :for="`day-description-${day.key}`">Descripción</label><Textarea :id="`day-description-${day.key}`" v-model.trim="day.description" rows="3" maxlength="3000" /></div>
          <div class="form-field"><label :for="`day-activities-${day.key}`">Actividades (una por línea)</label><Textarea :id="`day-activities-${day.key}`" v-model="day.activitiesText" rows="4" /></div>
          <div class="form-field"><label :for="`day-attractions-${day.key}`">Atracciones asociadas</label><MultiSelect :input-id="`day-attractions-${day.key}`" v-model="day.attraction_ids" :options="attractionChoices" option-label="label" option-value="id" display="chip" filter placeholder="Selecciona atracciones publicadas" /></div>
        </article>
      </div></template></Card>

      <Card><template #title>Fotografías</template><template #subtitle>Hasta 6 imágenes PNG o JPG de 5 MB. La primera será la portada.</template><template #content><div class="stack">
        <div class="form-field"><label for="package-photos">Agregar fotografías</label><input id="package-photos" type="file" accept="image/png,image/jpeg" multiple :disabled="reading || photos.length >= 6" @change="readPhotos"></div>
        <Message v-if="photoError" severity="error" :closable="false">{{ photoError }}</Message>
        <p v-if="reading" role="status">Procesando fotografías…</p>
        <div class="photo-grid"><div v-for="(photo, index) in photos" :key="photo.key" class="photo-item"><img :src="photo.preview" :alt="`Fotografía ${index + 1} del paquete`"><span>{{ index === 0 ? 'Portada' : `Fotografía ${index + 1}` }}</span><div class="cluster"><Button v-if="index" type="button" label="Usar de portada" size="small" text @click="makeCover(index)" /><Button type="button" label="Quitar" size="small" severity="danger" text @click="photos.splice(index, 1)" /></div></div></div>
      </div></template></Card>

      <div class="editor-actions"><span class="muted">{{ dirty ? 'Tienes cambios sin guardar' : 'Datos actualizados' }}</span><div class="cluster"><Button v-if="!create" type="button" label="Recargar datos" outlined severity="secondary" :disabled="saving || reading" @click="reload" /><Button type="submit" :label="create ? 'Crear paquete' : 'Guardar cambios'" icon="pi pi-save" :loading="saving" :disabled="saving || reading || (!create && !dirty)" /></div></div>
    </form>

    <Dialog v-model:visible="departureEditorOpen" modal header="Ajustar salida" :style="{ width: 'min(42rem, 94vw)' }">
      <div v-if="editingDeparture" class="stack">
        <Message v-if="departureError" severity="error" :closable="false">{{ departureError }}</Message>
        <Message severity="info" :closable="false">{{ formatDeparture(editingDeparture.starts_at) }}. La fecha y hora de salida provienen de la programación.</Message>
        <div class="schedule-grid departure-editor-grid">
          <div class="form-field"><label for="departure-meeting-time">Hora de encuentro *</label><DatePicker input-id="departure-meeting-time" v-model="departureMeetingTime" time-only hour-format="24" show-icon /></div>
          <div class="form-field"><label for="departure-cutoff">Cerrar compras *</label><InputNumber input-id="departure-cutoff" v-model="departureForm.booking_cutoff_hours" :min="1" :max="720" suffix=" horas antes" show-buttons /></div>
          <div class="form-field"><label for="departure-min">Cupo mínimo *</label><InputNumber input-id="departure-min" v-model="departureForm.min_capacity" :min="1" :max="500" show-buttons /></div>
          <div class="form-field"><label for="departure-max">Cupo máximo *</label><InputNumber input-id="departure-max" v-model="departureForm.max_capacity" :min="1" :max="500" show-buttons /></div>
        </div>
        <div class="form-field"><label for="departure-point">Punto de encuentro</label><InputText id="departure-point" v-model.trim="departureForm.meeting_point" maxlength="500" /></div>
        <div class="form-field"><label for="departure-instructions">Instrucciones</label><Textarea id="departure-instructions" v-model.trim="departureForm.instructions" rows="4" maxlength="3000" /></div>
      </div>
      <template #footer><Button type="button" label="Cerrar" severity="secondary" text :disabled="departureBusy" @click="departureEditorOpen = false" /><Button type="button" label="Guardar ajuste" icon="pi pi-save" :loading="departureBusy" @click="saveDeparture" /></template>
    </Dialog>

    <Dialog v-model:visible="departureCancelOpen" modal header="Cancelar salida" :style="{ width: 'min(36rem, 94vw)' }">
      <div v-if="cancellingDeparture" class="stack">
        <Message severity="warn" :closable="false">Se cancelará la salida del {{ formatDeparture(cancellingDeparture.starts_at) }}. La salida permanecerá registrada para conservar la trazabilidad.</Message>
        <Message v-if="departureError" severity="error" :closable="false">{{ departureError }}</Message>
        <div class="form-field"><label for="departure-cancel-reason">Motivo de cancelación *</label><Textarea id="departure-cancel-reason" v-model.trim="departureCancelReason" rows="4" minlength="5" maxlength="1000" /></div>
      </div>
      <template #footer><Button type="button" label="Volver" severity="secondary" text :disabled="departureBusy" @click="departureCancelOpen = false" /><Button type="button" label="Cancelar salida" icon="pi pi-times" severity="danger" :loading="departureBusy" @click="cancelDeparture" /></template>
    </Dialog>
  </div>
</template>
<script setup lang="ts">
import Button from 'primevue/button'
import Card from 'primevue/card'
import Checkbox from 'primevue/checkbox'
import DatePicker from 'primevue/datepicker'
import Dialog from 'primevue/dialog'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import MultiSelect from 'primevue/multiselect'
import Select from 'primevue/select'
import SelectButton from 'primevue/selectbutton'
import Textarea from 'primevue/textarea'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import type { PackageAttractionSummary, PackageDeparture, PackageFrequency, TourPackage } from '~/types/package'
import { apiError } from '~/utils/api-error'
const props = defineProps<{ create?: boolean; packageId?: number }>()
const emit = defineEmits<{ saved: [item: TourPackage] }>()
const auth = useAuthStore(); const config = useRuntimeConfig(); const confirm = useConfirm(); const toast = useToast()
const loading = ref(true); const saving = ref(false); const reading = ref(false); const loadError = ref(''); const error = ref(''); const photoError = ref('')
const record = ref<TourPackage | null>(null); const attractions = ref<PackageAttractionSummary[]>([])
const form = reactive({ name: '', description: '', duration_days: 1, duration_nights: 0, difficulty: '' as TourPackage['difficulty'], cancellation_allowed: false, cancellation_notice_hours: 0, published: false, version: 0 })
const scheduleEnabled = ref(false)
const scheduleForm = reactive({ id: 0, frequency_type: 'daily' as PackageFrequency, weekdays: [] as number[], default_min_capacity: 1, default_max_capacity: 10, booking_cutoff_hours: 24, maximum_advance_days: 180, default_meeting_point: '', default_instructions: '' })
const today = new Date(); today.setHours(0, 0, 0, 0)
const scheduleStartDate = ref<Date | null>(null); const scheduleEndDate = ref<Date | null>(null)
const scheduleDepartureTime = ref<Date | null>(clockToDate('08:00')); const scheduleMeetingTime = ref<Date | null>(clockToDate('07:45'))
const nationalPrice = ref<number | null>(0); const foreignSurcharge = ref<number | null>(0)
const includesText = ref(''); const excludesText = ref(''); const bringText = ref('')
const departureEditorOpen = ref(false); const departureCancelOpen = ref(false); const departureBusy = ref(false); const departureError = ref('')
const editingDeparture = ref<PackageDeparture | null>(null); const cancellingDeparture = ref<PackageDeparture | null>(null)
const departureMeetingTime = ref<Date | null>(null); const departureCancelReason = ref('')
const departureForm = reactive({ min_capacity: 1, max_capacity: 1, booking_cutoff_hours: 24, meeting_point: '', instructions: '' })
interface EditorDay { key: string; day_number: number; title: string; description: string; activitiesText: string; attraction_ids: number[] }
interface EditorPhoto { key: string; id?: number; data?: string; preview: string }
const itinerary = ref<EditorDay[]>([]); const photos = ref<EditorPhoto[]>([])
const difficultyOptions = [{ label: 'Por indicar', value: '' }, { label: 'Fácil', value: 'easy' }, { label: 'Moderada', value: 'moderate' }, { label: 'Exigente', value: 'demanding' }]
const frequencyOptions = [{ label: 'Salida única', value: 'single' }, { label: 'Salidas diarias', value: 'daily' }, { label: 'Días específicos', value: 'specific_weekdays' }]
const weekdayOptions = [{ label: 'Lun', value: 1 }, { label: 'Mar', value: 2 }, { label: 'Mié', value: 3 }, { label: 'Jue', value: 4 }, { label: 'Vie', value: 5 }, { label: 'Sáb', value: 6 }, { label: 'Dom', value: 7 }]
const attractionChoices = computed(() => attractions.value.map(item => ({ id: item.id, label: `${item.name} · ${item.city}, ${item.department}` })))
const lines = (value: string) => value.split(/\r?\n/).map(item => item.trim()).filter(Boolean)
function clockToDate(value?: string | null) {
  if (!value) return null
  const match = value.match(/(\d{1,2}):(\d{2})/)
  if (!match) return null
  const result = new Date(); result.setHours(Number(match[1]), Number(match[2]), 0, 0)
  return result
}
const dateOnly = (value?: string | null) => {
  if (!value) return null
  const match = value.match(/(\d{4})-(\d{2})-(\d{2})/)
  if (!match) return null
  return new Date(Number(match[1]), Number(match[2]) - 1, Number(match[3]), 12, 0, 0, 0)
}
const formatDateInput = (value: Date | null) => value ? `${value.getFullYear()}-${String(value.getMonth() + 1).padStart(2, '0')}-${String(value.getDate()).padStart(2, '0')}` : ''
const formatClock = (value: Date | null) => value ? `${String(value.getHours()).padStart(2, '0')}:${String(value.getMinutes()).padStart(2, '0')}` : ''
const boliviaClockToDate = (value: string) => {
  const parts = new Intl.DateTimeFormat('en-GB', { hour: '2-digit', minute: '2-digit', hour12: false, timeZone: 'America/La_Paz' }).formatToParts(new Date(value))
  const result = new Date()
  result.setHours(Number(parts.find(part => part.type === 'hour')?.value || 0), Number(parts.find(part => part.type === 'minute')?.value || 0), 0, 0)
  return result
}
const scheduleMaxEnd = computed(() => scheduleStartDate.value ? new Date(scheduleStartDate.value.getFullYear() + 1, scheduleStartDate.value.getMonth(), scheduleStartDate.value.getDate(), 12) : undefined)
const maximumBookableDate = computed(() => new Date(today.getFullYear(), today.getMonth(), today.getDate() + Math.max(1, scheduleForm.maximum_advance_days), 12))
const schedulePreview = computed(() => {
  const start = scheduleStartDate.value
  const end = scheduleForm.frequency_type === 'single' ? start : scheduleEndDate.value
  if (!start || !end || end < start) return [] as Date[]
  const result: Date[] = []
  for (let day = new Date(start.getFullYear(), start.getMonth(), start.getDate(), 12); day <= end && result.length <= 366; day = new Date(day.getFullYear(), day.getMonth(), day.getDate() + 1, 12)) {
    const isoDay = day.getDay() || 7
    if (scheduleForm.frequency_type === 'single' || scheduleForm.frequency_type === 'daily' || scheduleForm.weekdays.includes(isoDay)) result.push(day)
  }
  return result
})
const formatPreviewDate = (value: Date) => new Intl.DateTimeFormat('es-BO', { day: '2-digit', month: 'short' }).format(value)
const formatLongDate = (value: Date) => new Intl.DateTimeFormat('es-BO', { day: 'numeric', month: 'long', year: 'numeric' }).format(value)
const formatDeparture = (value: string) => new Intl.DateTimeFormat('es-BO', { dateStyle: 'medium', timeStyle: 'short', timeZone: 'America/La_Paz' }).format(new Date(value))
const formatWindow = (open: string, close: string) => `${new Intl.DateTimeFormat('es-BO', { dateStyle: 'short', timeZone: 'America/La_Paz' }).format(new Date(open))} al ${new Intl.DateTimeFormat('es-BO', { dateStyle: 'short', timeStyle: 'short', timeZone: 'America/La_Paz' }).format(new Date(close))}`
const departureStatus = (status: string) => ({ draft: 'Borrador', open: 'Venta abierta', confirmed: 'Mínimo alcanzado', closed: 'Venta cerrada', minimum_review: 'Decisión por cupo mínimo', cancelled: 'Cancelada', completed: 'Realizada' }[status] || status)
const initial = ref('')
const snapshot = () => JSON.stringify({ form, scheduleEnabled: scheduleEnabled.value, scheduleForm, scheduleStartDate: formatDateInput(scheduleStartDate.value), scheduleEndDate: formatDateInput(scheduleEndDate.value), scheduleDepartureTime: formatClock(scheduleDepartureTime.value), scheduleMeetingTime: formatClock(scheduleMeetingTime.value), nationalPrice: nationalPrice.value, foreignSurcharge: foreignSurcharge.value, includesText: includesText.value, excludesText: excludesText.value, bringText: bringText.value, itinerary: itinerary.value, photos: photos.value.map(({ key, id, data }) => ({ key, id, data })) })
const dirty = computed(() => !loading.value && snapshot() !== initial.value)
const endpoint = computed(() => `/agency/packages${props.create ? '' : `/${props.packageId}`}`)
const apply = (item: TourPackage) => {
  record.value = item
  Object.assign(form, { name: item.name, description: item.description, duration_days: item.duration_days, duration_nights: item.duration_nights, difficulty: item.difficulty, cancellation_allowed: item.cancellation_allowed, cancellation_notice_hours: item.cancellation_notice_hours, published: item.published, version: item.version })
  scheduleEnabled.value = !!item.schedule
  if (item.schedule) {
    Object.assign(scheduleForm, { id: item.schedule.id, frequency_type: item.schedule.frequency_type, weekdays: [...(item.schedule.weekdays || [])], default_min_capacity: item.schedule.default_min_capacity, default_max_capacity: item.schedule.default_max_capacity, booking_cutoff_hours: item.schedule.booking_cutoff_hours, maximum_advance_days: item.schedule.maximum_advance_days, default_meeting_point: item.schedule.default_meeting_point, default_instructions: item.schedule.default_instructions })
    scheduleStartDate.value = dateOnly(item.schedule.valid_from); scheduleEndDate.value = dateOnly(item.schedule.valid_until)
    scheduleDepartureTime.value = clockToDate(item.schedule.departure_time); scheduleMeetingTime.value = clockToDate(item.schedule.meeting_time)
  }
  nationalPrice.value = item.national_price_cents / 100; foreignSurcharge.value = item.foreign_surcharge_cents / 100
  includesText.value = item.includes.join('\n'); excludesText.value = item.excludes.join('\n'); bringText.value = item.bring.join('\n')
  itinerary.value = item.itinerary.map(day => ({ key: String(day.id), day_number: day.day_number, title: day.title, description: day.description, activitiesText: day.activities.join('\n'), attraction_ids: day.attractions.map(link => link.attraction_id) }))
  photos.value = item.photos.map(photo => ({ key: String(photo.id), id: photo.id, preview: `${config.public.apiBase}/agency/packages/${item.id}/photos/${photo.id}` }))
  initial.value = snapshot()
}
const load = async () => {
  loading.value = true; loadError.value = ''; error.value = ''
  try {
    const options = await auth.request<PackageAttractionSummary[]>('/agency/packages/attractions')
    attractions.value = options.data || []
    if (!props.create) {
      const result = await auth.request<TourPackage>(endpoint.value)
      if (!result.data) throw new Error('Missing package')
      apply(result.data)
    } else initial.value = snapshot()
  } catch (cause) { loadError.value = apiError(cause).message }
  finally { loading.value = false }
}
const addDay = () => {
  const used = new Set(itinerary.value.map(day => day.day_number)); let dayNumber = 1
  while (used.has(dayNumber)) dayNumber++
  if (dayNumber > form.duration_days) return
  itinerary.value.push({ key: crypto.randomUUID(), day_number: dayNumber, title: '', description: '', activitiesText: '', attraction_ids: [] })
  itinerary.value.sort((a, b) => a.day_number - b.day_number)
}
const removeDay = (index: number) => { itinerary.value.splice(index, 1) }
const clearCancellation = () => { if (!form.cancellation_allowed) form.cancellation_notice_hours = 0 }
watch(() => scheduleForm.frequency_type, (frequency) => {
  if (frequency === 'single') {
    scheduleForm.weekdays = []
    scheduleEndDate.value = scheduleStartDate.value ? new Date(scheduleStartDate.value) : null
  } else if (!scheduleEndDate.value && scheduleStartDate.value) {
    scheduleEndDate.value = new Date(scheduleStartDate.value)
  }
  if (frequency !== 'specific_weekdays') scheduleForm.weekdays = []
})
watch(scheduleStartDate, (start) => {
  if (!start) return
  if (scheduleForm.frequency_type === 'single') scheduleEndDate.value = new Date(start)
  else if (!scheduleEndDate.value || scheduleEndDate.value < start) scheduleEndDate.value = new Date(start)
})
const makeCover = (index: number) => { const photo = photos.value.splice(index, 1)[0]; if (photo) photos.value.unshift(photo) }
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
const replaceDeparture = (updated: PackageDeparture) => {
  if (!record.value?.departures) return
  const index = record.value.departures.findIndex(item => item.id === updated.id)
  if (index >= 0) record.value.departures.splice(index, 1, updated)
}
const editDeparture = (departure: PackageDeparture) => {
  if (dirty.value) return
  departureError.value = ''
  editingDeparture.value = departure
  departureMeetingTime.value = boliviaClockToDate(departure.meeting_at)
  Object.assign(departureForm, {
    min_capacity: departure.min_capacity,
    max_capacity: departure.max_capacity,
    booking_cutoff_hours: Math.round((new Date(departure.starts_at).getTime() - new Date(departure.booking_closes_at).getTime()) / 3600000),
    meeting_point: departure.meeting_point,
    instructions: departure.instructions
  })
  departureEditorOpen.value = true
}
const saveDeparture = async () => {
  const departure = editingDeparture.value
  departureError.value = ''
  if (!departure || !departureMeetingTime.value) { departureError.value = 'Selecciona la hora de encuentro.'; return }
  if (departureForm.min_capacity < 1 || departureForm.max_capacity < departureForm.min_capacity) { departureError.value = 'Revisa el cupo mínimo y máximo.'; return }
  if (departureForm.max_capacity < departure.held_capacity + departure.confirmed_capacity) { departureError.value = 'El cupo máximo no puede ser menor que los cupos ocupados.'; return }
  departureBusy.value = true
  try {
    const result = await auth.request<PackageDeparture>(`${endpoint.value}/departures/${departure.id}`, {
      method: 'PATCH',
      body: { ...departureForm, version: departure.version, meeting_time: formatClock(departureMeetingTime.value) }
    })
    if (!result.data) throw new Error('Missing departure')
    replaceDeparture(result.data)
    editingDeparture.value = result.data
    departureEditorOpen.value = false
    toast.add({ severity: 'success', summary: 'Salida ajustada', detail: 'La programación general ya no sobrescribirá esta salida.', life: 4500 })
  } catch (cause) { departureError.value = apiError(cause).message }
  finally { departureBusy.value = false }
}
const openDepartureCancellation = (departure: PackageDeparture) => {
  if (dirty.value) return
  departureError.value = ''
  departureCancelReason.value = ''
  cancellingDeparture.value = departure
  departureCancelOpen.value = true
}
const cancelDeparture = async () => {
  const departure = cancellingDeparture.value
  departureError.value = ''
  if (!departure) return
  if (departureCancelReason.value.trim().length < 5) { departureError.value = 'Explica el motivo de cancelación con al menos 5 caracteres.'; return }
  departureBusy.value = true
  try {
    const result = await auth.request<PackageDeparture>(`${endpoint.value}/departures/${departure.id}/cancel`, {
      method: 'POST', body: { version: departure.version, reason: departureCancelReason.value.trim() }
    })
    if (!result.data) throw new Error('Missing departure')
    replaceDeparture(result.data)
    cancellingDeparture.value = result.data
    departureCancelOpen.value = false
    toast.add({ severity: 'success', summary: 'Salida cancelada', detail: 'El motivo y la fecha de cancelación quedaron registrados.', life: 4500 })
  } catch (cause) { departureError.value = apiError(cause).message }
  finally { departureBusy.value = false }
}
const runMinimumRefund = async (departure: PackageDeparture) => {
  departureBusy.value = true; departureError.value = ''
  try {
    const result = await auth.request<PackageDeparture>(`${endpoint.value}/departures/${departure.id}/minimum-refund`, { method: 'POST', body: { version: departure.version } })
    if (!result.data) throw new Error('Missing departure')
    replaceDeparture(result.data)
    toast.add({ severity: 'success', summary: 'Reembolsos iniciados', detail: 'La salida quedó cancelada y las compras afectadas pasaron a reembolso pendiente.', life: 5000 })
  } catch (cause) { departureError.value = apiError(cause).message }
  finally { departureBusy.value = false }
}
const confirmMinimumRefund = (departure: PackageDeparture) => confirm.require({
  header: 'Confirmar cupo mínimo no alcanzado',
  message: `La salida del ${formatDeparture(departure.starts_at)} se cancelará y todas sus compras pasarán a reembolso pendiente.`,
  icon: 'pi pi-exclamation-triangle',
  rejectLabel: 'Volver',
  acceptLabel: 'Confirmar reembolsos',
  acceptClass: 'p-button-danger',
  accept: () => { void runMinimumRefund(departure) }
})
const save = async () => {
  error.value = ''
  if (!form.name) { error.value = 'Escribe el nombre del paquete.'; return }
  if (form.duration_nights >= form.duration_days) { error.value = 'Las noches deben ser menores que los días.'; return }
  if (itinerary.value.some(day => day.day_number > form.duration_days || !day.title)) { error.value = 'Revisa los días y títulos del itinerario.'; return }
  if (form.cancellation_allowed && form.cancellation_notice_hours < 1) { error.value = 'Indica la anticipación mínima para cancelar.'; return }
  if (scheduleEnabled.value) {
    const end = scheduleForm.frequency_type === 'single' ? scheduleStartDate.value : scheduleEndDate.value
    if (!scheduleStartDate.value || !end || end < scheduleStartDate.value || end > (scheduleMaxEnd.value || end)) { error.value = 'Selecciona un rango válido de hasta doce meses.'; return }
    if (!scheduleDepartureTime.value || !scheduleMeetingTime.value || formatClock(scheduleMeetingTime.value) > formatClock(scheduleDepartureTime.value)) { error.value = 'La hora de encuentro debe ser igual o anterior a la hora de salida.'; return }
    if (scheduleForm.frequency_type === 'specific_weekdays' && !scheduleForm.weekdays.length) { error.value = 'Selecciona al menos un día específico de salida.'; return }
    if (!schedulePreview.value.length) { error.value = 'La programación no genera ninguna salida.'; return }
    if (scheduleForm.default_min_capacity < 1 || scheduleForm.default_max_capacity < scheduleForm.default_min_capacity) { error.value = 'Revisa el cupo mínimo y máximo.'; return }
    if (scheduleForm.booking_cutoff_hours < 1 || scheduleForm.maximum_advance_days < 1 || scheduleForm.booking_cutoff_hours >= scheduleForm.maximum_advance_days * 24) { error.value = 'La anticipación máxima debe ser mayor que el cierre de compras.'; return }
    if (form.published && scheduleForm.frequency_type === 'single') {
      const departure = new Date(scheduleStartDate.value); departure.setHours(scheduleDepartureTime.value.getHours(), scheduleDepartureTime.value.getMinutes(), 0, 0)
      if (departure <= new Date()) { error.value = 'La salida única debe tener una fecha y hora futuras.'; return }
      const closes = new Date(departure.getTime() - scheduleForm.booking_cutoff_hours * 60 * 60 * 1000)
      if (closes <= new Date()) { error.value = 'La salida única ya está fuera de su plazo de compra.'; return }
    }
  }
  if (form.published && (form.description.length < 20 || !nationalPrice.value || !photos.value.length || !itinerary.value.length || !scheduleEnabled.value)) { error.value = 'Para publicar, completa la descripción, el precio, una fotografía, el itinerario y la programación.'; return }
  saving.value = true
  try {
    const scheduleEnd = scheduleForm.frequency_type === 'single' ? scheduleStartDate.value : scheduleEndDate.value
    const body = {
      ...form,
      national_price_cents: Math.round((nationalPrice.value || 0) * 100), foreign_surcharge_cents: Math.round((foreignSurcharge.value || 0) * 100),
      includes: lines(includesText.value), excludes: lines(excludesText.value), bring: lines(bringText.value),
      photos: photos.value.map(photo => photo.id ? { id: photo.id } : { data: photo.data }),
      itinerary: itinerary.value.map(day => ({ day_number: day.day_number, title: day.title, description: day.description, activities: lines(day.activitiesText), attraction_ids: day.attraction_ids })),
      schedule: scheduleEnabled.value ? {
        ...scheduleForm,
        valid_from: formatDateInput(scheduleStartDate.value), valid_until: formatDateInput(scheduleEnd),
        departure_time: formatClock(scheduleDepartureTime.value), meeting_time: formatClock(scheduleMeetingTime.value)
      } : null
    }
    const result = await auth.request<TourPackage>(endpoint.value, { method: props.create ? 'POST' : 'PUT', body })
    if (!result.data) throw new Error('Missing package')
    apply(result.data); toast.add({ severity: 'success', summary: props.create ? 'Paquete creado' : 'Paquete actualizado', life: 4000 }); emit('saved', result.data)
  } catch (cause) { error.value = apiError(cause).message }
  finally { saving.value = false }
}
const reload = () => { if (!dirty.value) { void load(); return }; confirm.require({ header: 'Recargar paquete', message: 'Se descartarán los cambios sin guardar.', acceptLabel: 'Recargar', rejectLabel: 'Seguir editando', accept: () => { void load() } }) }
onBeforeRouteLeave(() => {
  if (!dirty.value && !reading.value) return true
  return new Promise<boolean>(resolve => confirm.require({ header: 'Cambios sin guardar', message: '¿Salir sin guardar los cambios del paquete?', acceptLabel: 'Salir sin guardar', rejectLabel: 'Seguir editando', accept: () => resolve(true), reject: () => resolve(false), onHide: () => resolve(false) }))
})
onMounted(load)
</script>
<style scoped>
.editor-grid { display:grid; grid-template-columns:repeat(2,minmax(0,1fr)); gap:1.5rem; }
.three-columns { display:grid; grid-template-columns:repeat(3,minmax(0,1fr)); gap:1.25rem; }
.schedule-grid { display:grid; grid-template-columns:repeat(4,minmax(0,1fr)); gap:1rem; }
.departure-editor-grid { grid-template-columns:repeat(2,minmax(0,1fr)); }
.departure-list { display:grid; gap:.75rem; }
.departure-row { display:flex; justify-content:space-between; gap:1rem; padding:1rem; border:1px solid var(--color-border-soft); border-radius:var(--radius-sm); background:var(--color-background-deep); }
.departure-row > div { display:flex; flex-direction:column; gap:.35rem; }
.departure-row > div:last-child { align-items:flex-end; text-align:right; }
.itinerary-day { padding:1.25rem; border:1px solid var(--color-border-soft); border-radius:var(--radius-md); background:rgba(248,249,242,.025); }
.day-heading { display:flex; justify-content:space-between; align-items:center; gap:1rem; }
.day-heading h3 { margin:0; }
.photo-grid { display:grid; grid-template-columns:repeat(auto-fill,minmax(min(100%,15rem),1fr)); gap:1rem; }
.photo-item { padding:.75rem; border:1px solid var(--color-border); border-radius:var(--radius-md); }
.photo-item img { display:block; width:100%; margin-bottom:.5rem; aspect-ratio:4/3; object-fit:cover; border-radius:var(--radius-sm); }
.editor-actions { display:flex; flex-wrap:wrap; justify-content:space-between; align-items:center; gap:1rem; padding:1.25rem; border:1px solid var(--color-border); border-radius:var(--radius-md); background:var(--color-surface); }
input[type=file] { max-width:100%; padding:.75rem; border:1px solid var(--color-border); }
@media(max-width:900px) { .editor-grid,.three-columns,.schedule-grid { grid-template-columns:1fr; } .departure-row { flex-direction:column; } .departure-row > div:last-child { align-items:flex-start; text-align:left; } }
</style>
