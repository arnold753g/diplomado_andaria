<template>
  <section class="purchase-page stack">
    <NuxtLink :to="`/packages/${route.params.id}`" class="back-link"><i class="pi pi-arrow-left" aria-hidden="true" /> Volver al paquete</NuxtLink>
    <UiLoadingState v-if="loading" />
    <Message v-else-if="loadError" severity="error" :closable="false">{{ loadError }} <Button label="Reintentar" text @click="load" /></Message>
    <Card v-else-if="created" class="success-card">
      <template #title>Compra enviada a revisión</template>
      <template #content>
        <div class="stack">
          <Message severity="success" :closable="false">La agencia recibió tu comprobante. Los {{ created.capacity_count }} cupos pagados quedaron descontados mientras se revisa el pago.</Message>
          <dl class="summary-list"><dt>Referencia</dt><dd><strong>{{ created.reference }}</strong></dd><dt>Paquete</dt><dd>{{ created.package_name }}</dd><dt>Salida</dt><dd>{{ departureDate(created.departure_start) }}</dd><dt>Total pagado</dt><dd>{{ packageMoney(created.total_cents) }}</dd></dl>
          <NuxtLink class="primary-link" to="/app/purchases">Ver mis compras</NuxtLink>
        </div>
      </template>
    </Card>
    <template v-else-if="item && options">
      <header class="purchase-header"><div><span class="auth-eyebrow">Compra segura</span><h1>{{ item.name }}</h1><p>Selecciona una salida, registra a tu grupo y adjunta el comprobante del pago completo.</p></div><div class="price-block"><span>Tarifa nacional</span><strong>{{ packageMoney(item.national_price_cents) }}</strong><small>por persona que paga</small></div></header>

      <form class="purchase-layout" @submit.prevent="submit">
        <div class="stack">
          <Card><template #title>1. Salida</template><template #content><div class="form-field"><label for="purchase-departure">Fecha disponible *</label><Select id="purchase-departure" v-model="departureID" :options="departureOptions" option-label="label" option-value="value" placeholder="Selecciona una salida" fluid /></div><Message v-if="selectedDeparture && !selectedDeparture.minimum_reached" severity="warn" :closable="false">Esta salida todavía necesita {{ selectedDeparture.remaining_for_minimum }} participantes confirmados. Las compras se acumulan y, si no alcanza el mínimo, corresponderá la devolución.</Message></template></Card>

          <Card><template #title>2. Participantes</template><template #content><div class="participant-grid"><div class="form-field"><label for="purchase-national">Adultos nacionales *</label><InputNumber input-id="purchase-national" v-model="nationalAdults" :min="0" :max="100" show-buttons fluid /></div><div class="form-field"><label for="purchase-foreign">Adultos extranjeros *</label><InputNumber input-id="purchase-foreign" v-model="foreignAdults" :min="0" :max="100" show-buttons fluid /><small>Se suma {{ packageMoney(item.foreign_surcharge_cents) }} por extranjero.</small></div></div><div class="minor-heading"><div><strong>Menores</strong><p>Registra solo su edad. Los menores de {{ options.minimum_paying_age }} años no pagan ni descuentan cupo.</p></div><Button type="button" label="Agregar menor" icon="pi pi-plus" outlined @click="addMinor" /></div><div v-if="minors.length" class="minor-list"><div v-for="(minor, index) in minors" :key="index" class="minor-row"><div class="form-field"><label :for="`minor-age-${index}`">Edad</label><InputNumber :input-id="`minor-age-${index}`" v-model="minor.age" :min="0" :max="17" show-buttons /></div><label class="foreign-check"><Checkbox v-model="minor.is_foreign" binary /> Extranjero</label><Tag :value="minorPays(minor) ? 'Paga y ocupa cupo' : 'No paga ni ocupa cupo'" :severity="minorPays(minor) ? 'info' : 'secondary'" /><Button type="button" icon="pi pi-trash" aria-label="Quitar menor" severity="danger" text @click="minors.splice(index, 1)" /></div></div></template></Card>

          <Card><template #title>3. Pago y comprobante</template><template #content><div class="form-field"><label for="purchase-method">Medio de pago *</label><Select id="purchase-method" v-model="paymentMethod" :options="methodOptions" option-label="label" option-value="value" placeholder="Selecciona un medio" fluid /></div><div v-if="paymentMethod === 'transfer'" class="payment-data"><span>Banco</span><strong>{{ options.bank_name }}</strong><span>Titular</span><strong>{{ options.account_holder }}</strong><span>Número de cuenta</span><strong>{{ options.account_number }}</strong></div><img v-if="paymentMethod === 'qr' && options.qr_image" class="payment-qr" :src="options.qr_image" :alt="`QR de cobro de ${options.agency_name}`"><p v-if="options.payment_instructions" class="payment-instructions">{{ options.payment_instructions }}</p><div class="form-field"><label for="purchase-proof">Comprobante de pago *</label><input id="purchase-proof" type="file" accept="image/png,image/jpeg" @change="readProof"><small>PNG o JPG de hasta 5 MB.</small><span v-if="proofName" class="file-ready"><i class="pi pi-check-circle" aria-hidden="true" /> {{ proofName }}</span></div></template></Card>
        </div>

        <aside class="purchase-summary stack">
          <div><span class="auth-eyebrow">Resumen</span><h2>Tu compra</h2></div>
          <dl class="summary-list"><dt>Adultos nacionales</dt><dd>{{ nationalAdults || 0 }}</dd><dt>Adultos extranjeros</dt><dd>{{ foreignAdults || 0 }}</dd><dt>Menores que pagan</dt><dd>{{ payingMinors }}</dd><dt>Menores gratuitos</dt><dd>{{ freeMinors }}</dd><dt>Cupos descontados</dt><dd>{{ capacityCount }}</dd><template v-if="selectedDeparture"><dt>Cupos disponibles</dt><dd>{{ selectedDeparture.available_capacity }}</dd></template></dl>
          <div class="total"><span>Total a pagar</span><strong>{{ packageMoney(totalCents) }}</strong><small>Pago del 100 %</small></div>
          <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>
          <Button type="submit" label="Enviar compra y comprobante" icon="pi pi-send" fluid :loading="saving" :disabled="!canSubmit" />
          <small>El servidor volverá a comprobar cupos, horario y precio antes de registrar la compra.</small>
        </aside>
      </form>
    </template>
  </section>
</template>

<script setup lang="ts">
import Button from 'primevue/button'
import Card from 'primevue/card'
import Checkbox from 'primevue/checkbox'
import InputNumber from 'primevue/inputnumber'
import Message from 'primevue/message'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import type { ApiEnvelope } from '~/types/api'
import type { PublicPackageDetail } from '~/types/package'
import type { PaymentMethod, Purchase, PurchasePaymentOptions } from '~/types/purchase'
import { apiError } from '~/utils/api-error'
import { packageMoney } from '~/utils/packages'
import { paymentMethodLabel } from '~/utils/purchases'

definePageMeta({ layout: 'catalog', middleware: 'tourist', validate: route => /^[1-9]\d*$/.test(String(route.params.id)) })
useHead({ title: 'Comprar paquete' })
const route = useRoute(); const config = useRuntimeConfig(); const auth = useAuthStore()
const item = ref<PublicPackageDetail | null>(null); const options = ref<PurchasePaymentOptions | null>(null)
const loading = ref(true); const loadError = ref(''); const error = ref(''); const saving = ref(false); const created = ref<Purchase | null>(null)
const departureID = ref<number | null>(null); const nationalAdults = ref<number | null>(1); const foreignAdults = ref<number | null>(0)
const minors = reactive<Array<{ age: number | null; is_foreign: boolean }>>([]); const paymentMethod = ref<PaymentMethod | null>(null)
const proof = ref(''); const proofName = ref('')
const departureDate = (value: string) => new Intl.DateTimeFormat('es-BO', { dateStyle: 'full', timeStyle: 'short', timeZone: 'America/La_Paz' }).format(new Date(value))
const departureOptions = computed(() => (item.value?.departures || []).filter(value => value.bookable).map(value => ({ value: value.id, label: `${departureDate(value.starts_at)} · ${value.available_capacity} cupos` })))
const selectedDeparture = computed(() => item.value?.departures.find(value => value.id === departureID.value) || null)
const methodOptions = computed(() => (options.value?.methods || []).map(value => ({ value, label: paymentMethodLabel(value) })))
const minorPays = (minor: { age: number | null }) => minor.age !== null && minor.age >= (options.value?.minimum_paying_age ?? 18)
const payingMinors = computed(() => minors.filter(minorPays).length)
const freeMinors = computed(() => minors.length - payingMinors.value)
const capacityCount = computed(() => (nationalAdults.value || 0) + (foreignAdults.value || 0) + payingMinors.value)
const foreignPayingMinors = computed(() => minors.filter(minor => minorPays(minor) && minor.is_foreign).length)
const totalCents = computed(() => item.value ? capacityCount.value * item.value.national_price_cents + ((foreignAdults.value || 0) + foreignPayingMinors.value) * item.value.foreign_surcharge_cents : 0)
const canSubmit = computed(() => Boolean(selectedDeparture.value?.bookable && capacityCount.value > 0 && capacityCount.value <= (selectedDeparture.value?.available_capacity || 0) && paymentMethod.value && proof.value && !minors.some(minor => minor.age === null)))
const addMinor = () => { if (minors.length < 100) minors.push({ age: null, is_foreign: false }) }
const readProof = async (event: Event) => {
  error.value = ''; proof.value = ''; proofName.value = ''
  const input = event.target as HTMLInputElement; const file = input.files?.[0]
  if (!file) return
  if (!['image/png', 'image/jpeg'].includes(file.type) || file.size > 5 * 1024 * 1024) { error.value = 'Selecciona un comprobante PNG o JPG de hasta 5 MB.'; input.value = ''; return }
  proof.value = await new Promise<string>((resolve, reject) => { const reader = new FileReader(); reader.onload = () => resolve(String(reader.result)); reader.onerror = reject; reader.readAsDataURL(file) })
  proofName.value = file.name
}
const load = async () => {
  loading.value = true; loadError.value = ''
  try {
    const [detail, payment] = await Promise.all([
      $fetch<ApiEnvelope<PublicPackageDetail>>(`${config.public.apiBase}/packages/${route.params.id}`),
      auth.request<PurchasePaymentOptions>(`/me/packages/${route.params.id}/payment-options`)
    ])
    if (!detail.data || !payment.data) throw new Error('Missing purchase data')
    item.value = detail.data; options.value = payment.data
    const requested = Number(route.query.departure)
    departureID.value = detail.data.departures.some(value => value.id === requested && value.bookable) ? requested : departureOptions.value[0]?.value || null
    paymentMethod.value = payment.data.methods[0] || null
    if (!payment.data.methods.length) loadError.value = 'La agencia todavía no configuró un medio de pago.'
  } catch (cause) { loadError.value = apiError(cause, 'No se pudo preparar la compra').message }
  finally { loading.value = false }
}
const submit = async () => {
  if (!canSubmit.value || !departureID.value || !paymentMethod.value) { error.value = 'Completa la salida, participantes, medio de pago y comprobante.'; return }
  saving.value = true; error.value = ''
  try {
    const response = await auth.request<Purchase>('/me/purchases', { method: 'POST', body: { departure_id: departureID.value, payment_method: paymentMethod.value, national_adults: nationalAdults.value || 0, foreign_adults: foreignAdults.value || 0, minors: minors.map(minor => ({ age: minor.age, is_foreign: minor.is_foreign })), payment_proof: proof.value } })
    if (!response.data) throw new Error('Missing purchase')
    created.value = response.data
  } catch (cause) { error.value = apiError(cause, 'No se pudo registrar la compra').message }
  finally { saving.value = false }
}
onMounted(load)
</script>

<style scoped>
.back-link { display:inline-flex; align-items:center; gap:.5rem; width:max-content; text-decoration:none; }.purchase-header { display:flex; align-items:end; justify-content:space-between; gap:2rem; }.purchase-header h1 { margin:.35rem 0; font-size:clamp(2rem,4vw,3.4rem); }.purchase-header p { max-width:60ch; color:var(--color-text-muted); }.price-block { display:grid; min-width:14rem; padding:1.2rem; border:1px solid rgba(186,254,6,.25); border-radius:var(--radius-lg); background:var(--color-primary-soft); }.price-block strong { color:var(--color-primary); font-size:1.8rem; }.price-block span,.price-block small { color:var(--color-text-muted); }.purchase-layout { display:grid; grid-template-columns:minmax(0,1.5fr) minmax(18rem,.6fr); gap:1.5rem; align-items:start; }.participant-grid { display:grid; grid-template-columns:1fr 1fr; gap:1rem; }.minor-heading { display:flex; align-items:center; justify-content:space-between; gap:1rem; margin-top:1.5rem; }.minor-heading p { margin:.3rem 0 0; color:var(--color-text-muted); }.minor-list { display:grid; gap:.75rem; margin-top:1rem; }.minor-row { display:grid; grid-template-columns:minmax(8rem,.6fr) 1fr auto auto; align-items:end; gap:.75rem; padding:.75rem; border:1px solid var(--color-border); border-radius:var(--radius-sm); }.foreign-check { display:flex; align-items:center; gap:.5rem; min-height:2.7rem; }.payment-data,.summary-list { display:grid; grid-template-columns:1fr 1.2fr; gap:.55rem 1rem; margin:1rem 0; }.payment-data span,.summary-list dt { color:var(--color-text-muted); }.summary-list dd { margin:0; text-align:right; }.payment-qr { display:block; width:min(100%,22rem); margin:1rem auto; border-radius:var(--radius-md); }.payment-instructions { padding:1rem; border-radius:var(--radius-sm); background:var(--color-background-deep); white-space:pre-wrap; }.file-ready { display:block; margin-top:.5rem; color:var(--color-success); }.purchase-summary { position:sticky; top:1rem; padding:1.25rem; border:1px solid var(--color-border); border-radius:var(--radius-xl); background:var(--color-surface); }.purchase-summary h2 { margin:.2rem 0; }.total { display:grid; padding:1rem; border-radius:var(--radius-md); background:var(--color-primary-soft); }.total strong { color:var(--color-primary); font-size:2rem; }.total small { color:var(--color-text-muted); }.primary-link { display:flex; justify-content:center; padding:.8rem 1rem; border-radius:var(--radius-sm); background:var(--color-primary); color:#111; font-weight:800; text-decoration:none; }.success-card { max-width:48rem; margin:2rem auto; }
@media(max-width:900px) { .purchase-layout { grid-template-columns:1fr; }.purchase-summary { position:static; }.purchase-header { align-items:start; flex-direction:column; }.price-block { width:100%; } }
@media(max-width:650px) { .participant-grid,.minor-row { grid-template-columns:1fr; }.minor-heading { align-items:start; flex-direction:column; }.minor-row :deep(.p-tag) { width:max-content; }.payment-data { grid-template-columns:1fr; gap:.2rem; }.payment-data strong { margin-bottom:.6rem; } }
</style>
