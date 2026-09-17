<template><section class="stack"><NuxtLink to="/admin/attractions">← Volver a atracciones</NuxtLink><div class="page-heading"><h1>Gestionar atracción</h1><p>Actualiza la ficha, el estado o el encargado del lugar.</p></div><AttractionEditor :key="id" admin :attraction-id="id" /></section></template>
<script setup lang="ts">
import AttractionEditor from '~/components/attractions/AttractionEditor.vue'
definePageMeta({ layout: 'admin', middleware: 'admin', validate: route => /^[1-9]\d*$/.test(String(route.params.id)) })
useHead({ title: 'Gestionar atracción' })
const route = useRoute(); const id = computed(() => Number(route.params.id))
</script>
