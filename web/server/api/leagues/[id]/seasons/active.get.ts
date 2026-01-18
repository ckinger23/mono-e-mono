import { requireAuth, getSupabaseClient } from '~/server/utils/supabase'

export default defineEventHandler(async (event) => {
  await requireAuth(event)
  const client = await getSupabaseClient(event)
  const leagueId = getRouterParam(event, 'id')

  const { data: season, error } = await client
    .from('seasons')
    .select('*')
    .eq('league_id', leagueId)
    .eq('status', 'active')
    .single()

  if (error) {
    throw createError({
      statusCode: 404,
      statusMessage: 'No active season found',
    })
  }

  return season
})
