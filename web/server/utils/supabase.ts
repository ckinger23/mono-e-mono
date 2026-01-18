import { createClient } from '@supabase/supabase-js'
import type { H3Event } from 'h3'
import { serverSupabaseClient, serverSupabaseServiceRole, serverSupabaseUser } from '#supabase/server'

// Get authenticated Supabase client (respects RLS)
export async function getSupabaseClient(event: H3Event) {
  return serverSupabaseClient(event)
}

// Get service role client (bypasses RLS - use carefully!)
export async function getSupabaseAdmin(event: H3Event) {
  return serverSupabaseServiceRole(event)
}

// Get current authenticated user
export async function getUser(event: H3Event) {
  return serverSupabaseUser(event)
}

// Require authentication - throws 401 if not authenticated
export async function requireAuth(event: H3Event) {
  const user = await getUser(event)
  if (!user) {
    throw createError({
      statusCode: 401,
      statusMessage: 'Unauthorized',
    })
  }
  return user
}

// NFL Teams for random draw
export const NFL_TEAMS = [
  { abbrev: 'ARI', name: 'Arizona Cardinals' },
  { abbrev: 'ATL', name: 'Atlanta Falcons' },
  { abbrev: 'BAL', name: 'Baltimore Ravens' },
  { abbrev: 'BUF', name: 'Buffalo Bills' },
  { abbrev: 'CAR', name: 'Carolina Panthers' },
  { abbrev: 'CHI', name: 'Chicago Bears' },
  { abbrev: 'CIN', name: 'Cincinnati Bengals' },
  { abbrev: 'CLE', name: 'Cleveland Browns' },
  { abbrev: 'DAL', name: 'Dallas Cowboys' },
  { abbrev: 'DEN', name: 'Denver Broncos' },
  { abbrev: 'DET', name: 'Detroit Lions' },
  { abbrev: 'GB', name: 'Green Bay Packers' },
  { abbrev: 'HOU', name: 'Houston Texans' },
  { abbrev: 'IND', name: 'Indianapolis Colts' },
  { abbrev: 'JAX', name: 'Jacksonville Jaguars' },
  { abbrev: 'KC', name: 'Kansas City Chiefs' },
  { abbrev: 'LV', name: 'Las Vegas Raiders' },
  { abbrev: 'LAC', name: 'Los Angeles Chargers' },
  { abbrev: 'LAR', name: 'Los Angeles Rams' },
  { abbrev: 'MIA', name: 'Miami Dolphins' },
  { abbrev: 'MIN', name: 'Minnesota Vikings' },
  { abbrev: 'NE', name: 'New England Patriots' },
  { abbrev: 'NO', name: 'New Orleans Saints' },
  { abbrev: 'NYG', name: 'New York Giants' },
  { abbrev: 'NYJ', name: 'New York Jets' },
  { abbrev: 'PHI', name: 'Philadelphia Eagles' },
  { abbrev: 'PIT', name: 'Pittsburgh Steelers' },
  { abbrev: 'SF', name: 'San Francisco 49ers' },
  { abbrev: 'SEA', name: 'Seattle Seahawks' },
  { abbrev: 'TB', name: 'Tampa Bay Buccaneers' },
  { abbrev: 'TEN', name: 'Tennessee Titans' },
  { abbrev: 'WAS', name: 'Washington Commanders' },
]

// Roster requirements
export const ROSTER_REQUIREMENTS = ['QB', 'RB', 'RB', 'WR', 'WR', 'TE', 'DEF']
export const TOTAL_PICKS = 7

// Get a random NFL team
export function getRandomTeam(exclude: string[] = []): typeof NFL_TEAMS[0] {
  const available = NFL_TEAMS.filter(t => !exclude.includes(t.abbrev))
  if (available.length === 0) {
    // If all teams used, reset and pick from all
    return NFL_TEAMS[Math.floor(Math.random() * NFL_TEAMS.length)]
  }
  return available[Math.floor(Math.random() * available.length)]
}

// Calculate needed positions based on current picks
export function getNeededPositions(picks: { position: string }[]): string[] {
  const filled: Record<string, number> = {}
  for (const pick of picks) {
    filled[pick.position] = (filled[pick.position] || 0) + 1
  }

  const needed: string[] = []
  const requirements: Record<string, number> = { QB: 1, RB: 2, WR: 2, TE: 1, DEF: 1 }

  for (const [pos, count] of Object.entries(requirements)) {
    const have = filled[pos] || 0
    for (let i = have; i < count; i++) {
      needed.push(pos)
    }
  }

  return [...new Set(needed)] // Unique positions
}
