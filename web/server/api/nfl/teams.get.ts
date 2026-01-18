import { NFL_TEAMS } from '~/server/utils/supabase'

export default defineEventHandler(() => {
  return NFL_TEAMS
})
