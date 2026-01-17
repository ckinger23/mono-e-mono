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

const route = useRoute()
const authStore = useAuthStore()

onMounted(async () => {
  const accessToken = route.query.access_token as string
  const refreshToken = route.query.refresh_token as string

  if (accessToken && refreshToken) {
    authStore.tokens = {
      access_token: accessToken,
      refresh_token: refreshToken,
      expires_in: 86400, // 24 hours
    }
    authStore.saveTokens()
    await authStore.fetchUser()
    navigateTo('/dashboard')
  } else {
    navigateTo('/login')
  }
})
</script>
