import { getSupabaseAdmin } from '~/server/utils/supabase'
import { fetchAllPlayers } from '~/server/utils/sleeper'

// Admin endpoint to sync NFL players from Sleeper API
// In production, this should be protected by admin authentication
export default defineEventHandler(async (event) => {
  const client = await getSupabaseAdmin(event)

  console.log('Starting NFL player sync from Sleeper API...')

  const players = await fetchAllPlayers()
  const playerEntries = Object.entries(players)

  console.log(`Fetched ${playerEntries.length} players from Sleeper`)

  // Filter to fantasy-relevant positions
  const relevantPositions = new Set(['QB', 'RB', 'WR', 'TE', 'K', 'DEF'])

  const playersToUpsert = playerEntries
    .filter(([_, player]) => {
      return relevantPositions.has(player.position) && player.team
    })
    .map(([id, player]) => ({
      id,
      name: player.full_name || `${player.first_name} ${player.last_name}`,
      position: player.position,
      team: player.team,
      status: player.status || 'Active',
      updated_at: new Date().toISOString(),
    }))

  console.log(`Upserting ${playersToUpsert.length} fantasy-relevant players`)

  // Batch upsert in chunks of 500
  const chunkSize = 500
  let totalUpserted = 0
  let errors = 0

  for (let i = 0; i < playersToUpsert.length; i += chunkSize) {
    const chunk = playersToUpsert.slice(i, i + chunkSize)

    const { error } = await client
      .from('nfl_players')
      .upsert(chunk, { onConflict: 'id' })

    if (error) {
      console.error(`Error upserting chunk ${i / chunkSize}:`, error)
      errors++
    } else {
      totalUpserted += chunk.length
    }
  }

  console.log(`Sync complete. Upserted: ${totalUpserted}, Errors: ${errors}`)

  return {
    success: true,
    total_fetched: playerEntries.length,
    total_upserted: totalUpserted,
    errors,
  }
})
