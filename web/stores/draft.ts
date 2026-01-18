import { defineStore } from 'pinia'

interface DraftPick {
  id: string
  pick_number: number
  nfl_player_id: string
  player_name: string
  position: string
  team_drawn: string
  picked_at: string
}

interface TeamInfo {
  abbrev: string
  name: string
}

interface Player {
  id: string
  name: string
  position: string
  team: string
  status?: string
}

interface DraftState {
  id: string
  status: 'not_started' | 'pending' | 'in_progress' | 'complete'
  week: number
  current_pick: number
  current_team: string | null
  current_team_info: TeamInfo | null
  picks: DraftPick[]
  total_picks: number
  needed_positions: string[]
  started_at: string | null
  completed_at: string | null
}

export const useDraftStore = defineStore('draft', () => {
  const supabase = useSupabaseClient()

  const draftState = ref<DraftState | null>(null)
  const availablePlayers = ref<Player[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Computed
  const currentPick = computed(() => draftState.value?.current_pick || 1)
  const totalPicks = computed(() => draftState.value?.total_picks || 7)
  const neededPositions = computed(() => draftState.value?.needed_positions || [])
  const isComplete = computed(() => draftState.value?.status === 'complete')
  const roster = computed(() => draftState.value?.picks || [])
  const currentTeam = computed(() => draftState.value?.current_team_info)

  const progress = computed(() => {
    if (!draftState.value) return 0
    return ((draftState.value.current_pick - 1) / draftState.value.total_picks) * 100
  })

  // Realtime subscription
  let realtimeChannel: any = null

  async function fetchDraft(leagueId: string, seasonId: string, week: number) {
    loading.value = true
    error.value = null

    try {
      const data = await $fetch<DraftState>(
        `/api/leagues/${leagueId}/seasons/${seasonId}/week/${week}/draft`
      )

      draftState.value = data
      return data
    } catch (err: any) {
      if (err.statusCode !== 404) {
        error.value = err.data?.message || 'Failed to fetch draft'
      }
      throw err
    } finally {
      loading.value = false
    }
  }

  async function startDraft(leagueId: string, seasonId: string, week: number) {
    loading.value = true
    error.value = null

    try {
      const data = await $fetch<DraftState>(
        `/api/leagues/${leagueId}/seasons/${seasonId}/week/${week}/draft`,
        { method: 'POST' }
      )

      draftState.value = {
        ...data,
        picks: [],
        needed_positions: ['QB', 'RB', 'WR', 'TE', 'DEF'],
      }

      // Load players for the first team
      if (data.current_team) {
        await loadPlayersForTeam(data.current_team)
      }

      // Subscribe to realtime updates
      subscribeToUpdates(data.id)

      return data
    } catch (err: any) {
      error.value = err.data?.message || 'Failed to start draft'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function loadDraftState(draftId: string) {
    loading.value = true
    error.value = null

    try {
      const data = await $fetch<DraftState>(`/api/drafts/${draftId}`)
      draftState.value = data

      if (data.current_team) {
        await loadPlayersForTeam(data.current_team)
      }

      subscribeToUpdates(draftId)

      return data
    } catch (err: any) {
      error.value = err.data?.message || 'Failed to load draft'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function loadPlayersForTeam(teamAbbrev: string) {
    try {
      const players = await $fetch<Player[]>(`/api/nfl/players?team=${teamAbbrev}`)
      availablePlayers.value = players
    } catch (err: any) {
      console.error('Failed to load players:', err)
      availablePlayers.value = []
    }
  }

  async function drawTeam(draftId: string) {
    loading.value = true
    error.value = null

    try {
      const data = await $fetch<{ team: TeamInfo; current_pick: number }>(
        `/api/drafts/${draftId}/draw`,
        { method: 'POST' }
      )

      if (draftState.value) {
        draftState.value.current_team = data.team.abbrev
        draftState.value.current_team_info = data.team
        draftState.value.current_pick = data.current_pick
      }

      await loadPlayersForTeam(data.team.abbrev)

      return data
    } catch (err: any) {
      error.value = err.data?.message || 'Failed to draw team'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function makePick(draftId: string, playerId: string, playerName: string, position: string) {
    loading.value = true
    error.value = null

    try {
      const data = await $fetch<{
        pick: DraftPick
        draft: DraftState
        next_team: TeamInfo | null
        is_complete: boolean
        needed_positions: string[]
      }>(`/api/drafts/${draftId}/pick`, {
        method: 'POST',
        body: { player_id: playerId, player_name: playerName, position },
      })

      // Update local state
      draftState.value = {
        ...data.draft,
        current_team_info: data.next_team,
        needed_positions: data.needed_positions,
      }

      // Load players for next team
      if (data.next_team) {
        await loadPlayersForTeam(data.next_team.abbrev)
      } else {
        availablePlayers.value = []
      }

      return data
    } catch (err: any) {
      error.value = err.data?.message || 'Failed to make pick'
      throw err
    } finally {
      loading.value = false
    }
  }

  function subscribeToUpdates(draftId: string) {
    // Unsubscribe from previous channel
    if (realtimeChannel) {
      supabase.removeChannel(realtimeChannel)
    }

    // Subscribe to draft updates
    realtimeChannel = supabase
      .channel(`draft:${draftId}`)
      .on(
        'postgres_changes',
        {
          event: '*',
          schema: 'public',
          table: 'weekly_drafts',
          filter: `id=eq.${draftId}`,
        },
        (payload) => {
          console.log('Draft update:', payload)
          if (payload.new && draftState.value) {
            draftState.value = {
              ...draftState.value,
              ...(payload.new as any),
            }
          }
        }
      )
      .on(
        'postgres_changes',
        {
          event: 'INSERT',
          schema: 'public',
          table: 'draft_picks',
          filter: `weekly_draft_id=eq.${draftId}`,
        },
        (payload) => {
          console.log('New pick:', payload)
          if (payload.new && draftState.value) {
            const newPick = payload.new as DraftPick
            if (!draftState.value.picks.find(p => p.id === newPick.id)) {
              draftState.value.picks.push(newPick)
            }
          }
        }
      )
      .subscribe()
  }

  function unsubscribe() {
    if (realtimeChannel) {
      supabase.removeChannel(realtimeChannel)
      realtimeChannel = null
    }
  }

  function setError(err: string) {
    error.value = err
  }

  function clearError() {
    error.value = null
  }

  function reset() {
    unsubscribe()
    draftState.value = null
    availablePlayers.value = []
    loading.value = false
    error.value = null
  }

  return {
    // State
    draftState,
    availablePlayers,
    loading,
    error,

    // Computed
    currentPick,
    totalPicks,
    neededPositions,
    isComplete,
    roster,
    currentTeam,
    progress,

    // Actions
    fetchDraft,
    startDraft,
    loadDraftState,
    loadPlayersForTeam,
    drawTeam,
    makePick,
    subscribeToUpdates,
    unsubscribe,
    setError,
    clearError,
    reset,
  }
})
