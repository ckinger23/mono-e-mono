import { fetchNFLState } from '~/server/utils/sleeper'

export default defineEventHandler(async () => {
  const state = await fetchNFLState()
  return state
})
