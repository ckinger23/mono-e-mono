import { requireAuth, getSupabaseClient, getRandomTeam, NFL_TEAMS } from '~/server/utils/supabase'

export default defineEventHandler(async (event) => {
  const user = await requireAuth(event)
  const client = await getSupabaseClient(event)
  const draftId = getRouterParam(event, 'draftId')

  // Get draft with picks
  const { data: draft, error: draftError } = await client
    .from('weekly_drafts')
    .select(`
      *,
      member:league_members!inner(user_id),
      picks:draft_picks(team_drawn)
    `)
    .eq('id', draftId)
    .single()

  if (draftError || !draft) {
    throw createError({
      statusCode: 404,
      statusMessage: 'Draft not found',
    })
  }

  // Verify ownership
  if ((draft.member as any).user_id !== user.id) {
    throw createError({
      statusCode: 403,
      statusMessage: 'This is not your draft',
    })
  }

  if (draft.status === 'complete') {
    throw createError({
      statusCode: 400,
      statusMessage: 'Draft is already complete',
    })
  }

  // If there's already a current team waiting for a pick, return it
  if (draft.current_team) {
    const teamInfo = NFL_TEAMS.find(t => t.abbrev === draft.current_team)
    return {
      team: teamInfo,
      current_pick: draft.current_pick,
    }
  }

  // Get previously drawn teams
  const usedTeams = (draft.picks as any[])?.map(p => p.team_drawn) || []

  // Draw new team
  const newTeam = getRandomTeam(usedTeams)

  // Update draft with new team
  const { error: updateError } = await client
    .from('weekly_drafts')
    .update({ current_team: newTeam.abbrev })
    .eq('id', draftId)

  if (updateError) {
    throw createError({
      statusCode: 500,
      statusMessage: updateError.message,
    })
  }

  return {
    team: newTeam,
    current_pick: draft.current_pick,
  }
})
