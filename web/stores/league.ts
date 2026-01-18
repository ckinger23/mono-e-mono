import { defineStore } from 'pinia'

interface League {
  id: string
  name: string
  invite_code: string
  commissioner_id: string
  max_members: number
  scoring_type: string
  created_at: string
  is_commissioner?: boolean
  my_team_name?: string
  member_count?: number
}

interface LeagueMember {
  id: string
  user_id: string
  team_name: string
  joined_at: string
  profile?: {
    display_name: string
    avatar_url: string | null
  }
}

interface Season {
  id: string
  league_id: string
  year: number
  current_week: number
  status: string
  created_at: string
}

interface Standing {
  id: string
  member_id: string
  weekly_wins: number
  total_points: number
  best_week: number
  weeks_played: number
  current_rank: number
  member?: LeagueMember
}

export const useLeagueStore = defineStore('league', () => {
  const leagues = ref<League[]>([])
  const currentLeague = ref<League | null>(null)
  const members = ref<LeagueMember[]>([])
  const seasons = ref<Season[]>([])
  const currentSeason = ref<Season | null>(null)
  const standings = ref<Standing[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchLeagues() {
    loading.value = true
    error.value = null

    try {
      leagues.value = await $fetch<League[]>('/api/leagues')
    } catch (err: any) {
      error.value = err.data?.message || 'Failed to fetch leagues'
    } finally {
      loading.value = false
    }
  }

  async function fetchLeague(leagueId: string) {
    loading.value = true
    error.value = null

    try {
      currentLeague.value = await $fetch<League>(`/api/leagues/${leagueId}`)
      return currentLeague.value
    } catch (err: any) {
      error.value = err.data?.message || 'Failed to fetch league'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function createLeague(name: string, teamName: string, maxMembers = 10, scoringType = 'ppr') {
    loading.value = true
    error.value = null

    try {
      const league = await $fetch<League>('/api/leagues', {
        method: 'POST',
        body: { name, team_name: teamName, max_members: maxMembers, scoring_type: scoringType },
      })

      leagues.value.unshift(league)
      return league
    } catch (err: any) {
      error.value = err.data?.message || 'Failed to create league'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function joinLeague(inviteCode: string, teamName: string) {
    loading.value = true
    error.value = null

    try {
      const result = await $fetch<{ league: League; membership: LeagueMember }>('/api/leagues/join', {
        method: 'POST',
        body: { invite_code: inviteCode, team_name: teamName },
      })

      // Refresh leagues list
      await fetchLeagues()

      return result
    } catch (err: any) {
      error.value = err.data?.message || 'Failed to join league'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchMembers(leagueId: string) {
    try {
      members.value = await $fetch<LeagueMember[]>(`/api/leagues/${leagueId}/members`)
      return members.value
    } catch (err: any) {
      error.value = err.data?.message || 'Failed to fetch members'
      throw err
    }
  }

  async function fetchSeasons(leagueId: string) {
    try {
      seasons.value = await $fetch<Season[]>(`/api/leagues/${leagueId}/seasons`)
      return seasons.value
    } catch (err: any) {
      error.value = err.data?.message || 'Failed to fetch seasons'
      throw err
    }
  }

  async function fetchActiveSeason(leagueId: string) {
    try {
      currentSeason.value = await $fetch<Season>(`/api/leagues/${leagueId}/seasons/active`)
      return currentSeason.value
    } catch {
      currentSeason.value = null
      return null
    }
  }

  async function fetchStandings(leagueId: string, seasonId: string) {
    try {
      standings.value = await $fetch<Standing[]>(`/api/leagues/${leagueId}/seasons/${seasonId}/standings`)
      return standings.value
    } catch (err: any) {
      error.value = err.data?.message || 'Failed to fetch standings'
      throw err
    }
  }

  function clearCurrentLeague() {
    currentLeague.value = null
    members.value = []
    seasons.value = []
    currentSeason.value = null
    standings.value = []
  }

  function clearError() {
    error.value = null
  }

  return {
    // State
    leagues,
    currentLeague,
    members,
    seasons,
    currentSeason,
    standings,
    loading,
    error,

    // Actions
    fetchLeagues,
    fetchLeague,
    createLeague,
    joinLeague,
    fetchMembers,
    fetchSeasons,
    fetchActiveSeason,
    fetchStandings,
    clearCurrentLeague,
    clearError,
  }
})
