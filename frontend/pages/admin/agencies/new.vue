<template>
  <section>
    <NuxtLink to="/admin/agencies">← Volver a agencias</NuxtLink>
    <div class="page-heading"><h1>Crear agencia</h1><p>Registra sus datos y asigna un encargado disponible.</p></div>
    <AgencyEditor admin create @saved="created" />
  </section>
</template>
<script setup lang="ts">
import AgencyEditor from '~/components/agencies/AgencyEditor.vue'
import type { Agency } from '~/types/agency'
definePageMeta({ layout: 'admin', middleware: 'admin' })
useHead({ title: 'Crear agencia' })
const created = (agency: Agency) => navigateTo(`/admin/agencies/${agency.id}`)
</script>
