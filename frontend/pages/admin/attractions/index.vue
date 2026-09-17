<template><AttractionList admin /></template>
<script setup lang="ts">
import AttractionList from '~/components/attractions/AttractionList.vue'
definePageMeta({ layout: 'admin', middleware: 'admin' })
useHead({ title: 'Atracciones' })
</script>
