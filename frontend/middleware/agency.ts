export default defineNuxtRouteMiddleware(async (to) => {
  const auth = useAuthStore()
  if (!await auth.initialize()) return navigateTo(`/login?redirect=${encodeURIComponent(to.fullPath)}`)
  if (auth.user?.role !== 'encargado_agencia') return navigateTo('/403')
})
