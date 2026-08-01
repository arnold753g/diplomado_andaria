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
        <NuxtLink v-for="item in items" :key="item.to" :to="item.to" @click="menuOpen = false">
          <i :class="item.icon" aria-hidden="true" />
          <span>{{ item.label }}</span>
        </NuxtLink>
      </nav>
      <div class="sidebar-footer">
        <Button label="Cerrar sesión" icon="pi pi-sign-out" severity="secondary" outlined fluid :loading="loggingOut" @click="logout" />
      </div>
    </aside>

    <div class="app-content">
      <header class="app-header">
        <button class="icon-button mobile-only" type="button" aria-label="Abrir navegación" @click="menuOpen = true">
          <i class="pi pi-bars" aria-hidden="true" />
        </button>
        <div class="header-spacer" />
        <div class="header-account">
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
import { roleLabel } from '~/utils/roles'

interface NavigationItem { label: string, to: string, icon: string }

const props = defineProps<{ mode: 'admin' | 'user' }>()
const auth = useAuthStore()
const branding = useBranding()
const menuOpen = ref(false)
const loggingOut = ref(false)
const homePath = computed(() => props.mode === 'admin' ? '/admin' : '/app')
const userInitials = computed(() => {
  const names = [auth.user?.first_name, auth.user?.last_name].filter(Boolean)
  return names.map(value => String(value).charAt(0).toUpperCase()).join('').slice(0, 2) || 'U'
})
const items = computed<NavigationItem[]>(() => props.mode === 'admin'
  ? [
      { label: 'Dashboard', to: '/admin', icon: 'pi pi-home' },
      { label: 'Usuarios', to: '/admin/users', icon: 'pi pi-users' },
      { label: 'Mi perfil', to: '/app/profile', icon: 'pi pi-user' }
    ]
  : [
      { label: 'Inicio', to: '/app', icon: 'pi pi-home' },
      { label: 'Mi perfil', to: '/app/profile', icon: 'pi pi-user' }
    ])

const logout = async () => {
  loggingOut.value = true
  await auth.logout()
  await navigateTo('/login')
}
</script>
