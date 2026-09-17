export default defineNuxtRouteMiddleware(async (to) => {
  const auth = useAuthStore()
  if (!await auth.initialize()) return navigateTo(`/login?redirect=${encodeURIComponent(to.fullPath)}`)
  if (auth.user?.role !== 'encargado_atraccion') return abortNavigation(createError({ statusCode: 403, statusMessage: 'Acceso exclusivo para encargados de atracciones' }))
})
