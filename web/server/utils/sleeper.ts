// Sleeper API client for NFL data

const SLEEPER_BASE = 'https://api.sleeper.app/v1'

export interface SleeperPlayer {
  player_id: string
  first_name: string
  last_name: string
  full_name: string
  position: string
  team: string
  status: string
  injury_status: string
  active: boolean
}

export interface SleeperStats {
  player_id?: string
  // Passing
  pass_yd?: number
  pass_td?: number
  pass_int?: number
  pass_2pt?: number
  // Rushing
  rush_yd?: number
  rush_td?: number
  rush_2pt?: number
  // Receiving
  rec?: number
  rec_yd?: number
  rec_td?: number
  rec_2pt?: number
  // Fumbles
  fum_lost?: number
  // Defense
  def_td?: number
  def_int?: number
  sack?: number
  fum_rec?: number
  safe?: number
  pts_allow?: number
  blk_kick?: number
}

export interface NFLState {
  week: number
  season: string
  season_type: string
  display_week: number
}

// Fetch all NFL players
export async function fetchAllPlayers(): Promise<Record<string, SleeperPlayer>> {
  const response = await fetch(`${SLEEPER_BASE}/players/nfl`)
  if (!response.ok) {
    throw new Error(`Failed to fetch players: ${response.statusText}`)
  }
  return response.json()
}

// Fetch weekly stats
export async function fetchWeeklyStats(season: number, week: number): Promise<Record<string, SleeperStats>> {
  const response = await fetch(`${SLEEPER_BASE}/stats/nfl/regular/${season}/${week}`)
  if (!response.ok) {
    throw new Error(`Failed to fetch stats: ${response.statusText}`)
  }
  return response.json()
}

// Fetch NFL state (current week/season)
export async function fetchNFLState(): Promise<NFLState> {
  const response = await fetch(`${SLEEPER_BASE}/state/nfl`)
  if (!response.ok) {
    throw new Error(`Failed to fetch NFL state: ${response.statusText}`)
  }
  return response.json()
}

// Calculate fantasy points from stats
export function calculatePoints(stats: SleeperStats, scoringType: 'ppr' | 'standard' | 'half_ppr' = 'ppr'): number {
  let points = 0

  // Passing
  points += (stats.pass_yd || 0) * 0.04      // 1 point per 25 yards
  points += (stats.pass_td || 0) * 4          // 4 points per TD
  points += (stats.pass_int || 0) * -2        // -2 points per INT
  points += (stats.pass_2pt || 0) * 2         // 2 points per 2PT

  // Rushing
  points += (stats.rush_yd || 0) * 0.1        // 1 point per 10 yards
  points += (stats.rush_td || 0) * 6          // 6 points per TD
  points += (stats.rush_2pt || 0) * 2         // 2 points per 2PT

  // Receiving
  if (scoringType === 'ppr') {
    points += (stats.rec || 0) * 1            // 1 point per reception
  } else if (scoringType === 'half_ppr') {
    points += (stats.rec || 0) * 0.5          // 0.5 points per reception
  }
  points += (stats.rec_yd || 0) * 0.1         // 1 point per 10 yards
  points += (stats.rec_td || 0) * 6           // 6 points per TD
  points += (stats.rec_2pt || 0) * 2          // 2 points per 2PT

  // Fumbles
  points += (stats.fum_lost || 0) * -2        // -2 points per fumble lost

  // Defense
  points += (stats.def_td || 0) * 6           // 6 points per defensive TD
  points += (stats.def_int || 0) * 2          // 2 points per INT
  points += (stats.sack || 0) * 1             // 1 point per sack
  points += (stats.fum_rec || 0) * 2          // 2 points per fumble recovery
  points += (stats.safe || 0) * 2             // 2 points per safety
  points += (stats.blk_kick || 0) * 2         // 2 points per blocked kick

  // Points allowed (defense)
  const ptsAllowed = stats.pts_allow || 0
  if (ptsAllowed === 0) points += 10
  else if (ptsAllowed <= 6) points += 7
  else if (ptsAllowed <= 13) points += 4
  else if (ptsAllowed <= 20) points += 1
  else if (ptsAllowed <= 27) points += 0
  else if (ptsAllowed <= 34) points += -1
  else points += -4

  return Math.round(points * 100) / 100
}
