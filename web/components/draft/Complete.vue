<template>
  <div class="card">
    <div class="card-header bg-green-600 text-white text-center">
      <h3 class="text-2xl font-bold">Draft Complete!</h3>
      <p class="text-green-200 mt-1">Your roster is locked in for this week</p>
    </div>

    <div class="card-body">
      <!-- Score -->
      <div class="text-center py-6 border-b border-gray-200">
        <p class="text-sm text-gray-500 mb-2">Projected Points</p>
        <p class="text-5xl font-bold text-gray-900">{{ totalPoints.toFixed(2) }}</p>
      </div>

      <!-- Roster Summary -->
      <div class="py-4">
        <h4 class="text-sm font-semibold text-gray-700 uppercase tracking-wide mb-3">
          Your Roster
        </h4>
        <div class="space-y-2">
          <div
            v-for="pick in roster"
            :key="pick.pick_number"
            class="flex items-center justify-between py-2"
          >
            <div class="flex items-center">
              <span
                :class="[
                  'inline-flex items-center justify-center w-8 h-8 rounded text-xs font-bold mr-3',
                  positionColors[pick.position] || 'bg-gray-100 text-gray-800'
                ]"
              >
                {{ pick.position }}
              </span>
              <div>
                <div class="font-medium text-gray-900">{{ pick.player_name }}</div>
                <div class="text-xs text-gray-500">{{ pick.team_drawn }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Actions -->
      <div class="pt-4 border-t border-gray-200">
        <button @click="$emit('close')" class="btn-primary w-full">
          Back to League
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { DraftPick } from '~/types'

defineProps<{
  roster: DraftPick[]
  totalPoints: number
}>()

defineEmits<{
  close: []
}>()

const positionColors: Record<string, string> = {
  QB: 'bg-red-100 text-red-800',
  RB: 'bg-blue-100 text-blue-800',
  WR: 'bg-green-100 text-green-800',
  TE: 'bg-yellow-100 text-yellow-800',
  DEF: 'bg-purple-100 text-purple-800',
}
</script>
