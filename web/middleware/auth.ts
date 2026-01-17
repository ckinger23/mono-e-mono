export default defineNuxtRouteMiddleware((to) => {
  const authStore = useAuthStore()

  // Public routes that don't require authentication
  const publicRoutes = ['/', '/login', '/register', '/auth/callback']

  if (publicRoutes.includes(to.path)) {
    // Redirect to dashboard if already authenticated
    if (authStore.isAuthenticated && (to.path === '/login' || to.path === '/register')) {
      return navigateTo('/dashboard')
    }
    return
  }

  // Require authentication for all other routes
  if (!authStore.isAuthenticated) {
    return navigateTo('/login')
  }
})
