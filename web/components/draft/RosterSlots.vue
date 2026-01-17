<template>
  <div class="card">
    <div class="card-header">
      <h3 class="text-lg font-semibold text-gray-900">Your Roster</h3>
    </div>
    <div class="divide-y divide-gray-200">
      <div
        v-for="(slot, index) in rosterSlots"
        :key="index"
        :class="[
          'px-4 py-3 flex items-center',
          slot.filled ? 'bg-white' : 'bg-gray-50'
        ]"
      >
        <span
          :class="[
            'inline-flex items-center justify-center w-10 h-10 rounded-lg text-sm font-bold mr-3',
            slot.filled ? positionColors[slot.position] : 'bg-gray-200 text-gray-500'
          ]"
        >
          {{ slot.position }}
        </span>
        <div v-if="slot.pick" class="flex-1">
          <div class="font-medium text-gray-900">{{ slot.pick.player_name }}</div>
          <div class="text-sm text-gray-500">{{ slot.pick.team_drawn }}</div>
        </div>
        <div v-else class="flex-1 text-gray-400 italic">
          Empty
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { DraftPick } from '~/types'

const props = defineProps<{
  picks: DraftPick[]
}>()

const positionColors: Record<string, string> = {
  QB: 'bg-red-100 text-red-800',
  RB: 'bg-blue-100 text-blue-800',
  WR: 'bg-green-100 text-green-800',
  TE: 'bg-yellow-100 text-yellow-800',
  DEF: 'bg-purple-100 text-purple-800',
}

const rosterRequirements = ['QB', 'RB', 'RB', 'WR', 'WR', 'TE', 'DEF']

const rosterSlots = computed(() => {
  const slots: { position: string; pick: DraftPick | null; filled: boolean }[] = []
  const picksByPosition: Record<string, DraftPick[]> = {}

  // Group picks by position
  for (const pick of props.picks) {
    if (!picksByPosition[pick.position]) {
      picksByPosition[pick.position] = []
    }
    picksByPosition[pick.position].push(pick)
  }

  // Fill slots
  for (const position of rosterRequirements) {
    const picks = picksByPosition[position] || []
    const pick = picks.shift() || null
    slots.push({
      position,
      pick,
      filled: !!pick,
    })
  }

  return slots
})
</script>
