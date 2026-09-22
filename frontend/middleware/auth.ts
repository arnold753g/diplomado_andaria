export default defineNuxtRouteMiddleware(async (to) => {
  const auth = useAuthStore()
  if (!await auth.initialize()) return navigateTo(`/login?redirect=${encodeURIComponent(to.fullPath)}`)
})
