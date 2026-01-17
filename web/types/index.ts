// User types
export interface User {
  id: string
  email: string
  display_name: string
  avatar_url?: string
  created_at?: string
}

export interface TokenPair {
  access_token: string
  refresh_token: string
  expires_in: number
}

export interface AuthResponse {
  user: User
  tokens: TokenPair
}

// League types
export interface League {
  id: string
  name: string
  invite_code: string
  commissioner_id: string
  max_members: number
  scoring_type: 'ppr' | 'standard' | 'half_ppr'
  member_count: number
  created_at: string
}

export interface LeagueMember {
  id: string
  league_id: string
  user_id: string
  team_name: string
  joined_at: string
  display_name: string
  avatar_url?: string
}

// Season types
export interface Season {
  id: string
  league_id: string
  year: number
  current_week: number
  status: 'active' | 'complete'
  champion_member_id?: string
  created_at: string
}

export interface Standing {
  id: string
  season_id: string
  member_id: string
  weekly_wins: number
  total_points: number
  best_week: number
  weeks_played: number
  current_rank: number
  team_name: string
  display_name: string
  avatar_url?: string
}

// Draft types
export interface WeeklyDraft {
  id: string
  season_id: string
  member_id: string
  week: number
  status: 'pending' | 'in_progress' | 'complete' | 'not_started'
  started_at?: string
  completed_at?: string
  picks?: DraftPick[]
}

export interface DraftPick {
  id?: string
  pick_number: number
  nfl_player_id: string
  player_name: string
  position: string
  team_drawn: string
  picked_at?: string
}

export interface Player {
  id: string
  name: string
  position: 'QB' | 'RB' | 'WR' | 'TE' | 'DEF'
  team: string
}

export interface TeamInfo {
  name: string
  abbrev: string
}

export interface DraftState {
  draft_id: string
  season_id: string
  member_id: string
  week: number
  status: string
  current_pick: number
  total_picks: number
  picks: DraftPick[]
  needed_positions: string[]
}

export interface TeamDrawn {
  team: TeamInfo
  available_players: Player[]
  current_pick: number
  total_picks: number
}

export interface PickConfirmed {
  pick: DraftPick
  current_pick: number
  total_picks: number
  is_complete: boolean
  roster: DraftPick[]
}

export interface DraftComplete {
  draft_id: string
  week: number
  roster: DraftPick[]
  total_points: number
  message: string
}

// Weekly results
export interface WeeklyResult {
  id: string
  member_id: string
  week: number
  total_points: number
  weekly_rank: number
  is_weekly_winner: boolean
  team_name: string
  display_name: string
  avatar_url?: string
}

// API Error
export interface ApiError {
  error: string
  message: string
}

// WebSocket message types
export type WSMessageType =
  | 'draft_state'
  | 'team_drawn'
  | 'pick_confirmed'
  | 'draft_complete'
  | 'error'
  | 'draw_team'
  | 'make_pick'
  | 'get_state'

export interface WSMessage<T = unknown> {
  type: WSMessageType
  payload: T
}
