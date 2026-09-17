<template><AttractionList favorites /></template>
<script setup lang="ts">
import AttractionList from '~/components/attractions/AttractionList.vue'
definePageMeta({ layout: 'app', middleware: 'tourist' })
useHead({ title: 'Mis favoritos' })
</script>
