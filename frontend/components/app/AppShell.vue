<template>
  <div class="app-shell">
    <a class="skip-link" href="#main-content">Saltar al contenido</a>
    <button v-if="menuOpen" class="sidebar-overlay" type="button" aria-label="Cerrar navegación" @click="menuOpen = false" />
    <aside class="app-sidebar" :class="{ 'is-open': menuOpen }" aria-label="Navegación principal">
      <NuxtLink class="sidebar-brand" :to="homePath" @click="menuOpen = false">
        <img :src="branding.logo" :alt="branding.name">
        <span class="brand-name">{{ branding.name }}</span>
      </NuxtLink>
      <nav class="sidebar-nav">
        <NuxtLink v-for="item in items" :key="item.to" :to="item.to" :class="{ 'nav-active': navigationActive(item.to) }" @click="menuOpen = false">
          <i :class="item.icon" aria-hidden="true" />
          <span>{{ item.label }}</span>
        </NuxtLink>
      </nav>
      <div class="sidebar-footer">
        <template v-if="mode === 'guest'">
          <NuxtLink class="sidebar-login" to="/login" @click="menuOpen = false"><i class="pi pi-sign-in" aria-hidden="true" /> Ingresar</NuxtLink>
          <NuxtLink class="sidebar-register" to="/register" @click="menuOpen = false">Crear cuenta</NuxtLink>
        </template>
        <Button v-else label="Cerrar sesión" icon="pi pi-sign-out" severity="secondary" outlined fluid :loading="loggingOut" @click="logout" />
      </div>
    </aside>

    <div class="app-content">
      <header class="app-header">
        <button class="icon-button mobile-only" type="button" aria-label="Abrir navegación" @click="menuOpen = true">
          <i class="pi pi-bars" aria-hidden="true" />
        </button>
        <div class="header-spacer" />
        <div v-if="mode !== 'guest'" class="header-account">
          <div class="header-user">
            <strong>{{ auth.fullName }}</strong>
            <span>{{ roleLabel(auth.user?.role) }}</span>
          </div>
          <div class="header-avatar" aria-hidden="true">{{ userInitials }}</div>
        </div>
      </header>
      <main id="main-content" class="app-main" tabindex="-1">
        <div class="page-container"><slot /></div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import Button from 'primevue/button'
import type { NavigationMode } from '~/utils/navigation'
import { navigationItems } from '~/utils/navigation'
import { roleLabel } from '~/utils/roles'

const props = defineProps<{ mode: NavigationMode }>()
const auth = useAuthStore()
const branding = useBranding()
const menuOpen = ref(false)
const loggingOut = ref(false)
const homePath = computed(() => props.mode === 'admin' ? '/admin' : props.mode === 'guest' ? '/' : '/app')
const userInitials = computed(() => {
  const names = [auth.user?.first_name, auth.user?.last_name].filter(Boolean)
  return names.map(value => String(value).charAt(0).toUpperCase()).join('').slice(0, 2) || 'U'
})
const items = computed(() => navigationItems(props.mode, auth.user?.role))
const route = useRoute()
const navigationActive = (path: string) => {
  const matches = items.value.filter(item => route.path === item.to || route.path.startsWith(`${item.to}/`))
  return matches.sort((a, b) => b.to.length - a.to.length)[0]?.to === path
}

const logout = async () => {
  loggingOut.value = true
  await auth.logout()
  await navigateTo('/login')
}
</script>

<style scoped>
.sidebar-login,.sidebar-register { display:flex; align-items:center; justify-content:center; gap:.55rem; width:100%; padding:.72rem 1rem; border-radius:var(--radius-sm); font-weight:700; text-decoration:none; }
.sidebar-login { border:1px solid var(--color-primary); }
.sidebar-register { margin-top:.55rem; border:1px solid var(--color-border); color:var(--color-text); }
</style>
