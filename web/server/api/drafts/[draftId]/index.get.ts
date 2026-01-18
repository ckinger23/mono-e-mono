import { requireAuth, getSupabaseClient, getNeededPositions, NFL_TEAMS, TOTAL_PICKS } from '~/server/utils/supabase'

export default defineEventHandler(async (event) => {
  const user = await requireAuth(event)
  const client = await getSupabaseClient(event)
  const draftId = getRouterParam(event, 'draftId')

  // Get draft with picks
  const { data: draft, error: draftError } = await client
    .from('weekly_drafts')
    .select(`
      *,
      member:league_members!inner(user_id, team_name),
      picks:draft_picks(*)
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

  const picks = (draft.picks as any[]) || []
  const currentTeamInfo = draft.current_team
    ? NFL_TEAMS.find(t => t.abbrev === draft.current_team)
    : null

  return {
    id: draft.id,
    status: draft.status,
    week: draft.week,
    current_pick: draft.current_pick,
    current_team: draft.current_team,
    current_team_info: currentTeamInfo,
    picks: picks.sort((a: any, b: any) => a.pick_number - b.pick_number),
    total_picks: TOTAL_PICKS,
    needed_positions: draft.status === 'complete' ? [] : getNeededPositions(picks),
    started_at: draft.started_at,
    completed_at: draft.completed_at,
  }
})
