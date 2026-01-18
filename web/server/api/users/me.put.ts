import { requireAuth, getSupabaseClient } from '~/server/utils/supabase'

export default defineEventHandler(async (event) => {
  const user = await requireAuth(event)
  const client = await getSupabaseClient(event)
  const body = await readBody(event)

  const updates: Record<string, any> = {
    updated_at: new Date().toISOString(),
  }

  if (body.display_name) {
    updates.display_name = body.display_name
  }

  if (body.avatar_url !== undefined) {
    updates.avatar_url = body.avatar_url
  }

  const { data, error } = await client
    .from('profiles')
    .update(updates)
    .eq('id', user.id)
    .select()
    .single()

  if (error) {
    throw createError({
      statusCode: 500,
      statusMessage: error.message,
    })
  }

  return data
})
