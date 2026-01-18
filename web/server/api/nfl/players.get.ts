import { getSupabaseClient } from '~/server/utils/supabase'

export default defineEventHandler(async (event) => {
  const client = await getSupabaseClient(event)
  const query = getQuery(event)

  const team = query.team as string
  const position = query.position as string

  let dbQuery = client
    .from('nfl_players')
    .select('*')
    .not('team', 'is', null)

  if (team) {
    dbQuery = dbQuery.eq('team', team.toUpperCase())
  }

  if (position) {
    dbQuery = dbQuery.eq('position', position.toUpperCase())
  }

  const { data: players, error } = await dbQuery
    .order('position')
    .order('name')
    .limit(200)

  if (error) {
    throw createError({
      statusCode: 500,
      statusMessage: error.message,
    })
  }

  return players
})
