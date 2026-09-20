<template>
  <section class="stack">
    <header class="page-heading">
      <div><span class="auth-eyebrow">Historial personal</span><h1>Mis compras</h1><p>Consulta pagos, cancelaciones y devoluciones de cada salida.</p></div>
      <NuxtLink class="browse-link" to="/packages"><i class="pi pi-compass" aria-hidden="true" /> Explorar paquetes</NuxtLink>
    </header>

    <UiLoadingState v-if="loading" />
    <Message v-else-if="error" severity="error" :closable="false">{{ error }} <Button label="Reintentar" text @click="load" /></Message>
    <UiEmptyState v-else-if="!purchases.length" icon="pi pi-shopping-bag" title="Todavía no tienes compras" description="Elige una salida disponible y envía tu comprobante de pago."><NuxtLink class="browse-link" to="/packages">Ver paquetes</NuxtLink></UiEmptyState>

    <div v-else class="purchase-list">
      <Card v-for="purchase in purchases" :key="purchase.id" class="purchase-card">
        <template #title>
          <div class="card-title"><div><small>{{ purchase.reference }}</small><h2>{{ purchase.package_name }}</h2></div><Tag :value="purchaseStatus(purchase.status)" :severity="purchaseSeverity(purchase.status)" /></div>
        </template>
        <template #content>
          <div class="card-content">
            <dl><dt>Agencia</dt><dd>{{ purchase.agency_name }}</dd><dt>Salida</dt><dd>{{ dateTime(purchase.departure_start) }}</dd><dt>Participantes que ocupan cupo</dt><dd>{{ purchase.capacity_count }}</dd><dt>Menores gratuitos registrados</dt><dd>{{ purchase.free_minor_count }}</dd><dt>Medio de pago</dt><dd>{{ paymentMethodLabel(purchase.payment_method) }}</dd><dt>Total</dt><dd><strong>{{ packageMoney(purchase.total_cents) }}</strong></dd></dl>

            <Message v-if="purchase.status === 'payment_review'" severity="warn" :closable="false">La agencia está revisando el comprobante. Los cupos ya fueron descontados.</Message>
            <div v-if="purchase.status === 'correction_requested'" class="action-box">
              <Message severity="info" :closable="false">{{ purchase.rejection_reason }} Los cupos se mantienen mientras corriges el respaldo.</Message>
              <div class="form-field"><label :for="`proof-${purchase.id}`">Nuevo comprobante</label><input :id="`proof-${purchase.id}`" type="file" accept="image/png,image/jpeg" @change="readPaymentProof(purchase.id, $event)"></div>
              <Button label="Volver a enviar" icon="pi pi-send" :loading="saving === purchase.id" :disabled="!paymentProofs[purchase.id]" @click="updateProof(purchase)" />
            </div>
            <Message v-if="purchase.status === 'payment_rejected'" severity="error" :closable="false">{{ purchase.rejection_reason }}</Message>

            <div v-if="purchase.cancellable && !cancelOpen[purchase.id]" class="cancel-summary">
              <p>Esta compra puede cancelarse hasta {{ dateTime(purchase.cancellation_deadline!) }} con devolución del 100 %.</p>
              <Button label="Cancelar compra" icon="pi pi-times" severity="danger" outlined @click="openCancellation(purchase)" />
            </div>
            <div v-if="cancelOpen[purchase.id]" class="action-box">
              <Message severity="warn" :closable="false">Al confirmar, los cupos se liberarán y la agencia tendrá 72 horas para realizar la devolución.</Message>
              <div class="form-field"><label :for="`cancel-reason-${purchase.id}`">Motivo de cancelación *</label><Textarea :id="`cancel-reason-${purchase.id}`" v-model.trim="refundForm(purchase.id).reason" rows="3" maxlength="1000" /></div>
              <RefundDestinationForm :id-prefix="`cancel-${purchase.id}`" :form="refundForm(purchase.id)" @read-qr="readRefundQR(purchase.id, $event)" />
              <div class="cluster"><Button label="Volver" severity="secondary" text :disabled="saving === purchase.id" @click="cancelOpen[purchase.id] = false" /><Button label="Confirmar cancelación" icon="pi pi-times" severity="danger" :loading="saving === purchase.id" @click="cancelPurchase(purchase)" /></div>
            </div>

            <div v-if="purchase.status === 'refund_pending'" class="action-box">
              <Message :severity="purchase.refund_overdue ? 'error' : 'warn'" :closable="false">{{ purchase.refund_reason }} <span v-if="purchase.refund_due_at">La devolución debe registrarse hasta {{ dateTime(purchase.refund_due_at) }}.</span></Message>
              <template v-if="purchase.refund_method && !destinationOpen[purchase.id]">
                <p><strong>Destino:</strong> {{ refundMethodLabel(purchase.refund_method) }}</p>
                <img v-if="purchase.refund_method === 'qr'" class="refund-image" :src="refundQRURL(purchase.id)" :alt="`QR para el reembolso ${purchase.reference}`">
                <dl v-else class="refund-bank"><dt>Banco</dt><dd>{{ purchase.refund_bank_name }}</dd><dt>Titular</dt><dd>{{ purchase.refund_account_holder }}</dd><dt>Cuenta</dt><dd>{{ purchase.refund_account_number }}</dd></dl>
                <Button label="Cambiar datos de devolución" icon="pi pi-pencil" outlined @click="openDestination(purchase)" />
              </template>
              <template v-else>
                <Message v-if="!purchase.refund_method" severity="info" :closable="false">Indica dónde debe realizar la agencia la devolución.</Message>
                <RefundDestinationForm :id-prefix="`refund-${purchase.id}`" :form="refundForm(purchase.id)" @read-qr="readRefundQR(purchase.id, $event)" />
                <div class="cluster"><Button v-if="purchase.refund_method" label="Volver" severity="secondary" text @click="destinationOpen[purchase.id] = false" /><Button label="Guardar datos" icon="pi pi-save" :loading="saving === purchase.id" @click="saveDestination(purchase)" /></div>
              </template>
            </div>

            <div v-if="purchase.status === 'refunded'" class="action-box">
              <Message severity="success" :closable="false">La agencia registró la devolución el {{ dateTime(purchase.refunded_at!) }}.</Message>
              <p v-if="purchase.refund_reference"><strong>Referencia:</strong> {{ purchase.refund_reference }}</p>
              <img v-if="purchase.has_refund_proof" class="refund-image proof" :src="refundProofURL(purchase.id)" :alt="`Comprobante de devolución ${purchase.reference}`">
            </div>
            <p v-if="purchase.meeting_point"><strong>Punto de encuentro:</strong> {{ purchase.meeting_point }}</p>
          </div>
        </template>
      </Card>
    </div>
  </section>
</template>

<script setup lang="ts">
import Button from 'primevue/button'
import Card from 'primevue/card'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import Textarea from 'primevue/textarea'
import type { Purchase, PurchasePage } from '~/types/purchase'
import { apiError } from '~/utils/api-error'
import { packageMoney } from '~/utils/packages'
import { paymentMethodLabel, purchaseSeverity, purchaseStatus, refundMethodLabel } from '~/utils/purchases'

definePageMeta({ layout: 'app', middleware: 'tourist' })
useHead({ title: 'Mis compras' })
type RefundFormState = { method: 'qr' | 'bank_transfer'; qr: string; bankName: string; holder: string; account: string; reason: string }
const auth = useAuthStore(); const config = useRuntimeConfig()
const purchases = ref<Purchase[]>([]); const loading = ref(true); const error = ref(''); const saving = ref<number | null>(null)
const paymentProofs = reactive<Record<number, string>>({}); const cancelOpen = reactive<Record<number, boolean>>({}); const destinationOpen = reactive<Record<number, boolean>>({}); const refundForms = reactive<Record<number, RefundFormState>>({})
const dateTime = (value: string) => new Intl.DateTimeFormat('es-BO', { dateStyle: 'medium', timeStyle: 'short', timeZone: 'America/La_Paz' }).format(new Date(value))
const refundQRURL = (id: number) => `${config.public.apiBase}/me/purchases/${id}/refund-qr`
const refundProofURL = (id: number) => `${config.public.apiBase}/me/purchases/${id}/refund-proof`
const refundForm = (id: number) => refundForms[id] ||= { method: 'qr', qr: '', bankName: '', holder: '', account: '', reason: '' }
const load = async () => { loading.value = true; error.value = ''; try { const response = await auth.request<PurchasePage>('/me/purchases?limit=50'); purchases.value = response.data?.purchases || [] } catch (cause) { error.value = apiError(cause, 'No se pudieron cargar tus compras').message } finally { loading.value = false } }
const readImage = async (event: Event) => { const input = event.target as HTMLInputElement; const file = input.files?.[0]; if (!file) return ''; if (!['image/png', 'image/jpeg'].includes(file.type) || file.size > 5 * 1024 * 1024) { error.value = 'Selecciona un PNG o JPG de hasta 5 MB.'; input.value = ''; return '' }; return await new Promise<string>((resolve, reject) => { const reader = new FileReader(); reader.onload = () => resolve(String(reader.result)); reader.onerror = reject; reader.readAsDataURL(file) }) }
const readPaymentProof = async (id: number, event: Event) => { const value = await readImage(event); if (value) paymentProofs[id] = value }
const readRefundQR = async (id: number, event: Event) => { const value = await readImage(event); if (value) refundForm(id).qr = value }
const destinationBody = (id: number) => { const form = refundForm(id); return { refund_method: form.method, refund_qr: form.method === 'qr' ? form.qr : '', refund_bank_name: form.method === 'bank_transfer' ? form.bankName.trim() : '', refund_account_holder: form.method === 'bank_transfer' ? form.holder.trim() : '', refund_account_number: form.method === 'bank_transfer' ? form.account.trim() : '' } }
const validDestination = (id: number) => { const form = refundForm(id); if (form.method === 'qr') return !!form.qr; return form.bankName.trim().length >= 2 && form.holder.trim().length >= 3 && form.account.trim().length >= 3 }
const openCancellation = (purchase: Purchase) => { refundForms[purchase.id] = { method: 'qr', qr: '', bankName: '', holder: '', account: '', reason: '' }; cancelOpen[purchase.id] = true }
const openDestination = (purchase: Purchase) => { refundForms[purchase.id] = { method: purchase.refund_method || 'qr', qr: '', bankName: purchase.refund_bank_name || '', holder: purchase.refund_account_holder || '', account: purchase.refund_account_number || '', reason: '' }; destinationOpen[purchase.id] = true }
const updateProof = async (purchase: Purchase) => { if (!paymentProofs[purchase.id]) return; saving.value = purchase.id; error.value = ''; try { await auth.request(`/me/purchases/${purchase.id}/proof`, { method: 'PATCH', body: { version: purchase.version, payment_proof: paymentProofs[purchase.id] } }); delete paymentProofs[purchase.id]; await load() } catch (cause) { error.value = apiError(cause, 'No se pudo actualizar el comprobante').message } finally { saving.value = null } }
const cancelPurchase = async (purchase: Purchase) => { const form = refundForm(purchase.id); if (form.reason.trim().length < 5 || !validDestination(purchase.id)) { error.value = 'Escribe un motivo y completa el QR o la cuenta para la devolución.'; return }; saving.value = purchase.id; error.value = ''; try { await auth.request(`/me/purchases/${purchase.id}/cancel`, { method: 'POST', body: { version: purchase.version, reason: form.reason.trim(), ...destinationBody(purchase.id) } }); cancelOpen[purchase.id] = false; await load() } catch (cause) { error.value = apiError(cause, 'No se pudo cancelar la compra').message } finally { saving.value = null } }
const saveDestination = async (purchase: Purchase) => { if (!validDestination(purchase.id)) { error.value = 'Completa el QR o los datos bancarios para la devolución.'; return }; saving.value = purchase.id; error.value = ''; try { await auth.request(`/me/purchases/${purchase.id}/refund-destination`, { method: 'PATCH', body: { version: purchase.version, ...destinationBody(purchase.id) } }); destinationOpen[purchase.id] = false; await load() } catch (cause) { error.value = apiError(cause, 'No se pudieron guardar los datos de devolución').message } finally { saving.value = null } }
onMounted(load)
</script>

<style scoped>
.page-heading { display:flex; align-items:end; justify-content:space-between; gap:1rem; }.page-heading h1 { margin:.25rem 0; font-size:clamp(2rem,4vw,3.3rem); }.page-heading p { margin:0; color:var(--color-text-muted); }.browse-link { display:inline-flex; align-items:center; justify-content:center; gap:.45rem; padding:.7rem 1rem; border:1px solid var(--color-primary); border-radius:var(--radius-sm); font-weight:700; text-decoration:none; }.purchase-list { display:grid; grid-template-columns:repeat(auto-fit,minmax(min(100%,24rem),1fr)); gap:1rem; }.card-title { display:flex; align-items:start; justify-content:space-between; gap:1rem; }.card-title small { color:var(--color-primary); letter-spacing:.04em; }.card-title h2 { margin:.25rem 0; font-size:1.35rem; }.card-content,.action-box { display:grid; gap:1rem; }.card-content dl,.refund-bank { display:grid; grid-template-columns:1fr 1.2fr; gap:.55rem 1rem; margin:0; }.card-content dt { color:var(--color-text-muted); }.card-content dd { margin:0; text-align:right; }.card-content p { margin:0; color:var(--color-text-muted); }.action-box,.cancel-summary { padding:1rem; border:1px solid var(--color-border); border-radius:var(--radius-sm); }.cancel-summary { display:grid; gap:.75rem; }.refund-image { width:100%; max-height:18rem; object-fit:contain; border:1px solid var(--color-border); border-radius:var(--radius-sm); background:#fff; }.proof { max-height:22rem; }
@media(max-width:650px) { .page-heading { align-items:start; flex-direction:column; }.browse-link { width:100%; }.card-title { flex-direction:column; }.card-content dl,.refund-bank { grid-template-columns:1fr; gap:.2rem; }.card-content dd { margin-bottom:.6rem; text-align:left; } }
</style>
