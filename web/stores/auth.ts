import { defineStore } from 'pinia'

interface Profile {
  id: string
  display_name: string
  avatar_url: string | null
  created_at: string
  updated_at: string
}

export const useAuthStore = defineStore('auth', () => {
  const supabase = useSupabaseClient()
  const user = useSupabaseUser()

  const profile = ref<Profile | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => !!user.value)

  const currentUser = computed(() => {
    if (!user.value) return null
    return {
      id: user.value.id,
      email: user.value.email || '',
      display_name: profile.value?.display_name || user.value.email || '',
      avatar_url: profile.value?.avatar_url,
    }
  })

  async function fetchProfile() {
    if (!user.value) return null

    loading.value = true
    error.value = null

    try {
      const data = await $fetch('/api/users/me')
      profile.value = data as Profile
      return profile.value
    } catch (err: any) {
      error.value = err.data?.message || 'Failed to fetch profile'
      return null
    } finally {
      loading.value = false
    }
  }

  async function updateProfile(updates: Partial<Pick<Profile, 'display_name' | 'avatar_url'>>) {
    if (!user.value) return null

    loading.value = true
    error.value = null

    try {
      const data = await $fetch('/api/users/me', {
        method: 'PUT',
        body: updates,
      })
      profile.value = data as Profile
      return profile.value
    } catch (err: any) {
      error.value = err.data?.message || 'Failed to update profile'
      return null
    } finally {
      loading.value = false
    }
  }

  async function register(email: string, password: string, displayName: string) {
    loading.value = true
    error.value = null

    try {
      const { error: signUpError } = await supabase.auth.signUp({
        email,
        password,
        options: {
          data: {
            display_name: displayName,
          },
        },
      })

      if (signUpError) throw signUpError

      return true
    } catch (err: any) {
      error.value = err.message || 'Failed to sign up'
      return false
    } finally {
      loading.value = false
    }
  }

  async function login(email: string, password: string) {
    loading.value = true
    error.value = null

    try {
      const { error: signInError } = await supabase.auth.signInWithPassword({
        email,
        password,
      })

      if (signInError) throw signInError

      return true
    } catch (err: any) {
      error.value = err.message || 'Failed to sign in'
      return false
    } finally {
      loading.value = false
    }
  }

  async function signInWithGoogle() {
    loading.value = true
    error.value = null

    try {
      const { error: oauthError } = await supabase.auth.signInWithOAuth({
        provider: 'google',
        options: {
          redirectTo: `${window.location.origin}/auth/callback`,
        },
      })

      if (oauthError) throw oauthError

      return true
    } catch (err: any) {
      error.value = err.message || 'Failed to sign in with Google'
      return false
    } finally {
      loading.value = false
    }
  }

  async function signInWithGithub() {
    loading.value = true
    error.value = null

    try {
      const { error: oauthError } = await supabase.auth.signInWithOAuth({
        provider: 'github',
        options: {
          redirectTo: `${window.location.origin}/auth/callback`,
        },
      })

      if (oauthError) throw oauthError

      return true
    } catch (err: any) {
      error.value = err.message || 'Failed to sign in with GitHub'
      return false
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    loading.value = true

    try {
      await supabase.auth.signOut()
      profile.value = null
      navigateTo('/login')
    } catch (err: any) {
      error.value = err.message || 'Failed to sign out'
    } finally {
      loading.value = false
    }
  }

  function clearError() {
    error.value = null
  }

  // Watch for auth changes
  watch(user, async (newUser) => {
    if (newUser) {
      await fetchProfile()
    } else {
      profile.value = null
    }
  }, { immediate: true })

  return {
    // State
    user,
    profile,
    loading,
    error,

    // Computed
    isAuthenticated,
    currentUser,

    // Actions
    fetchProfile,
    updateProfile,
    register,
    login,
    signInWithGoogle,
    signInWithGithub,
    logout,
    clearError,
  }
})
