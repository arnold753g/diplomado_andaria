<template>
  <div class="receipt-actions">
    <div class="receipt-links">
      <a :href="url" target="_blank" rel="noopener noreferrer"><i class="pi pi-file-pdf" aria-hidden="true" /> Ver comprobante PDF</a>
      <a :href="`${url}?download=1`"><i class="pi pi-download" aria-hidden="true" /> Descargar PDF</a>
    </div>
    <small>Incluye las tarifas de la compra y el estado al generar el documento. La confirmación del pago no autoriza por sí sola la salida.</small>
  </div>
</template>
<script setup lang="ts">
const props = withDefaults(defineProps<{ purchaseId: number; scope?: 'me' | 'agency' }>(), { scope: 'me' })
const config = useRuntimeConfig()
const url = computed(() => `${config.public.apiBase}/${props.scope}/purchases/${props.purchaseId}/receipt`)
</script>
<style scoped>
.receipt-actions{grid-column:1/-1;display:grid;gap:.65rem;padding:1rem;border:1px solid var(--color-border);border-radius:var(--radius-sm)}
.receipt-links{display:flex;flex-wrap:wrap;gap:.7rem}.receipt-links a{display:inline-flex;align-items:center;justify-content:center;gap:.6rem;min-height:44px;padding:.65rem .9rem;border:1px solid var(--color-border);border-radius:var(--radius-sm);color:var(--color-text);text-decoration:none}
.receipt-links a:focus-visible{outline:2px solid var(--color-primary);outline-offset:3px}.receipt-actions small{color:var(--color-text-muted);line-height:1.6}
@media(max-width:520px){.receipt-links{flex-direction:column}.receipt-links a{width:100%}}
</style>
