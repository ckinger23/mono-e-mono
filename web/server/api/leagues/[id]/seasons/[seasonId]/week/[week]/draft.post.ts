import { requireAuth, getSupabaseClient, getRandomTeam, TOTAL_PICKS } from '~/server/utils/supabase'

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

  // Check for existing draft
  const { data: existingDraft } = await client
    .from('weekly_drafts')
    .select('*')
    .eq('season_id', seasonId)
    .eq('member_id', membership.id)
    .eq('week', week)
    .single()

  if (existingDraft) {
    if (existingDraft.status === 'complete') {
      throw createError({
        statusCode: 400,
        statusMessage: 'Draft already completed for this week',
      })
    }
    // Return existing in-progress draft
    return existingDraft
  }

  // Draw first random team
  const firstTeam = getRandomTeam()

  // Create new draft
  const { data: draft, error: draftError } = await client
    .from('weekly_drafts')
    .insert({
      season_id: seasonId,
      member_id: membership.id,
      week,
      status: 'in_progress',
      current_pick: 1,
      current_team: firstTeam.abbrev,
      started_at: new Date().toISOString(),
    })
    .select()
    .single()

  if (draftError) {
    throw createError({
      statusCode: 500,
      statusMessage: draftError.message,
    })
  }

  return {
    ...draft,
    current_team_info: firstTeam,
    picks: [],
    total_picks: TOTAL_PICKS,
  }
})
