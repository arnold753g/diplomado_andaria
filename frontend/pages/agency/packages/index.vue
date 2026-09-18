<template><PackageList /></template>
<script setup lang="ts">
import PackageList from '~/components/packages/PackageList.vue'
definePageMeta({ layout: 'app', middleware: 'agency' })
useHead({ title: 'Paquetes turísticos' })
</script>
