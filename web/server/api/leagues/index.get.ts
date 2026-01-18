import { requireAuth, getSupabaseClient } from '~/server/utils/supabase'

export default defineEventHandler(async (event) => {
  const user = await requireAuth(event)
  const client = await getSupabaseClient(event)

  // Get leagues where user is a member
  const { data: memberships, error: memberError } = await client
    .from('league_members')
    .select(`
      league_id,
      team_name,
      leagues (
        id,
        name,
        invite_code,
        commissioner_id,
        max_members,
        scoring_type,
        created_at
      )
    `)
    .eq('user_id', user.id)

  if (memberError) {
    throw createError({
      statusCode: 500,
      statusMessage: memberError.message,
    })
  }

  // Format response
  const leagues = memberships?.map(m => ({
    ...m.leagues,
    my_team_name: m.team_name,
    is_commissioner: (m.leagues as any)?.commissioner_id === user.id,
  })) || []

  return leagues
})
