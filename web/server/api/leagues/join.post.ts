import { requireAuth, getSupabaseClient } from '~/server/utils/supabase'

export default defineEventHandler(async (event) => {
  const user = await requireAuth(event)
  const client = await getSupabaseClient(event)
  const body = await readBody(event)

  if (!body.invite_code?.trim()) {
    throw createError({
      statusCode: 400,
      statusMessage: 'Invite code is required',
    })
  }

  if (!body.team_name?.trim()) {
    throw createError({
      statusCode: 400,
      statusMessage: 'Team name is required',
    })
  }

  // Find league by invite code
  const { data: league, error: leagueError } = await client
    .from('leagues')
    .select('*, league_members(count)')
    .eq('invite_code', body.invite_code.trim().toUpperCase())
    .single()

  if (leagueError || !league) {
    throw createError({
      statusCode: 404,
      statusMessage: 'League not found with that invite code',
    })
  }

  // Check if already a member
  const { data: existing } = await client
    .from('league_members')
    .select('id')
    .eq('league_id', league.id)
    .eq('user_id', user.id)
    .single()

  if (existing) {
    throw createError({
      statusCode: 400,
      statusMessage: 'You are already a member of this league',
    })
  }

  // Check if league is full
  const memberCount = (league.league_members as any)?.[0]?.count || 0
  if (memberCount >= league.max_members) {
    throw createError({
      statusCode: 400,
      statusMessage: 'This league is full',
    })
  }

  // Join league
  const { data: membership, error: joinError } = await client
    .from('league_members')
    .insert({
      league_id: league.id,
      user_id: user.id,
      team_name: body.team_name.trim(),
    })
    .select()
    .single()

  if (joinError) {
    throw createError({
      statusCode: 500,
      statusMessage: joinError.message,
    })
  }

  return {
    league,
    membership,
  }
})
