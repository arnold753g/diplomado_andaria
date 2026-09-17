<template><AttractionList managed /></template>
<script setup lang="ts">
import AttractionList from '~/components/attractions/AttractionList.vue'
definePageMeta({ layout: 'app', middleware: 'attraction' })
useHead({ title: 'Mis atracciones' })
</script>
