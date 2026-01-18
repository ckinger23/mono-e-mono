<template>
  <div class="min-h-screen flex items-center justify-center">
    <div class="text-center">
      <CommonLoadingSpinner size="lg" text="Completing sign in..." />
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: false,
})

const user = useSupabaseUser()

// The @nuxtjs/supabase module handles the OAuth callback automatically
// We just need to wait for the user to be set and then redirect
watch(user, (newUser) => {
  if (newUser) {
    navigateTo('/dashboard')
  }
}, { immediate: true })

// Fallback: if no user after 5 seconds, redirect to login
onMounted(() => {
  setTimeout(() => {
    if (!user.value) {
      navigateTo('/login')
    }
  }, 5000)
})
</script>
