<template>
  <div
    :class="[
      'p-4 rounded-lg border transition-all',
      canSelect
        ? 'border-gray-200 hover:border-primary-500 hover:shadow-md cursor-pointer'
        : 'border-gray-100 bg-gray-50 opacity-50 cursor-not-allowed'
    ]"
    @click="handleClick"
  >
    <div class="flex items-center justify-between">
      <div class="flex items-center">
        <span
          :class="[
            'inline-flex items-center justify-center w-10 h-10 rounded-lg text-sm font-bold mr-3',
            positionColors[player.position] || 'bg-gray-100 text-gray-800'
          ]"
        >
          {{ player.position }}
        </span>
        <div>
          <div class="font-medium text-gray-900">{{ player.name }}</div>
          <div class="text-sm text-gray-500">{{ player.team }}</div>
        </div>
      </div>
      <button
        v-if="canSelect"
        class="btn-primary text-sm py-1 px-3"
        @click.stop="handleClick"
      >
        Draft
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Player } from '~/types'

const props = defineProps<{
  player: Player
  canSelect: boolean
}>()

const emit = defineEmits<{
  select: [player: Player]
}>()

const positionColors: Record<string, string> = {
  QB: 'bg-red-100 text-red-800',
  RB: 'bg-blue-100 text-blue-800',
  WR: 'bg-green-100 text-green-800',
  TE: 'bg-yellow-100 text-yellow-800',
  DEF: 'bg-purple-100 text-purple-800',
}

const handleClick = () => {
  if (props.canSelect) {
    emit('select', props.player)
  }
}
</script>
