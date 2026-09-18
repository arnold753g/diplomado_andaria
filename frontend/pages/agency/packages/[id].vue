<template><section class="stack"><NuxtLink to="/agency/packages">← Volver a paquetes</NuxtLink><div class="page-heading"><h1>Gestionar paquete turístico</h1><p>Actualiza sus tarifas, itinerario, atracciones y publicación.</p></div><PackageEditor :key="id" :package-id="id" /></section></template>
<script setup lang="ts">
import PackageEditor from '~/components/packages/PackageEditor.vue'
const route = useRoute(); const id = computed(() => Number(route.params.id))
definePageMeta({ layout: 'app', middleware: 'agency', validate: route => /^[1-9]\d*$/.test(String(route.params.id)) })
useHead({ title: 'Gestionar paquete' })
</script>
