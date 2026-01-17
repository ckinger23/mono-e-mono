import { defineStore } from 'pinia'
import type { User, TokenPair, AuthResponse } from '~/types'

interface AuthState {
  user: User | null
  tokens: TokenPair | null
  loading: boolean
  error: string | null
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    user: null,
    tokens: null,
    loading: false,
    error: null,
  }),

  getters: {
    isAuthenticated: (state) => !!state.tokens?.access_token,
    accessToken: (state) => state.tokens?.access_token,
  },

  actions: {
    async register(email: string, password: string, displayName: string) {
      this.loading = true
      this.error = null

      try {
        const config = useRuntimeConfig()
        const response = await $fetch<AuthResponse>(`${config.public.apiBase}/api/auth/register`, {
          method: 'POST',
          body: { email, password, display_name: displayName },
        })

        this.user = response.user
        this.tokens = response.tokens
        this.saveTokens()

        return response
      } catch (err: any) {
        this.error = err.data?.message || 'Registration failed'
        throw err
      } finally {
        this.loading = false
      }
    },

    async login(email: string, password: string) {
      this.loading = true
      this.error = null

      try {
        const config = useRuntimeConfig()
        const response = await $fetch<AuthResponse>(`${config.public.apiBase}/api/auth/login`, {
          method: 'POST',
          body: { email, password },
        })

        this.user = response.user
        this.tokens = response.tokens
        this.saveTokens()

        return response
      } catch (err: any) {
        this.error = err.data?.message || 'Login failed'
        throw err
      } finally {
        this.loading = false
      }
    },

    async refreshToken() {
      if (!this.tokens?.refresh_token) return

      try {
        const config = useRuntimeConfig()
        const response = await $fetch<TokenPair>(`${config.public.apiBase}/api/auth/refresh`, {
          method: 'POST',
          body: { refresh_token: this.tokens.refresh_token },
        })

        this.tokens = response
        this.saveTokens()

        return response
      } catch {
        this.logout()
        throw new Error('Session expired')
      }
    },

    async fetchUser() {
      if (!this.tokens?.access_token) return

      try {
        const config = useRuntimeConfig()
        const response = await $fetch<User>(`${config.public.apiBase}/api/users/me`, {
          headers: {
            Authorization: `Bearer ${this.tokens.access_token}`,
          },
        })

        this.user = response
        return response
      } catch {
        this.logout()
      }
    },

    logout() {
      this.user = null
      this.tokens = null
      this.clearTokens()
      navigateTo('/login')
    },

    saveTokens() {
      if (import.meta.client && this.tokens) {
        localStorage.setItem('mono_tokens', JSON.stringify(this.tokens))
      }
    },

    loadTokens() {
      if (import.meta.client) {
        const stored = localStorage.getItem('mono_tokens')
        if (stored) {
          try {
            this.tokens = JSON.parse(stored)
          } catch {
            this.clearTokens()
          }
        }
      }
    },

    clearTokens() {
      if (import.meta.client) {
        localStorage.removeItem('mono_tokens')
      }
    },

    async init() {
      this.loadTokens()
      if (this.tokens?.access_token) {
        await this.fetchUser()
      }
    },
  },
})
