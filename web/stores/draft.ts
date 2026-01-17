import { defineStore } from 'pinia'
import type { WeeklyDraft, DraftState, TeamDrawn, DraftPick, Player } from '~/types'

interface DraftStoreState {
  currentDraft: WeeklyDraft | null
  draftState: DraftState | null
  currentTeam: TeamDrawn | null
  roster: DraftPick[]
  availablePlayers: Player[]
  isConnected: boolean
  isComplete: boolean
  totalPoints: number | null
  loading: boolean
  error: string | null
}

export const useDraftStore = defineStore('draft', {
  state: (): DraftStoreState => ({
    currentDraft: null,
    draftState: null,
    currentTeam: null,
    roster: [],
    availablePlayers: [],
    isConnected: false,
    isComplete: false,
    totalPoints: null,
    loading: false,
    error: null,
  }),

  getters: {
    currentPick: (state) => state.draftState?.current_pick || 1,
    totalPicks: (state) => state.draftState?.total_picks || 7,
    neededPositions: (state) => state.draftState?.needed_positions || [],
    progress: (state) => {
      if (!state.draftState) return 0
      return ((state.draftState.current_pick - 1) / state.draftState.total_picks) * 100
    },
  },

  actions: {
    async fetchDraft(leagueId: string, seasonId: string, week: number) {
      this.loading = true
      this.error = null

      try {
        const { $api } = useNuxtApp()
        const draft = await $api<WeeklyDraft>(
          `/api/leagues/${leagueId}/seasons/${seasonId}/week/${week}/draft`
        )

        this.currentDraft = draft
        if (draft.picks) {
          this.roster = draft.picks
        }

        return draft
      } catch (err: any) {
        this.error = err.data?.message || 'Failed to fetch draft'
        throw err
      } finally {
        this.loading = false
      }
    },

    async startDraft(leagueId: string, seasonId: string, week: number) {
      this.loading = true
      this.error = null

      try {
        const { $api } = useNuxtApp()
        const draft = await $api<WeeklyDraft>(
          `/api/leagues/${leagueId}/seasons/${seasonId}/week/${week}/draft`,
          { method: 'POST' }
        )

        this.currentDraft = draft
        return draft
      } catch (err: any) {
        this.error = err.data?.message || 'Failed to start draft'
        throw err
      } finally {
        this.loading = false
      }
    },

    setConnected(connected: boolean) {
      this.isConnected = connected
    },

    setDraftState(state: DraftState) {
      this.draftState = state
      this.roster = state.picks || []
      this.isComplete = state.status === 'complete'
    },

    setCurrentTeam(teamData: TeamDrawn) {
      this.currentTeam = teamData
      this.availablePlayers = teamData.available_players || []
    },

    addPick(pick: DraftPick) {
      this.roster.push(pick)
    },

    setComplete(totalPoints: number) {
      this.isComplete = true
      this.totalPoints = totalPoints
      this.currentTeam = null
      this.availablePlayers = []
    },

    setError(error: string) {
      this.error = error
    },

    clearError() {
      this.error = null
    },

    reset() {
      this.currentDraft = null
      this.draftState = null
      this.currentTeam = null
      this.roster = []
      this.availablePlayers = []
      this.isConnected = false
      this.isComplete = false
      this.totalPoints = null
      this.loading = false
      this.error = null
    },
  },
})
