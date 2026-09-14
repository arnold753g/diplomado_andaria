<template>
  <UiLoadingState v-if="loading" />
  <UiEmptyState v-else-if="notAssigned" title="Aún no tienes una agencia asignada" description="El administrador debe crear tu agencia y asignarte como encargado." />
  <div v-else class="stack">
    <Message v-if="error" severity="error" :closable="false">{{ error }}<ul v-if="Object.keys(fieldErrors).length"><li v-for="(message, field) in fieldErrors" :key="field">{{ message }}</li></ul></Message>
    <Button v-if="!ready" label="Reintentar" icon="pi pi-refresh" @click="load" />
    <form v-if="ready" class="stack" @submit.prevent="submit">
      <Message v-if="!admin && form.status === 'inactive'" severity="warn" :closable="false">Tu agencia está desactivada. Puedes consultar sus datos; contacta al administrador para modificarla.</Message>
      <Message v-if="record?.manager_status === 'inactive'" severity="warn" :closable="false">El encargado está desactivado. Reactiva su cuenta o asigna otro encargado.</Message>
      <fieldset class="agency-fieldset" :disabled="locked">
        <div class="agency-editor-grid">
          <Card>
            <template #title><i class="pi pi-building" aria-hidden="true" /> Información de la agencia</template>
            <template #content><div class="stack">
              <div class="form-field"><label for="agency-name">Nombre comercial *</label><InputText id="agency-name" v-model.trim="form.name" required minlength="2" maxlength="160" /></div>
              <div class="form-field"><label for="agency-description">Descripción</label><Textarea id="agency-description" v-model.trim="form.description" rows="4" maxlength="3000" auto-resize /></div>
              <div class="detail-grid">
                <div class="form-field"><label for="agency-department">Departamento *</label><Select input-id="agency-department" v-model="form.department" :options="departments" :disabled="locked" /></div>
                <div class="form-field"><label for="agency-city">Ciudad o municipio *</label><InputText id="agency-city" v-model.trim="form.city" required minlength="2" maxlength="100" /></div>
              </div>
              <div class="form-field"><label for="agency-address">Dirección *</label><InputText id="agency-address" v-model.trim="form.address" required minlength="3" maxlength="250" /></div>
              <div class="form-field"><label for="agency-phone">Teléfono de contacto *</label><InputText id="agency-phone" v-model.trim="form.phone" type="tel" required minlength="7" maxlength="30" /></div>
              <div class="form-field"><label for="agency-email">Correo de contacto *</label><InputText id="agency-email" v-model.trim="form.email" type="email" required maxlength="320" /></div>
            </div></template>
          </Card>
          <div class="stack">
            <Card>
              <template #title><i class="pi pi-user" aria-hidden="true" /> Responsable y publicación</template>
              <template #content><div class="stack">
                <template v-if="admin">
                  <div class="form-field"><label for="manager-search">Buscar un encargado disponible</label><div class="cluster"><InputText id="manager-search" v-model.trim="managerSearch" placeholder="Nombre o correo" @keydown.enter.prevent="searchManagers" /><Button label="Buscar" icon="pi pi-search" :loading="managerLoading" @click="searchManagers" /></div></div>
                  <Message v-if="managerError" severity="error" :closable="false">{{ managerError }}</Message>
                  <div class="form-field"><label for="agency-manager">Encargado de agencia *</label><Select input-id="agency-manager" v-model="form.manager_id" :options="managerChoices" option-label="label" option-value="id" placeholder="Selecciona un encargado" :disabled="locked" /></div>
                  <small class="form-help">Cada encargado puede gestionar una sola agencia. Se muestran hasta 50 resultados; usa la búsqueda para encontrarlo.</small>
                  <NuxtLink to="/admin/users">Crear o gestionar encargados</NuxtLink>
                  <div class="form-field"><label for="agency-state">Estado</label><Select input-id="agency-state" v-model="form.status" :options="states" option-label="label" option-value="value" :disabled="locked" @change="hideInactive" /></div>
                </template>
                <p v-else class="muted">Encargado: {{ record?.manager_name }}</p>
                <div class="agency-check"><Checkbox input-id="agency-published" v-model="form.published" binary :disabled="locked || form.status === 'inactive'" /><label for="agency-published">Visible para turistas</label></div>
                <small class="form-help">Una agencia desactivada permanece oculta.</small>
              </div></template>
            </Card>
            <Card v-if="!admin">
              <template #title><i class="pi pi-users" aria-hidden="true" /> Política de menores</template>
              <template #content><div class="stack">
                <div class="form-field"><label for="agency-age">Edad mínima de pago *</label><InputNumber input-id="agency-age" v-model="form.minimum_paying_age" :min="0" :max="18" :use-grouping="false" :disabled="locked" show-buttons /></div>
                <p class="muted">{{ ageExplanation }}</p>
                <small class="form-help">Se registran las edades de los menores. El total de viajeros y los cupos se contabilizan por separado.</small>
              </div></template>
            </Card>
          </div>
        </div>
        <Card v-if="!admin" class="agency-payments">
          <template #title><i class="pi pi-wallet" aria-hidden="true" /> Medios de pago</template>
          <template #subtitle>Configura cómo recibirás el pago completo de los paquetes, en bolivianos.</template>
          <template #content>
            <div class="agency-editor-grid">
              <div class="stack">
                <div class="agency-check"><Checkbox input-id="agency-accepts-qr" v-model="form.accepts_qr" binary :disabled="locked" /><label for="agency-accepts-qr">Aceptar pago con QR</label></div>
                <div class="form-field"><label for="agency-qr">Imagen del QR de cobro</label><input id="agency-qr" type="file" accept="image/png,image/jpeg" @change="readQR"><small class="form-help">PNG o JPG, hasta 500 KB y 2048 × 2048 píxeles.</small></div>
                <Message v-if="qrError" severity="error" :closable="false">{{ qrError }}</Message>
                <img v-if="form.qr_image" class="agency-qr" :src="form.qr_image" alt="Vista previa del QR de cobro de la agencia">
                <Button v-if="form.qr_image" label="Quitar QR" icon="pi pi-trash" text severity="danger" :disabled="locked" @click="removeQR" />
              </div>
              <div class="stack">
                <div class="agency-check"><Checkbox input-id="agency-accepts-transfer" v-model="form.accepts_transfer" binary :disabled="locked" /><label for="agency-accepts-transfer">Aceptar transferencia bancaria</label></div>
                <div class="form-field"><label for="agency-bank">Banco{{ form.accepts_transfer ? ' *' : '' }}</label><InputText id="agency-bank" v-model.trim="form.bank_name" :required="form.accepts_transfer" maxlength="100" /></div>
                <div class="form-field"><label for="agency-holder">Titular de la cuenta{{ form.accepts_transfer ? ' *' : '' }}</label><InputText id="agency-holder" v-model.trim="form.account_holder" :required="form.accepts_transfer" maxlength="160" /></div>
                <div class="form-field"><label for="agency-account">Número de cuenta{{ form.accepts_transfer ? ' *' : '' }}</label><InputText id="agency-account" v-model.trim="form.account_number" :required="form.accepts_transfer" maxlength="50" /></div>
              </div>
            </div>
            <div class="form-field" style="margin-top: var(--space-5)"><label for="agency-payment-notes">Instrucciones de pago</label><Textarea id="agency-payment-notes" v-model.trim="form.payment_instructions" rows="3" maxlength="1000" /></div>
            <Message v-if="!form.accepts_qr && !form.accepts_transfer" severity="info" :closable="false">Puedes guardar la agencia ahora y configurar sus medios de pago antes de vender paquetes.</Message>
          </template>
        </Card>
      </fieldset>
      <div class="agency-actions">
        <span class="muted">{{ dirty ? 'Tienes cambios sin guardar' : 'Datos actualizados' }}</span>
        <div class="cluster">
          <Button v-if="!create" label="Recargar datos" icon="pi pi-refresh" outlined severity="secondary" :disabled="saving" @click="reload" />
          <Button type="submit" :label="create ? 'Crear agencia' : 'Guardar cambios'" icon="pi pi-save" :loading="saving" :disabled="locked || readingQR || (!create && !dirty)" />
        </div>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
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
import type { Agency, AgencyManager } from '~/types/agency'
import { apiError } from '~/utils/api-error'

const props = defineProps<{ admin?: boolean, create?: boolean, agencyId?: number }>()
const emit = defineEmits<{ saved: [agency: Agency] }>()
const auth = useAuthStore()
const toast = useToast()
const confirm = useConfirm()
const loading = ref(true)
const ready = ref(false)
const notAssigned = ref(false)
const saving = ref(false)
const error = ref('')
const fieldErrors = ref<Record<string, string>>({})
const departments = ref<string[]>([])
const record = ref<Agency | null>(null)
const managers = ref<AgencyManager[]>([])
const managerSearch = ref('')
const managerLoading = ref(false)
const managerError = ref('')
const readingQR = ref(false)
const qrError = ref('')
const states = [{ label: 'Activa', value: 'active' }, { label: 'Inactiva', value: 'inactive' }]
const form = reactive({ name: '', description: '', department: 'Tarija', city: '', address: '', phone: '', email: '', manager_id: null as number | null, status: 'active', published: false, minimum_paying_age: 6 as number | null, accepts_qr: false, accepts_transfer: false, bank_name: '', account_holder: '', account_number: '', payment_instructions: '', qr_image: '', version: 0 })
const initial = ref('')
const dirty = computed(() => ready.value && JSON.stringify(form) !== initial.value)
const locked = computed(() => saving.value || (!props.admin && form.status === 'inactive'))
const ageExplanation = computed(() => form.minimum_paying_age === 0 ? 'Todos los viajeros pagan y descuentan cupo.' : `Menores de ${form.minimum_paying_age ?? '…'} años no pagan ni descuentan cupo. Desde esa edad, pagan y descuentan cupo.`)
const managerChoices = computed(() => {
  const choices = managers.value.map(manager => ({ id: manager.id, label: `${manager.first_name} ${manager.last_name} · ${manager.email}` }))
  if (record.value && !choices.some(manager => manager.id === record.value?.manager_id)) choices.unshift({ id: record.value.manager_id, label: `${record.value.manager_name} · ${record.value.manager_email}${record.value.manager_status === 'inactive' ? ' (inactivo)' : ''}` })
  return choices
})
const endpoint = computed(() => props.admin ? props.create ? '/admin/agencies' : `/admin/agencies/${props.agencyId}` : '/agency')
const apply = (agency: Agency) => {
  record.value = agency
  Object.assign(form, { name: agency.name, description: agency.description, department: agency.department, city: agency.city, address: agency.address, phone: agency.phone, email: agency.email, manager_id: agency.manager_id, status: agency.status, published: agency.published, minimum_paying_age: agency.minimum_paying_age, accepts_qr: agency.accepts_qr, accepts_transfer: agency.accepts_transfer, bank_name: agency.bank_name, account_holder: agency.account_holder, account_number: agency.account_number, payment_instructions: agency.payment_instructions, qr_image: agency.qr_image || '', version: agency.version })
  initial.value = JSON.stringify(form)
}
let managerRequest = 0
const searchManagers = async () => {
  const current = ++managerRequest
  managerLoading.value = true; managerError.value = ''
  try {
    const response = await auth.request<AgencyManager[]>(`/admin/agencies/managers?search=${encodeURIComponent(managerSearch.value)}`)
    if (current === managerRequest) managers.value = response.data || []
  } catch (cause) { if (current === managerRequest) managerError.value = apiError(cause).message }
  finally { if (current === managerRequest) managerLoading.value = false }
}
const load = async () => {
  loading.value = true; ready.value = false; notAssigned.value = false; error.value = ''; fieldErrors.value = {}; qrError.value = ''
  try {
    const options = await auth.request<{ departments: string[] }>(props.admin ? '/admin/agencies/options' : '/agency/options')
    departments.value = options.data?.departments || []
    if (!props.create) {
      const response = await auth.request<Agency>(endpoint.value)
      if (!response.data) throw new Error('Missing agency')
      apply(response.data)
    }
    if (props.admin) await searchManagers()
    initial.value = JSON.stringify(form); ready.value = true
  } catch (cause) {
    const failure = apiError(cause)
    if (!props.admin && failure.code === 'AGENCY_NOT_FOUND') notAssigned.value = true
    else error.value = failure.message
  } finally { loading.value = false }
}
const reload = () => {
  if (!dirty.value) { void load(); return }
  confirm.require({ header: 'Recargar agencia', message: 'Se descartarán tus cambios sin guardar.', acceptLabel: 'Recargar', rejectLabel: 'Seguir editando', accept: () => { void load() } })
}
const hideInactive = () => { if (form.status === 'inactive') form.published = false }
const removeQR = () => { form.qr_image = ''; form.accepts_qr = false; qrError.value = '' }
const readQR = async (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  qrError.value = ''
  if (!['image/png', 'image/jpeg'].includes(file.type) || file.size > 500 * 1024) { qrError.value = 'Selecciona un PNG o JPG de hasta 500 KB.'; input.value = ''; return }
  readingQR.value = true
  try {
    form.qr_image = await new Promise<string>((resolve, reject) => { const reader = new FileReader(); reader.onload = () => resolve(String(reader.result)); reader.onerror = reject; reader.readAsDataURL(file) })
  } catch { qrError.value = 'No se pudo leer la imagen.' }
  finally { readingQR.value = false; input.value = '' }
}
const submit = async () => {
  if (locked.value || readingQR.value) return
  error.value = ''; fieldErrors.value = {}
  if (props.admin && !form.manager_id) { error.value = 'Selecciona un encargado de agencia disponible.'; return }
  if (!props.admin && (form.minimum_paying_age === null || !Number.isInteger(form.minimum_paying_age) || form.minimum_paying_age < 0 || form.minimum_paying_age > 18)) { error.value = 'La edad mínima debe estar entre 0 y 18 años.'; return }
  if (!props.admin && form.accepts_qr && !form.qr_image) { error.value = 'Sube la imagen QR antes de habilitar ese medio de pago.'; return }
  saving.value = true
  try {
    const { manager_id, status, ...editable } = form
    const body = props.admin
      ? { name: form.name, description: form.description, department: form.department, city: form.city, address: form.address, phone: form.phone, email: form.email, published: form.published, version: form.version, manager_id, status }
      : editable
    const response = await auth.request<Agency>(endpoint.value, { method: props.create ? 'POST' : 'PUT', body })
    if (!response.data) throw new Error('Missing agency')
    apply(response.data)
    toast.add({ severity: 'success', summary: props.create ? 'Agencia creada' : 'Agencia actualizada', life: 4000 })
    emit('saved', response.data)
  } catch (cause) { const failure = apiError(cause); error.value = failure.message; fieldErrors.value = failure.fields || {} }
  finally { saving.value = false }
}
onBeforeRouteLeave(() => {
  if (!dirty.value) return true
  return new Promise<boolean>(resolve => confirm.require({ header: 'Cambios sin guardar', message: '¿Salir sin guardar los cambios de la agencia?', acceptLabel: 'Salir sin guardar', rejectLabel: 'Seguir editando', accept: () => resolve(true), reject: () => resolve(false), onHide: () => resolve(false) }))
})
onMounted(load)
</script>

<style scoped>
.agency-fieldset { border: 0; padding: 0; margin: 0; min-width: 0; }
.agency-editor-grid { display: grid; grid-template-columns: minmax(0,1fr) minmax(0,1fr); gap: var(--space-5); }
.agency-payments { margin-top: var(--space-5); }
.agency-check { display: flex; align-items: center; gap: .75rem; }
.agency-check label { cursor: pointer; }
.agency-qr { width: 100%; max-width: 17rem; max-height: 20rem; object-fit: contain; background: white; padding: .75rem; border-radius: var(--radius-md); }
.agency-actions { display: flex; flex-wrap: wrap; justify-content: space-between; align-items: center; gap: var(--space-4); padding: var(--space-5); border: 1px solid var(--color-border); border-radius: var(--radius-md); background: var(--color-surface); }
input[type=file] { max-width: 100%; padding: .75rem; border: 1px solid var(--color-border); border-radius: var(--radius-sm); }
input[type=file]::file-selector-button { padding: .5rem; margin-right: .5rem; cursor: pointer; }
@media(max-width: 900px) { .agency-editor-grid { grid-template-columns: 1fr; } }
</style>
