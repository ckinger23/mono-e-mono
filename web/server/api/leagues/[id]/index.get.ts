import { requireAuth, getSupabaseClient } from '~/server/utils/supabase'

export default defineEventHandler(async (event) => {
  const user = await requireAuth(event)
  const client = await getSupabaseClient(event)
  const leagueId = getRouterParam(event, 'id')

  const { data: league, error } = await client
    .from('leagues')
    .select(`
      *,
      commissioner:profiles!leagues_commissioner_id_fkey(id, display_name, avatar_url),
      league_members(
        id,
        user_id,
        team_name,
        joined_at,
        profile:profiles(display_name, avatar_url)
      )
    `)
    .eq('id', leagueId)
    .single()

  if (error) {
    throw createError({
      statusCode: 404,
      statusMessage: 'League not found',
    })
  }

  // Get user's membership
  const membership = league.league_members?.find((m: any) => m.user_id === user.id)

  return {
    ...league,
    is_commissioner: league.commissioner_id === user.id,
    my_membership: membership,
    member_count: league.league_members?.length || 0,
  }
})
