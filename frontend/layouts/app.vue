<template><ClientOnly><AppShell :mode="auth.isAdmin ? 'admin' : 'user'"><slot /></AppShell><template #fallback><main class="app-main"><UiLoadingState /></main></template></ClientOnly></template>
<script setup lang="ts">const auth = useAuthStore()</script>
