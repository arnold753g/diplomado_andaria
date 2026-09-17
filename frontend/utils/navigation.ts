import type { Role } from '../types/auth.ts'

export interface NavigationItem { label: string; to: string; icon: string }
export type NavigationMode = 'admin' | 'user' | 'guest'

export const navigationItems = (mode: NavigationMode, role?: Role): NavigationItem[] => {
  if (mode === 'guest') {
    return [
      { label: 'Inicio', to: '/', icon: 'pi pi-home' },
      { label: 'Atracciones', to: '/attractions', icon: 'pi pi-map-marker' },
      { label: 'Paquetes turísticos', to: '/packages', icon: 'pi pi-briefcase' }
    ]
  }
  if (mode === 'admin') {
    return [
      { label: 'Dashboard', to: '/admin', icon: 'pi pi-home' },
      { label: 'Usuarios', to: '/admin/users', icon: 'pi pi-users' },
      { label: 'Agencias', to: '/admin/agencies', icon: 'pi pi-building' },
      { label: 'Atracciones', to: '/admin/attractions', icon: 'pi pi-map-marker' },
      { label: 'Catálogo de atracciones', to: '/attractions', icon: 'pi pi-compass' },
      { label: 'Catálogo de paquetes', to: '/packages', icon: 'pi pi-briefcase' },
      { label: 'Mi perfil', to: '/app/profile', icon: 'pi pi-user' }
    ]
  }
  return [
    { label: 'Inicio', to: '/app', icon: 'pi pi-home' },
    { label: 'Explorar atracciones', to: '/attractions', icon: 'pi pi-compass' },
    { label: 'Explorar paquetes', to: '/packages', icon: 'pi pi-compass' },
    ...(role === 'turista' ? [
      { label: 'Mis compras', to: '/app/purchases', icon: 'pi pi-shopping-bag' },
      { label: 'Mis favoritos', to: '/app/favorites', icon: 'pi pi-heart' }
    ] : []),
    ...(role === 'encargado_atraccion' ? [{ label: 'Mis atracciones', to: '/managed-attractions', icon: 'pi pi-map-marker' }] : []),
    ...(role === 'encargado_agencia' ? [
      { label: 'Mi agencia', to: '/agency', icon: 'pi pi-building' },
      { label: 'Gestionar paquetes', to: '/agency/packages', icon: 'pi pi-briefcase' },
      { label: 'Revisar compras', to: '/agency/purchases', icon: 'pi pi-receipt' }
    ] : []),
    { label: 'Mi perfil', to: '/app/profile', icon: 'pi pi-user' }
  ]
}
