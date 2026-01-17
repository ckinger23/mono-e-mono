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
