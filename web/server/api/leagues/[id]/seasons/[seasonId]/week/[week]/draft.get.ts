import { requireAuth, getSupabaseClient } from '~/server/utils/supabase'

export default defineEventHandler(async (event) => {
  const user = await requireAuth(event)
  const client = await getSupabaseClient(event)
  const leagueId = getRouterParam(event, 'id')
  const seasonId = getRouterParam(event, 'seasonId')
  const week = parseInt(getRouterParam(event, 'week') || '1')

  // Get user's league membership
  const { data: membership, error: memberError } = await client
    .from('league_members')
    .select('id')
    .eq('league_id', leagueId)
    .eq('user_id', user.id)
    .single()

  if (memberError || !membership) {
    throw createError({
      statusCode: 403,
      statusMessage: 'You are not a member of this league',
    })
  }

  // Get or check for existing draft
  const { data: draft, error: draftError } = await client
    .from('weekly_drafts')
    .select(`
      *,
      picks:draft_picks(*)
    `)
    .eq('season_id', seasonId)
    .eq('member_id', membership.id)
    .eq('week', week)
    .single()

  if (draftError && draftError.code !== 'PGRST116') {
    throw createError({
      statusCode: 500,
      statusMessage: draftError.message,
    })
  }

  // No draft exists yet
  if (!draft) {
    return {
      status: 'not_started',
      week,
      picks: [],
    }
  }

  return {
    id: draft.id,
    status: draft.status,
    week: draft.week,
    current_pick: draft.current_pick,
    current_team: draft.current_team,
    picks: draft.picks || [],
    started_at: draft.started_at,
    completed_at: draft.completed_at,
  }
})
