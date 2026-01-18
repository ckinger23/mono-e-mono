import { requireAuth, getSupabaseClient } from '~/server/utils/supabase'

export default defineEventHandler(async (event) => {
  await requireAuth(event)
  const client = await getSupabaseClient(event)
  const leagueId = getRouterParam(event, 'id')

  const { data: seasons, error } = await client
    .from('seasons')
    .select('*')
    .eq('league_id', leagueId)
    .order('year', { ascending: false })

  if (error) {
    throw createError({
      statusCode: 500,
      statusMessage: error.message,
    })
  }

  return seasons
})
