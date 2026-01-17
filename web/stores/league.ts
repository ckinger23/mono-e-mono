import { defineStore } from 'pinia'
import type { League, LeagueMember, Season, Standing, WeeklyResult } from '~/types'

interface LeagueState {
  leagues: League[]
  currentLeague: League | null
  members: LeagueMember[]
  seasons: Season[]
  currentSeason: Season | null
  standings: Standing[]
  weeklyResults: WeeklyResult[]
  loading: boolean
  error: string | null
}

export const useLeagueStore = defineStore('league', {
  state: (): LeagueState => ({
    leagues: [],
    currentLeague: null,
    members: [],
    seasons: [],
    currentSeason: null,
    standings: [],
    weeklyResults: [],
    loading: false,
    error: null,
  }),

  actions: {
    async fetchLeagues() {
      this.loading = true
      this.error = null

      try {
        const { $api } = useNuxtApp()
        this.leagues = await $api<League[]>('/api/leagues')
      } catch (err: any) {
        this.error = err.data?.message || 'Failed to fetch leagues'
      } finally {
        this.loading = false
      }
    },

    async fetchLeague(leagueId: string) {
      this.loading = true
      this.error = null

      try {
        const { $api } = useNuxtApp()
        this.currentLeague = await $api<League>(`/api/leagues/${leagueId}`)
        return this.currentLeague
      } catch (err: any) {
        this.error = err.data?.message || 'Failed to fetch league'
        throw err
      } finally {
        this.loading = false
      }
    },

    async createLeague(name: string, teamName: string, maxMembers = 10, scoringType = 'ppr') {
      this.loading = true
      this.error = null

      try {
        const { $api } = useNuxtApp()
        const league = await $api<League>('/api/leagues', {
          method: 'POST',
          body: { name, team_name: teamName, max_members: maxMembers, scoring_type: scoringType },
        })

        this.leagues.unshift(league)
        return league
      } catch (err: any) {
        this.error = err.data?.message || 'Failed to create league'
        throw err
      } finally {
        this.loading = false
      }
    },

    async joinLeague(inviteCode: string, teamName: string) {
      this.loading = true
      this.error = null

      try {
        const { $api } = useNuxtApp()
        const member = await $api<LeagueMember>('/api/leagues/join', {
          method: 'POST',
          body: { invite_code: inviteCode, team_name: teamName },
        })

        // Refresh leagues list
        await this.fetchLeagues()

        return member
      } catch (err: any) {
        this.error = err.data?.message || 'Failed to join league'
        throw err
      } finally {
        this.loading = false
      }
    },

    async fetchMembers(leagueId: string) {
      try {
        const { $api } = useNuxtApp()
        this.members = await $api<LeagueMember[]>(`/api/leagues/${leagueId}/members`)
        return this.members
      } catch (err: any) {
        this.error = err.data?.message || 'Failed to fetch members'
        throw err
      }
    },

    async fetchSeasons(leagueId: string) {
      try {
        const { $api } = useNuxtApp()
        this.seasons = await $api<Season[]>(`/api/leagues/${leagueId}/seasons`)
        return this.seasons
      } catch (err: any) {
        this.error = err.data?.message || 'Failed to fetch seasons'
        throw err
      }
    },

    async fetchActiveSeason(leagueId: string) {
      try {
        const { $api } = useNuxtApp()
        this.currentSeason = await $api<Season>(`/api/leagues/${leagueId}/seasons/active`)
        return this.currentSeason
      } catch {
        this.currentSeason = null
        return null
      }
    },

    async createSeason(leagueId: string, year?: number) {
      this.loading = true

      try {
        const { $api } = useNuxtApp()
        const season = await $api<Season>(`/api/leagues/${leagueId}/seasons`, {
          method: 'POST',
          body: { year: year || new Date().getFullYear() },
        })

        this.seasons.unshift(season)
        this.currentSeason = season
        return season
      } catch (err: any) {
        this.error = err.data?.message || 'Failed to create season'
        throw err
      } finally {
        this.loading = false
      }
    },

    async fetchStandings(leagueId: string, seasonId: string) {
      try {
        const { $api } = useNuxtApp()
        this.standings = await $api<Standing[]>(`/api/leagues/${leagueId}/seasons/${seasonId}/standings`)
        return this.standings
      } catch (err: any) {
        this.error = err.data?.message || 'Failed to fetch standings'
        throw err
      }
    },

    async fetchWeeklyResults(leagueId: string, seasonId: string, week: number) {
      try {
        const { $api } = useNuxtApp()
        this.weeklyResults = await $api<WeeklyResult[]>(
          `/api/leagues/${leagueId}/seasons/${seasonId}/week/${week}/results`
        )
        return this.weeklyResults
      } catch (err: any) {
        this.error = err.data?.message || 'Failed to fetch weekly results'
        throw err
      }
    },

    clearCurrentLeague() {
      this.currentLeague = null
      this.members = []
      this.seasons = []
      this.currentSeason = null
      this.standings = []
      this.weeklyResults = []
    },
  },
})
