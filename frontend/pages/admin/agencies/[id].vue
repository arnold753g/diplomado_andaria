<template>
  <section>
    <NuxtLink to="/admin/agencies">← Volver a agencias</NuxtLink>
    <div class="page-heading"><h1>Editar agencia</h1><p>Gestiona la información, el encargado y el acceso de esta agencia.</p></div>
    <AgencyEditor :key="id" admin :agency-id="id" />
  </section>
</template>
<script setup lang="ts">
import AgencyEditor from '~/components/agencies/AgencyEditor.vue'
definePageMeta({ layout: 'admin', middleware: 'admin' })
useHead({ title: 'Editar agencia' })
const route = useRoute()
const id = computed(() => Number(route.params.id))
</script>
