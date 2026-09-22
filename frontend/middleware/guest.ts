export default defineNuxtRouteMiddleware(async () => {
  const auth = useAuthStore()
  if (await auth.initialize()) return navigateTo(auth.isAdmin ? '/admin' : '/app')
})
