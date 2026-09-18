<template><section class="stack"><NuxtLink to="/agency/packages">← Volver a paquetes</NuxtLink><div class="page-heading"><h1>Crear paquete turístico</h1><p>Guarda un borrador y completa su itinerario antes de publicarlo.</p></div><PackageEditor create @saved="created" /></section></template>
<script setup lang="ts">
import PackageEditor from '~/components/packages/PackageEditor.vue'
import type { TourPackage } from '~/types/package'
definePageMeta({ layout: 'app', middleware: 'agency' })
useHead({ title: 'Crear paquete' })
const created = (item: TourPackage) => navigateTo(`/agency/packages/${item.id}`, { replace: true })
</script>
