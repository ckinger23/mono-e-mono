import { requireAuth, getSupabaseClient } from '~/server/utils/supabase'

export default defineEventHandler(async (event) => {
  const user = await requireAuth(event)
  const client = await getSupabaseClient(event)
  const body = await readBody(event)

  if (!body.name?.trim()) {
    throw createError({
      statusCode: 400,
      statusMessage: 'League name is required',
    })
  }

  if (!body.team_name?.trim()) {
    throw createError({
      statusCode: 400,
      statusMessage: 'Team name is required',
    })
  }

  // Create league
  const { data: league, error: leagueError } = await client
    .from('leagues')
    .insert({
      name: body.name.trim(),
      commissioner_id: user.id,
      max_members: body.max_members || 10,
      scoring_type: body.scoring_type || 'ppr',
    })
    .select()
    .single()

  if (leagueError) {
    throw createError({
      statusCode: 500,
      statusMessage: leagueError.message,
    })
  }

  // Add creator as first member
  const { error: memberError } = await client
    .from('league_members')
    .insert({
      league_id: league.id,
      user_id: user.id,
      team_name: body.team_name.trim(),
    })

  if (memberError) {
    // Rollback league creation
    await client.from('leagues').delete().eq('id', league.id)
    throw createError({
      statusCode: 500,
      statusMessage: memberError.message,
    })
  }

  // Create first season
  const currentYear = new Date().getFullYear()
  const { error: seasonError } = await client
    .from('seasons')
    .insert({
      league_id: league.id,
      year: currentYear,
      current_week: 1,
      status: 'active',
    })

  if (seasonError) {
    console.error('Failed to create season:', seasonError)
  }

  return league
})
