<template>
  <div class="destination-form stack">
    <div class="form-field"><label :for="`${idPrefix}-method`">Recibir la devolución mediante *</label><Select :input-id="`${idPrefix}-method`" v-model="form.method" :options="methods" option-label="label" option-value="value" /></div>
    <div v-if="form.method === 'qr'" class="form-field"><label :for="`${idPrefix}-qr`">QR para la devolución *</label><input :id="`${idPrefix}-qr`" type="file" accept="image/png,image/jpeg" @change="$emit('readQr', $event)"><small>PNG o JPG de hasta 5 MB.</small></div>
    <div v-else class="bank-grid">
      <div class="form-field"><label :for="`${idPrefix}-bank`">Banco *</label><InputText :id="`${idPrefix}-bank`" v-model.trim="form.bankName" maxlength="120" /></div>
      <div class="form-field"><label :for="`${idPrefix}-holder`">Titular *</label><InputText :id="`${idPrefix}-holder`" v-model.trim="form.holder" maxlength="160" /></div>
      <div class="form-field"><label :for="`${idPrefix}-account`">Número de cuenta *</label><InputText :id="`${idPrefix}-account`" v-model.trim="form.account" maxlength="80" /></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
defineProps<{ idPrefix: string; form: { method: 'qr' | 'bank_transfer'; qr: string; bankName: string; holder: string; account: string } }>()
defineEmits<{ readQr: [event: Event] }>()
const methods = [{ label: 'QR', value: 'qr' }, { label: 'Cuenta bancaria', value: 'bank_transfer' }]
</script>

<style scoped>
.destination-form { padding:.85rem; border-radius:var(--radius-sm); background:var(--color-surface-elevated); }.bank-grid { display:grid; grid-template-columns:repeat(2,minmax(0,1fr)); gap:1rem; }.bank-grid .form-field:last-child { grid-column:1/-1; }
@media(max-width:650px) { .bank-grid { grid-template-columns:1fr; }.bank-grid .form-field:last-child { grid-column:auto; } }
</style>
