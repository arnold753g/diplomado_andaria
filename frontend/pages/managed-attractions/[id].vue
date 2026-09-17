<template><section class="stack"><NuxtLink to="/managed-attractions">← Mis atracciones</NuxtLink><div class="page-heading"><h1>Editar atracción</h1><p>Prepara la información del lugar y publícala cuando esté lista.</p></div><AttractionEditor :key="id" :attraction-id="id" /></section></template>
<script setup lang="ts">
import AttractionEditor from '~/components/attractions/AttractionEditor.vue'
definePageMeta({ layout: 'app', middleware: 'attraction', validate: route => /^[1-9]\d*$/.test(String(route.params.id)) })
useHead({ title: 'Editar atracción' })
const route = useRoute(); const id = computed(() => Number(route.params.id))
</script>
