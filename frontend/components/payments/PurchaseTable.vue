<template>
  <table class="purchase-table">
    <caption class="sr-only">{{ tab === 'tourists' ? 'Compradores y grupos de la salida' : 'Pagos de la salida' }}</caption>
    <thead><tr><th scope="col">Comprador responsable</th><template v-if="tab === 'tourists'"><th scope="col">Contacto</th><th scope="col">Adultos / menores</th><th scope="col">Turistas / cupos</th></template><template v-else><th scope="col">Importe</th><th scope="col">{{ tab === 'refunds' ? 'Plazo de devolución' : 'Grupo' }}</th></template><th scope="col">Estado</th><th scope="col">Detalle</th></tr></thead>
    <tbody><tr v-for="purchase in purchases" :key="purchase.id"><td data-label="Comprador"><strong>{{ purchase.tourist_name }}</strong><small>{{ purchase.reference }}</small></td>
      <template v-if="tab === 'tourists'"><td data-label="Contacto"><span>{{ purchase.tourist_phone || 'Sin teléfono' }}</span><small>{{ purchase.tourist_email }}</small></td><td data-label="Adultos / menores">{{ purchase.national_adults + purchase.foreign_adults }} / {{ purchase.minors.length }}</td><td data-label="Turistas / cupos">{{ purchaseTravelers(purchase) }} / {{ purchase.capacity_count }}</td></template>
      <template v-else><td data-label="Importe"><strong>{{ packageMoney(purchase.total_cents) }}</strong><small>{{ paymentMethodLabel(purchase.payment_method) }}</small></td><td :data-label="tab === 'refunds' ? 'Plazo' : 'Grupo'"><template v-if="tab === 'refunds'">{{ purchase.refund_due_at ? paymentDateTime(purchase.refund_due_at) : '—' }}<small v-if="purchase.refund_overdue" class="overdue">Plazo vencido</small></template><template v-else>{{ purchaseTravelers(purchase) }} {{ purchaseTravelers(purchase) === 1 ? 'turista' : 'turistas' }}<small>{{ purchase.capacity_count }} {{ purchase.capacity_count === 1 ? 'cupo' : 'cupos' }}</small></template></td></template>
      <td data-label="Estado"><Tag :value="purchaseStatus(purchase.status)" :severity="purchaseSeverity(purchase.status)" /></td><td data-label="Detalle"><Button label="Ver detalle" icon="pi pi-eye" severity="secondary" text @click="$emit('open', purchase)" /></td>
    </tr></tbody>
  </table>
</template>
<script setup lang="ts">
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import type { Purchase } from '~/types/purchase'
import { packageMoney } from '~/utils/packages'
import { purchaseTravelers, paymentDateTime } from '~/utils/payments'
import { paymentMethodLabel, purchaseSeverity, purchaseStatus } from '~/utils/purchases'
defineProps<{ purchases: Purchase[]; tab: 'payments' | 'tourists' | 'refunds' }>()
defineEmits<{ open: [purchase: Purchase] }>()
</script>
<style scoped>
.purchase-table{width:100%;border-collapse:collapse;font-size:.9rem;table-layout:fixed}.purchase-table th{text-align:left;padding:1rem .7rem;font-size:.8rem;font-weight:600;color:var(--color-text-muted);background:var(--color-surface-elevated)}.purchase-table td{padding:1rem .7rem;border-bottom:1px solid var(--color-border);vertical-align:middle;overflow-wrap:anywhere}.purchase-table td strong,.purchase-table td small{display:block}.purchase-table td small{color:var(--color-text-muted);font-size:.76rem;margin-top:.3rem}.purchase-table th:first-child{width:24%}.overdue{color:var(--color-danger)!important}.sr-only{position:absolute;width:1px;height:1px;overflow:hidden;clip:rect(0,0,0,0);white-space:nowrap}
@media(max-width:1150px){.purchase-table{font-size:.8rem}}
@media(max-width:760px){.purchase-table,.purchase-table tbody,.purchase-table tr,.purchase-table td{display:block;width:100%}.purchase-table thead{display:none}.purchase-table tr{padding:.8rem 0;border-bottom:1px solid var(--color-border)}.purchase-table td{border:0;padding:.4rem 0;display:grid;grid-template-columns:7.4rem minmax(0,1fr);gap:.2rem .7rem}.purchase-table td::before{content:attr(data-label);grid-column:1;grid-row:1/4;color:var(--color-text-muted);font-size:.8rem}.purchase-table td>*{grid-column:2}.purchase-table td>.p-button{justify-self:start}}
@media(max-width:380px){.purchase-table td{grid-template-columns:6.5rem minmax(0,1fr)}}
</style>
