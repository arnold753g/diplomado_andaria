<template>
<div>
<Message v-if="error" severity="error" :closable="false">{{ error }}</Message>
      <Card class="review-card">
        <template #title><div class="card-title"><div><small>{{ purchase.reference }}</small><h2>{{ purchase.package_name }}</h2></div><Tag :value="purchaseStatus(purchase.status)" :severity="purchaseSeverity(purchase.status)" /></div></template>
        <template #content>
          <div class="review-content">
            <PaymentsReceiptActions v-if="purchase.receipt_available" :purchase-id="purchase.id" scope="agency" />
            <div><dl><dt>Turista</dt><dd>{{ purchase.tourist_name }}</dd><dt>Salida</dt><dd>{{ dateTime(purchase.departure_start) }}</dd><dt>Cupos</dt><dd>{{ purchase.capacity_count }}</dd><dt>Medio pagado</dt><dd>{{ paymentMethodLabel(purchase.payment_method) }}</dd><dt>Total</dt><dd><strong>{{ packageMoney(purchase.total_cents) }}</strong></dd></dl><p v-if="purchase.tourist_document"><strong>Documento:</strong> {{ purchase.tourist_document }}</p></div>
            <div><span class="image-label">Comprobante de pago</span><img loading="lazy" class="proof" :src="paymentProofURL(purchase.id)" :alt="`Comprobante ${purchase.reference}`"></div>

            <div v-if="purchase.status === 'payment_review'" class="review-actions">
              <div class="form-field"><label :for="`reason-${purchase.id}`">Motivo para corrección o reembolso</label><InputText :id="`reason-${purchase.id}`" v-model.trim="reason" maxlength="1000" placeholder="Ej. comprobante ilegible" /></div>
              <div class="cluster"><Button label="Confirmar pago" icon="pi pi-check" :loading="saving" @click="review('confirm')" /><Button label="Solicitar corrección" icon="pi pi-refresh" severity="info" outlined :loading="saving" @click="review('request_correction')" /><Button label="Enviar a reembolso" icon="pi pi-wallet" severity="danger" outlined :loading="saving" @click="review('refund')" /></div>
            </div>
            <Message v-if="purchase.status === 'correction_requested'" severity="info" :closable="false">Se solicitó un nuevo comprobante: {{ purchase.rejection_reason }}</Message>

            <div v-if="purchase.status === 'refund_pending'" class="refund-box">
              <Message :severity="purchase.refund_overdue ? 'error' : 'warn'" :closable="false">{{ purchase.refund_reason }} <span v-if="purchase.refund_due_at">Plazo: {{ dateTime(purchase.refund_due_at) }}.</span></Message>
              <Message v-if="!purchase.refund_method" severity="info" :closable="false">El turista todavía debe registrar su QR o cuenta bancaria. No se puede cerrar el reembolso hasta recibir esos datos.</Message>
              <template v-else>
                <h3>{{ refundMethodLabel(purchase.refund_method) }}</h3>
                <img v-if="purchase.refund_method === 'qr'" class="proof refund-qr" :src="refundQRURL(purchase.id)" :alt="`QR de devolución ${purchase.reference}`">
                <dl v-else class="bank-details"><dt>Banco</dt><dd>{{ purchase.refund_bank_name }}</dd><dt>Titular</dt><dd>{{ purchase.refund_account_holder }}</dd><dt>Cuenta</dt><dd>{{ purchase.refund_account_number }}</dd></dl>
                <div class="form-field"><label :for="`refund-proof-${purchase.id}`">Comprobante de devolución *</label><input :id="`refund-proof-${purchase.id}`" type="file" accept="image/png,image/jpeg" @change="readRefundProof($event)"><small>PNG o JPG de hasta 5 MB.</small></div>
                <div class="form-field"><label :for="`refund-reference-${purchase.id}`">Referencia de la operación</label><InputText :id="`refund-reference-${purchase.id}`" v-model.trim="refundReference" maxlength="160" /></div>
                <Button label="Marcar devolución completada" icon="pi pi-check-circle" :loading="saving" :disabled="!refundProof" @click="completeRefund()" />
              </template>
            </div>

            <div v-if="purchase.status === 'refunded'" class="refund-box">
              <Message severity="success" :closable="false">Devolución registrada el {{ dateTime(purchase.refunded_at!) }}.</Message>
              <p v-if="purchase.refund_reference"><strong>Referencia:</strong> {{ purchase.refund_reference }}</p>
              <img v-if="purchase.has_refund_proof" class="proof" :src="refundProofURL(purchase.id)" :alt="`Comprobante de devolución ${purchase.reference}`">
            </div>
          </div>
        </template>
      </Card>
</div>
</template>
<script setup lang="ts">
import Button from 'primevue/button'
import Card from 'primevue/card'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import PaymentsReceiptActions from '~/components/payments/ReceiptActions.vue'
import type { Purchase } from '~/types/purchase'
import { apiError } from '~/utils/api-error'
import { packageMoney } from '~/utils/packages'
import { paymentMethodLabel, purchaseSeverity, purchaseStatus, refundMethodLabel } from '~/utils/purchases'
const props = defineProps<{ purchase: Purchase }>()
const emit = defineEmits<{ updated: [] }>()
const auth = useAuthStore(); const config = useRuntimeConfig()
const reason = ref(''); const refundProof = ref(''); const refundReference = ref(''); const saving = ref(false); const error = ref('')
const dateTime = (value: string) => new Intl.DateTimeFormat('es-BO', { dateStyle: 'medium', timeStyle: 'short', timeZone: 'America/La_Paz' }).format(new Date(value))
const paymentProofURL = (id: number) => `${config.public.apiBase}/agency/purchases/${id}/proof`
const refundQRURL = (id: number) => `${config.public.apiBase}/agency/purchases/${id}/refund-qr`
const refundProofURL = (id: number) => `${config.public.apiBase}/agency/purchases/${id}/refund-proof`
const review = async (decision: 'confirm' | 'request_correction' | 'refund') => {
  if (saving.value) return
  if (decision !== 'confirm' && reason.value.length < 5) { error.value = 'Escribe un motivo de al menos cinco caracteres.'; return }
  saving.value = true; error.value = ''
  try { await auth.request(`/agency/purchases/${props.purchase.id}/review`, { method: 'PATCH', body: { version: props.purchase.version, decision, reason: reason.value } }); emit('updated') }
  catch (cause) { error.value = apiError(cause, 'No se pudo registrar la revisión').message }
  finally { saving.value = false }
}
const readRefundProof = async (event: Event) => {
  const input = event.target as HTMLInputElement; const file = input.files?.[0]; refundProof.value = ''; if (!file) return
  if (!['image/png', 'image/jpeg'].includes(file.type) || file.size > 5 * 1024 * 1024) { error.value = 'Selecciona un PNG o JPG de hasta 5 MB.'; input.value = ''; return }
  try { refundProof.value = await new Promise<string>((resolve, reject) => { const reader = new FileReader(); reader.onload = () => resolve(String(reader.result)); reader.onerror = reject; reader.readAsDataURL(file) }) }
  catch { error.value = 'No se pudo leer el comprobante. Selecciónalo de nuevo.' }
}
const completeRefund = async () => {
  if (!refundProof.value || saving.value) return
  saving.value = true; error.value = ''
  try { await auth.request(`/agency/purchases/${props.purchase.id}/refund`, { method: 'PATCH', body: { version: props.purchase.version, refund_proof: refundProof.value, refund_reference: refundReference.value } }); emit('updated') }
  catch (cause) { error.value = apiError(cause, 'No se pudo completar el reembolso').message }
  finally { saving.value = false }
}
</script>
<style scoped>
.page-heading { display:flex; align-items:end; justify-content:space-between; gap:1rem; }.page-heading h1 { margin:.25rem 0; font-size:clamp(2rem,4vw,3.3rem); }.page-heading p { margin:0; color:var(--color-text-muted); }.review-list { display:grid; gap:1rem; }.card-title { display:flex; align-items:start; justify-content:space-between; gap:1rem; }.card-title small { color:var(--color-primary); }.card-title h2 { margin:.25rem 0; }.review-content { display:grid; grid-template-columns:minmax(16rem,1fr) minmax(14rem,.65fr); gap:1.25rem; align-items:start; }.review-content dl { display:grid; grid-template-columns:1fr 1.2fr; gap:.5rem 1rem; margin:0; }.review-content dt { color:var(--color-text-muted); }.review-content dd { margin:0; text-align:right; }.proof { width:100%; max-height:22rem; object-fit:contain; border:1px solid var(--color-border); border-radius:var(--radius-md); background:#fff; }.image-label { display:block; margin-bottom:.4rem; color:var(--color-text-muted); }.review-actions,.refund-box { grid-column:1/-1; display:grid; gap:.85rem; padding:1rem; border:1px solid var(--color-border); border-radius:var(--radius-sm); }.review-actions { border-radius:0; border-width:1px 0 0; }.refund-box h3,.refund-box p { margin:0; }.refund-qr { max-width:24rem; }.bank-details { max-width:34rem; }
@media(max-width:800px) { .review-content { grid-template-columns:1fr; }.page-heading { align-items:start; flex-direction:column; }.card-title { flex-direction:column; }.review-content dl { grid-template-columns:1fr; gap:.2rem; }.review-content dd { margin-bottom:.6rem; text-align:left; } }
</style>
