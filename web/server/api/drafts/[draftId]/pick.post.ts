import { requireAuth, getSupabaseClient, getNeededPositions, getRandomTeam, TOTAL_PICKS, NFL_TEAMS } from '~/server/utils/supabase'

export default defineEventHandler(async (event) => {
  const user = await requireAuth(event)
  const client = await getSupabaseClient(event)
  const draftId = getRouterParam(event, 'draftId')
  const body = await readBody(event)

  if (!body.player_id || !body.player_name || !body.position) {
    throw createError({
      statusCode: 400,
      statusMessage: 'player_id, player_name, and position are required',
    })
  }

  // Get draft with picks
  const { data: draft, error: draftError } = await client
    .from('weekly_drafts')
    .select(`
      *,
      member:league_members!inner(user_id),
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

  if (draft.status === 'complete') {
    throw createError({
      statusCode: 400,
      statusMessage: 'Draft is already complete',
    })
  }

  if (!draft.current_team) {
    throw createError({
      statusCode: 400,
      statusMessage: 'No team drawn yet - draw a team first',
    })
  }

  // Verify position is still needed
  const currentPicks = (draft.picks as any[]) || []
  const neededPositions = getNeededPositions(currentPicks)

  if (!neededPositions.includes(body.position)) {
    throw createError({
      statusCode: 400,
      statusMessage: `Position ${body.position} is not needed. Needed: ${neededPositions.join(', ')}`,
    })
  }

  // Insert pick
  const { data: pick, error: pickError } = await client
    .from('draft_picks')
    .insert({
      weekly_draft_id: draftId,
      pick_number: draft.current_pick,
      nfl_player_id: body.player_id,
      player_name: body.player_name,
      position: body.position,
      team_drawn: draft.current_team,
    })
    .select()
    .single()

  if (pickError) {
    throw createError({
      statusCode: 500,
      statusMessage: pickError.message,
    })
  }

  const newPickNumber = draft.current_pick + 1
  const isComplete = newPickNumber > TOTAL_PICKS

  // Get teams used so far for next draw
  const usedTeams = [...currentPicks.map((p: any) => p.team_drawn), draft.current_team]

  // Update draft
  const updates: any = {
    current_pick: newPickNumber,
    current_team: null, // Clear current team
  }

  if (isComplete) {
    updates.status = 'complete'
    updates.completed_at = new Date().toISOString()
  } else {
    // Draw next team automatically
    const nextTeam = getRandomTeam(usedTeams)
    updates.current_team = nextTeam.abbrev
  }

  const { error: updateError } = await client
    .from('weekly_drafts')
    .update(updates)
    .eq('id', draftId)

  if (updateError) {
    throw createError({
      statusCode: 500,
      statusMessage: updateError.message,
    })
  }

  // Get updated draft state
  const { data: updatedDraft } = await client
    .from('weekly_drafts')
    .select(`
      *,
      picks:draft_picks(*)
    `)
    .eq('id', draftId)
    .single()

  return {
    pick,
    draft: updatedDraft,
    next_team: updates.current_team ? NFL_TEAMS.find(t => t.abbrev === updates.current_team) : null,
    is_complete: isComplete,
    needed_positions: isComplete ? [] : getNeededPositions([...currentPicks, pick]),
  }
})
