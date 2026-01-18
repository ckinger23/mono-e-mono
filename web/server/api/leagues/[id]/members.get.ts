import { requireAuth, getSupabaseClient } from '~/server/utils/supabase'

export default defineEventHandler(async (event) => {
  await requireAuth(event)
  const client = await getSupabaseClient(event)
  const leagueId = getRouterParam(event, 'id')

  const { data: members, error } = await client
    .from('league_members')
    .select(`
      id,
      user_id,
      team_name,
      joined_at,
      profile:profiles(display_name, avatar_url)
    `)
    .eq('league_id', leagueId)
    .order('joined_at', { ascending: true })

  if (error) {
    throw createError({
      statusCode: 500,
      statusMessage: error.message,
    })
  }

  return members
})
