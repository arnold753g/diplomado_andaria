<template>
  <NuxtLink v-if="to" class="dashboard-metric" :class="`tone-${tone}`" :to="to">
    <span class="metric-icon" aria-hidden="true"><i :class="icon" /></span>
    <span class="metric-copy"><small>{{ label }}</small><strong>{{ value }}</strong><span v-if="hint">{{ hint }}</span></span>
    <i class="pi pi-arrow-right metric-arrow" aria-hidden="true" />
  </NuxtLink>
  <div v-else class="dashboard-metric" :class="`tone-${tone}`">
    <span class="metric-icon" aria-hidden="true"><i :class="icon" /></span>
    <span class="metric-copy"><small>{{ label }}</small><strong>{{ value }}</strong><span v-if="hint">{{ hint }}</span></span>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{ label: string; value: number | string; icon: string; hint?: string; to?: string; tone?: 'primary' | 'info' | 'warning' | 'danger' | 'success' }>(), { hint: '', to: '', tone: 'primary' })
</script>

<style scoped>
.dashboard-metric { position:relative; min-height:9rem; display:flex; align-items:flex-start; gap:1rem; padding:1.25rem; overflow:hidden; border:1px solid var(--color-border); border-radius:var(--radius-lg); background:linear-gradient(145deg,var(--color-surface),var(--color-surface-elevated)); color:var(--color-text); text-decoration:none; box-shadow:var(--shadow-sm); transition:transform var(--motion-fast),border-color var(--motion-fast); }
a.dashboard-metric:hover { transform:translateY(-2px); border-color:var(--metric-color); color:var(--color-text); }
.metric-icon { flex:0 0 auto; width:2.8rem; height:2.8rem; display:grid; place-items:center; border-radius:var(--radius-sm); background:color-mix(in srgb,var(--metric-color) 15%,transparent); color:var(--metric-color); font-size:1.2rem; }
.metric-copy { min-width:0; display:grid; gap:.2rem; }.metric-copy small,.metric-copy span { color:var(--color-text-muted); }.metric-copy strong { color:var(--color-text); font-size:clamp(1.65rem,3vw,2.25rem); line-height:1; font-variant-numeric:tabular-nums; }.metric-copy span { font-size:.84rem; line-height:1.35; }
.metric-arrow { margin-left:auto; color:var(--metric-color); }.tone-primary { --metric-color:var(--color-primary); }.tone-info { --metric-color:var(--color-info); }.tone-warning { --metric-color:var(--color-warning); }.tone-danger { --metric-color:var(--color-danger); }.tone-success { --metric-color:var(--color-success); }
@media(max-width:640px) { .dashboard-metric { min-height:7.5rem; } }
</style>
