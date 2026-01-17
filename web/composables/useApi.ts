import type { UseFetchOptions } from 'nuxt/app'

export function useApi<T>(url: string, options: UseFetchOptions<T> = {}) {
  const config = useRuntimeConfig()
  const authStore = useAuthStore()

  const defaults: UseFetchOptions<T> = {
    baseURL: config.public.apiBase as string,
    key: url,
    headers: authStore.accessToken
      ? { Authorization: `Bearer ${authStore.accessToken}` }
      : {},
    onResponseError({ response }) {
      if (response.status === 401) {
        authStore.logout()
      }
    },
  }

  const params = {
    ...defaults,
    ...options,
    headers: {
      ...defaults.headers,
      ...options.headers,
    },
  }

  return useFetch<T>(url, params as any)
}

// Plugin to provide $api for direct fetch calls
export default defineNuxtPlugin(() => {
  const config = useRuntimeConfig()

  const api = async <T>(url: string, options: RequestInit & { body?: any } = {}): Promise<T> => {
    const authStore = useAuthStore()

    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...(options.headers || {}),
    }

    if (authStore.accessToken) {
      (headers as Record<string, string>)['Authorization'] = `Bearer ${authStore.accessToken}`
    }

    const response = await fetch(`${config.public.apiBase}${url}`, {
      ...options,
      headers,
      body: options.body ? JSON.stringify(options.body) : undefined,
    })

    if (!response.ok) {
      if (response.status === 401) {
        authStore.logout()
      }
      const error = await response.json().catch(() => ({ message: 'Request failed' }))
      throw { data: error, status: response.status }
    }

    if (response.status === 204) {
      return {} as T
    }

    return response.json()
  }

  return {
    provide: {
      api,
    },
  }
})
