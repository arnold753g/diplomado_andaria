<template>
  <ClientOnly>
    <AppShell v-if="ready" :mode="auth.isAuthenticated ? auth.isAdmin ? 'admin' : 'user' : 'guest'"><slot /></AppShell>
    <main v-else class="app-main"><UiLoadingState /></main>
    <template #fallback><main class="app-main"><UiLoadingState /></main></template>
  </ClientOnly>
</template>
<script setup lang="ts">
const auth = useAuthStore()
const ready = ref(false)
onMounted(async () => { await auth.initialize(); ready.value = true })
</script>
