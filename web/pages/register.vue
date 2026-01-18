<template>
  <div>
    <div class="text-center mb-8">
      <h1 class="text-3xl font-bold text-white">Create Account</h1>
      <p class="mt-2 text-primary-200">Join the weekly draft revolution</p>
    </div>

    <div class="card">
      <div class="card-body">
        <CommonAlert
          v-if="success"
          type="success"
          message="Account created! Please check your email to confirm your account."
        />

        <form v-if="!success" @submit.prevent="handleSubmit" class="space-y-6">
          <CommonAlert
            v-if="error"
            type="error"
            :message="error"
            dismissible
            @dismiss="error = ''"
          />

          <div>
            <label for="displayName" class="label">Display Name</label>
            <input
              id="displayName"
              v-model="displayName"
              type="text"
              required
              class="input"
              placeholder="Your name"
            />
          </div>

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
              minlength="8"
              class="input"
              placeholder="At least 8 characters"
            />
          </div>

          <div>
            <label for="confirmPassword" class="label">Confirm Password</label>
            <input
              id="confirmPassword"
              v-model="confirmPassword"
              type="password"
              required
              class="input"
              placeholder="Confirm your password"
            />
          </div>

          <button
            type="submit"
            :disabled="loading"
            class="btn-primary w-full"
          >
            <CommonLoadingSpinner v-if="loading" size="sm" />
            <span v-else>Create Account</span>
          </button>
        </form>

        <div v-if="!success" class="mt-6">
          <div class="relative">
            <div class="absolute inset-0 flex items-center">
              <div class="w-full border-t border-gray-300" />
            </div>
            <div class="relative flex justify-center text-sm">
              <span class="px-2 bg-white text-gray-500">Or continue with</span>
            </div>
          </div>

          <div class="mt-6 grid grid-cols-2 gap-3">
            <button
              type="button"
              @click="handleGoogleSignIn"
              :disabled="loading"
              class="btn-secondary w-full"
            >
              Google
            </button>
            <button
              type="button"
              @click="handleGithubSignIn"
              :disabled="loading"
              class="btn-secondary w-full"
            >
              GitHub
            </button>
          </div>
        </div>

        <p class="mt-6 text-center text-sm text-gray-600">
          Already have an account?
          <NuxtLink to="/login" class="text-primary-600 hover:text-primary-500 font-medium">
            Sign in
          </NuxtLink>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'auth',
})

const authStore = useAuthStore()
const user = useSupabaseUser()

const displayName = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const error = ref('')
const success = ref(false)

// Redirect if already logged in
watch(user, (newUser) => {
  if (newUser) {
    navigateTo('/dashboard')
  }
}, { immediate: true })

const handleSubmit = async () => {
  if (password.value !== confirmPassword.value) {
    error.value = 'Passwords do not match'
    return
  }

  loading.value = true
  error.value = ''

  try {
    const result = await authStore.register(email.value, password.value, displayName.value)
    if (result) {
      success.value = true
    } else {
      error.value = authStore.error || 'Registration failed. Please try again.'
    }
  } catch (err: any) {
    error.value = err.message || 'Registration failed. Please try again.'
  } finally {
    loading.value = false
  }
}

const handleGoogleSignIn = async () => {
  loading.value = true
  error.value = ''

  try {
    await authStore.signInWithGoogle()
  } catch (err: any) {
    error.value = err.message || 'Google sign in failed.'
  } finally {
    loading.value = false
  }
}

const handleGithubSignIn = async () => {
  loading.value = true
  error.value = ''

  try {
    await authStore.signInWithGithub()
  } catch (err: any) {
    error.value = err.message || 'GitHub sign in failed.'
  } finally {
    loading.value = false
  }
}
</script>
