<template>
  <nav class="bg-white shadow-sm">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex justify-between h-16">
        <div class="flex items-center">
          <NuxtLink to="/dashboard" class="flex items-center">
            <span class="text-xl font-bold text-primary-600">Mono-e-Mono</span>
          </NuxtLink>

          <div v-if="authStore.isAuthenticated" class="hidden sm:ml-8 sm:flex sm:space-x-4">
            <NuxtLink
              to="/dashboard"
              class="text-gray-600 hover:text-gray-900 px-3 py-2 text-sm font-medium"
              active-class="text-primary-600"
            >
              Dashboard
            </NuxtLink>
            <NuxtLink
              to="/leagues/create"
              class="text-gray-600 hover:text-gray-900 px-3 py-2 text-sm font-medium"
              active-class="text-primary-600"
            >
              Create League
            </NuxtLink>
            <NuxtLink
              to="/leagues/join"
              class="text-gray-600 hover:text-gray-900 px-3 py-2 text-sm font-medium"
              active-class="text-primary-600"
            >
              Join League
            </NuxtLink>
          </div>
        </div>

        <div class="flex items-center">
          <template v-if="authStore.isAuthenticated">
            <div class="flex items-center space-x-4">
              <span class="text-sm text-gray-600">{{ authStore.user?.display_name }}</span>
              <button
                @click="handleLogout"
                class="text-gray-600 hover:text-gray-900 text-sm font-medium"
              >
                Logout
              </button>
            </div>
          </template>
          <template v-else>
            <NuxtLink to="/login" class="btn-secondary mr-2">
              Login
            </NuxtLink>
            <NuxtLink to="/register" class="btn-primary">
              Sign Up
            </NuxtLink>
          </template>
        </div>
      </div>
    </div>
  </nav>
</template>

<script setup lang="ts">
const authStore = useAuthStore()

const handleLogout = () => {
  authStore.logout()
}
</script>
