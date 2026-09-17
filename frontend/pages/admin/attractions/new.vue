<template><section class="stack"><NuxtLink to="/admin/attractions">← Volver a atracciones</NuxtLink><div class="page-heading"><h1>Crear atracción</h1><p>Registra el lugar y asigna su encargado. Puedes empezar con un borrador.</p></div><AttractionEditor admin create @saved="created" /></section></template>
<script setup lang="ts">
import AttractionEditor from '~/components/attractions/AttractionEditor.vue'
import type { Attraction } from '~/types/attraction'
definePageMeta({ layout: 'admin', middleware: 'admin' })
useHead({ title: 'Crear atracción' })
const created = (a: Attraction) => navigateTo(`/admin/attractions/${a.id}`)
</script>
