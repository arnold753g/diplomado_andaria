<template>
  <section>
    <div class="page-heading"><span class="auth-eyebrow">Gestión de agencia</span><h1>Mi agencia</h1><p>Actualiza la información y los medios de pago de tu agencia.</p></div>
    <AgencyEditor />
  </section>
</template>
<script setup lang="ts">
import AgencyEditor from '~/components/agencies/AgencyEditor.vue'
definePageMeta({ layout: 'app', middleware: 'agency' })
useHead({ title: 'Mi agencia' })
</script>
