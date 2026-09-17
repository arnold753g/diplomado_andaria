<template><div class="time-select"><Select :input-id="`${id}-hour`" :model-value="hour" :options="hours" placeholder="Hora" :aria-label="`${label}: hora`" :disabled="disabled" @update:model-value="setHour" /><span aria-hidden="true">:</span><Select :input-id="`${id}-minute`" :model-value="minute" :options="minutes" placeholder="Minutos" :aria-label="`${label}: minutos`" :disabled="disabled" @update:model-value="setMinute" /></div></template>
<script setup lang="ts">
import Select from 'primevue/select'
const props = defineProps<{ modelValue: string; id: string; label: string; disabled?: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()
const hours = Array.from({ length: 24 }, (_, i) => String(i).padStart(2, '0'))
const minutes = Array.from({ length: 60 }, (_, i) => String(i).padStart(2, '0'))
const hour = computed(() => props.modelValue ? props.modelValue.slice(0, 2) : null)
const minute = computed(() => props.modelValue ? props.modelValue.slice(3, 5) : null)
const setHour = (h: string) => emit('update:modelValue', `${h}:${minute.value || '00'}`)
const setMinute = (m: string) => emit('update:modelValue', `${hour.value || '00'}:${m}`)
</script>
<style scoped>.time-select { display:grid; grid-template-columns:minmax(0,1fr) auto minmax(0,1fr); align-items:center; gap:.5rem; }</style>
