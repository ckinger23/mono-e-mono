<template>
  <div>
    <div class="text-center mb-8">
      <h1 class="text-3xl font-bold text-white">Welcome Back</h1>
      <p class="mt-2 text-primary-200">Sign in to your account</p>
    </div>

    <div class="card">
      <div class="card-body">
        <form @submit.prevent="handleSubmit" class="space-y-6">
          <CommonAlert
            v-if="error"
            type="error"
            :message="error"
            dismissible
            @dismiss="error = ''"
          />

          <div>
            <label for="email" class="label">Email</label>
            <input
              id="email"
              v-model="email"
              type="email"
              required
              class="input"
              placeholder="you@example.com"
            />
          </div>

          <div>
            <label for="password" class="label">Password</label>
            <input
              id="password"
              v-model="password"
              type="password"
              required
              class="input"
              placeholder="Enter your password"
            />
          </div>

          <button
            type="submit"
            :disabled="loading"
            class="btn-primary w-full"
          >
            <CommonLoadingSpinner v-if="loading" size="sm" />
            <span v-else>Sign In</span>
          </button>
        </form>

        <div class="mt-6">
          <div class="relative">
            <div class="absolute inset-0 flex items-center">
              <div class="w-full border-t border-gray-300" />
            </div>
            <div class="relative flex justify-center text-sm">
              <span class="px-2 bg-white text-gray-500">Or continue with</span>
            </div>
          </div>

          <div class="mt-6 grid grid-cols-2 gap-3">
            <a
              :href="`${config.public.apiBase}/api/auth/google`"
              class="btn-secondary w-full"
            >
              Google
            </a>
            <a
              :href="`${config.public.apiBase}/api/auth/github`"
              class="btn-secondary w-full"
            >
              GitHub
            </a>
          </div>
        </div>

        <p class="mt-6 text-center text-sm text-gray-600">
          Don't have an account?
          <NuxtLink to="/register" class="text-primary-600 hover:text-primary-500 font-medium">
            Sign up
          </NuxtLink>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'auth',
  middleware: 'auth',
})

const config = useRuntimeConfig()
const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

const handleSubmit = async () => {
  loading.value = true
  error.value = ''

  try {
    await authStore.login(email.value, password.value)
    navigateTo('/dashboard')
  } catch (err: any) {
    error.value = err.data?.message || 'Login failed. Please try again.'
  } finally {
    loading.value = false
  }
}
</script>
