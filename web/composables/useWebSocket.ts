import type { WSMessage, DraftState, TeamDrawn, PickConfirmed, DraftComplete } from '~/types'

interface UseWebSocketOptions {
  onOpen?: () => void
  onClose?: () => void
  onError?: (error: Event) => void
  onMessage?: (message: WSMessage) => void
}

export function useWebSocket(options: UseWebSocketOptions = {}) {
  const socket = ref<WebSocket | null>(null)
  const isConnected = ref(false)
  const error = ref<string | null>(null)

  const connect = (url: string) => {
    if (socket.value?.readyState === WebSocket.OPEN) {
      return
    }

    try {
      socket.value = new WebSocket(url)

      socket.value.onopen = () => {
        isConnected.value = true
        error.value = null
        options.onOpen?.()
      }

      socket.value.onclose = () => {
        isConnected.value = false
        options.onClose?.()
      }

      socket.value.onerror = (e) => {
        error.value = 'WebSocket connection error'
        options.onError?.(e)
      }

      socket.value.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data) as WSMessage
          options.onMessage?.(message)
        } catch {
          console.error('Failed to parse WebSocket message')
        }
      }
    } catch (e) {
      error.value = 'Failed to connect'
      console.error('WebSocket connection failed:', e)
    }
  }

  const disconnect = () => {
    if (socket.value) {
      socket.value.close()
      socket.value = null
    }
    isConnected.value = false
  }

  const send = (type: string, payload?: unknown) => {
    if (socket.value?.readyState === WebSocket.OPEN) {
      socket.value.send(JSON.stringify({ type, payload }))
    } else {
      console.error('WebSocket is not connected')
    }
  }

  onUnmounted(() => {
    disconnect()
  })

  return {
    socket,
    isConnected,
    error,
    connect,
    disconnect,
    send,
  }
}

// Draft-specific WebSocket composable
export function useDraftWebSocket(draftId: string) {
  const config = useRuntimeConfig()
  const authStore = useAuthStore()
  const draftStore = useDraftStore()

  const { socket, isConnected, error, connect, disconnect, send } = useWebSocket({
    onOpen() {
      draftStore.setConnected(true)
    },
    onClose() {
      draftStore.setConnected(false)
    },
    onError() {
      draftStore.setError('Connection lost')
    },
    onMessage(message) {
      handleMessage(message)
    },
  })

  const handleMessage = (message: WSMessage) => {
    switch (message.type) {
      case 'draft_state':
        draftStore.setDraftState(message.payload as DraftState)
        break

      case 'team_drawn':
        draftStore.setCurrentTeam(message.payload as TeamDrawn)
        break

      case 'pick_confirmed':
        const pickData = message.payload as PickConfirmed
        draftStore.addPick(pickData.pick)
        draftStore.setDraftState({
          ...draftStore.draftState!,
          current_pick: pickData.current_pick,
          picks: pickData.roster,
        })
        break

      case 'draft_complete':
        const completeData = message.payload as DraftComplete
        draftStore.setComplete(completeData.total_points)
        break

      case 'error':
        draftStore.setError((message.payload as { message: string }).message)
        break
    }
  }

  const connectDraft = () => {
    const wsUrl = `${config.public.wsBase}/ws/draft/${draftId}?token=${authStore.accessToken}`
    connect(wsUrl)
  }

  const drawTeam = () => {
    send('draw_team')
  }

  const makePick = (playerId: string, position: string) => {
    send('make_pick', { player_id: playerId, position })
  }

  const getState = () => {
    send('get_state')
  }

  return {
    socket,
    isConnected,
    error,
    connectDraft,
    disconnect,
    drawTeam,
    makePick,
    getState,
  }
}
