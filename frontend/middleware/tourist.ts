export default defineNuxtRouteMiddleware(async (to) => {
  const auth = useAuthStore()
  if (!await auth.initialize()) return navigateTo(`/login?redirect=${encodeURIComponent(to.fullPath)}`)
  if (auth.user?.role !== 'turista') return abortNavigation(createError({ statusCode: 403, statusMessage: 'Los favoritos pertenecen a cuentas de turista' }))
})
