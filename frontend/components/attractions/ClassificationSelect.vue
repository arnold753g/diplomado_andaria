<template>
  <div class="stack">
    <div class="form-field"><label for="classification-category">Categoría</label><Select input-id="classification-category" v-model="browsing" :options="categories" option-label="name" option-value="id" placeholder="Selecciona una categoría" :disabled="disabled" /></div>
    <p class="muted">Selecciona de una a cuatro subcategorías. La primera será la principal; puedes combinar categorías.</p>
    <div v-if="available.length" class="classification-options" role="group" aria-label="Subcategorías disponibles"><Button v-for="s in available" :key="s.id" :label="s.name" outlined size="small" :disabled="disabled || modelValue.length >= 4 || modelValue.includes(s.id)" @click="emit('update:modelValue', [...modelValue, s.id])" /></div>
    <p v-if="!modelValue.length" class="muted">Todavía no seleccionaste subcategorías.</p>
    <ol v-else class="selected-classifications"><li v-for="(id, i) in modelValue" :key="id"><div><strong>{{ lookup(id)?.name }}</strong><small>{{ categoryName(id) }} · {{ i === 0 ? 'Principal' : 'Complementaria' }}</small></div><div class="cluster"><Button v-if="i > 0" label="Hacer principal" size="small" text :disabled="disabled" @click="makePrimary(id)" /><Button icon="pi pi-times" :aria-label="`Quitar ${lookup(id)?.name}`" severity="danger" text :disabled="disabled" @click="emit('update:modelValue', modelValue.filter(value => value !== id))" /></div></li></ol>
    <small>{{ modelValue.length }} de 4 subcategorías seleccionadas.</small>
  </div>
</template>
<script setup lang="ts">
import Select from 'primevue/select'
import Button from 'primevue/button'
import type { AttractionCategory } from '~/types/attraction'
const props = defineProps<{ modelValue: number[]; categories: AttractionCategory[]; disabled?: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [value: number[]] }>()
const browsing = ref<number | null>(null)
const available = computed(() => props.categories.find(c => c.id === browsing.value)?.subcategories || [])
const lookup = (id: number) => props.categories.flatMap(c => c.subcategories).find(s => s.id === id)
const categoryName = (id: number) => props.categories.find(c => c.subcategories.some(s => s.id === id))?.name
const makePrimary = (id: number) => emit('update:modelValue', [id, ...props.modelValue.filter(value => value !== id)])
watch(() => [props.categories, props.modelValue] as const, () => {
  if (browsing.value === null && props.modelValue.length) browsing.value = lookup(props.modelValue[0]!)?.category_id || null
}, { immediate: true })
</script>
<style scoped>
.classification-options { display:flex; gap:.5rem; flex-wrap:wrap; }
.selected-classifications { list-style:none; padding:0; margin:0; display:grid; gap:.6rem; }
.selected-classifications li { display:flex; justify-content:space-between; flex-wrap:wrap; align-items:center; gap:.5rem; padding:.75rem; border:1px solid var(--color-border); border-radius:var(--radius-sm); }
.selected-classifications small { display:block; color:var(--color-text-muted); margin-top:.25rem; }
</style>
