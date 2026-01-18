export default defineNuxtRouteMiddleware((to) => {
  const user = useSupabaseUser()

  // The @nuxtjs/supabase module handles most redirects automatically
  // This middleware adds additional protection for authenticated routes

  // If there's no user and trying to access a protected route
  if (!user.value) {
    return navigateTo('/login')
  }
})
