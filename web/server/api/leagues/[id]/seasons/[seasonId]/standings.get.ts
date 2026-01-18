import { requireAuth, getSupabaseClient } from '~/server/utils/supabase'

export default defineEventHandler(async (event) => {
  await requireAuth(event)
  const client = await getSupabaseClient(event)
  const seasonId = getRouterParam(event, 'seasonId')

  const { data: standings, error } = await client
    .from('standings')
    .select(`
      *,
      member:league_members(
        id,
        team_name,
        user_id,
        profile:profiles(display_name, avatar_url)
      )
    `)
    .eq('season_id', seasonId)
    .order('current_rank', { ascending: true })

  if (error) {
    throw createError({
      statusCode: 500,
      statusMessage: error.message,
    })
  }

  return standings
})
