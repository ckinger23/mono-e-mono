<template>
  <div class="card">
    <div class="card-header bg-primary-600 text-white">
      <div class="flex items-center justify-between">
        <div>
          <h3 class="text-lg font-semibold">{{ team.name }}</h3>
          <p class="text-primary-200 text-sm">Select a player from this team</p>
        </div>
        <span class="text-2xl font-bold">{{ team.abbrev }}</span>
      </div>
    </div>

    <div class="card-body">
      <!-- Position Filter -->
      <div class="mb-4">
        <div class="flex flex-wrap gap-2">
          <button
            @click="selectedPosition = null"
            :class="[
              'px-3 py-1 rounded-full text-sm font-medium transition-colors',
              !selectedPosition
                ? 'bg-primary-600 text-white'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            ]"
          >
            All
          </button>
          <button
            v-for="pos in uniquePositions"
            :key="pos"
            @click="selectedPosition = pos"
            :class="[
              'px-3 py-1 rounded-full text-sm font-medium transition-colors',
              selectedPosition === pos
                ? 'bg-primary-600 text-white'
                : neededPositions.includes(pos)
                  ? 'bg-green-100 text-green-800 hover:bg-green-200'
                  : 'bg-gray-100 text-gray-400 cursor-not-allowed'
            ]"
            :disabled="!neededPositions.includes(pos)"
          >
            {{ pos }}
          </button>
        </div>
      </div>

      <!-- Player List -->
      <div class="space-y-2 max-h-96 overflow-y-auto">
        <DraftPlayerCard
          v-for="player in filteredPlayers"
          :key="player.id"
          :player="player"
          :can-select="neededPositions.includes(player.position)"
          @select="handleSelect"
        />

        <div v-if="filteredPlayers.length === 0" class="text-center py-8 text-gray-500">
          No available players at this position
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface TeamInfo {
  abbrev: string
  name: string
}

interface Player {
  id: string
  name: string
  position: string
  team: string
}

const props = defineProps<{
  team: TeamInfo
  availablePlayers: Player[]
  neededPositions: string[]
}>()

const emit = defineEmits<{
  pick: [playerId: string, playerName: string, position: string]
}>()

const selectedPosition = ref<string | null>(null)

const uniquePositions = computed(() => {
  const positions = new Set(props.availablePlayers.map(p => p.position))
  return Array.from(positions).sort()
})

const filteredPlayers = computed(() => {
  if (!selectedPosition.value) {
    return props.availablePlayers.filter(p => props.neededPositions.includes(p.position))
  }
  return props.availablePlayers.filter(p => p.position === selectedPosition.value)
})

const handleSelect = (player: Player) => {
  if (props.neededPositions.includes(player.position)) {
    emit('pick', player.id, player.name, player.position)
  }
}
</script>
