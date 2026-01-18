<template>
  <div class="grid lg:grid-cols-3 gap-6">
    <!-- Left: Team Draw & Player Selection -->
    <div class="lg:col-span-2 space-y-6">
      <!-- Progress -->
      <div class="card p-4">
        <div class="flex items-center justify-between mb-2">
          <span class="text-sm font-medium text-gray-700">
            Pick {{ draftStore.currentPick }} of {{ draftStore.totalPicks }}
          </span>
          <span class="text-sm text-gray-500">
            {{ Math.round(draftStore.progress) }}% complete
          </span>
        </div>
        <div class="w-full bg-gray-200 rounded-full h-2">
          <div
            class="bg-primary-600 h-2 rounded-full transition-all duration-500"
            :style="{ width: `${draftStore.progress}%` }"
          />
        </div>
      </div>

      <!-- Team Draw -->
      <DraftTeamDraw
        v-if="draftStore.currentTeam && !draftStore.isComplete"
        :team="draftStore.currentTeam"
        :available-players="draftStore.availablePlayers"
        :needed-positions="draftStore.neededPositions"
        @pick="handlePick"
      />

      <!-- Waiting State -->
      <div v-else-if="!draftStore.isComplete && !draftStore.currentTeam" class="card p-8 text-center">
        <CommonLoadingSpinner size="lg" text="Drawing team..." />
      </div>

      <!-- Draft Complete -->
      <DraftComplete
        v-if="draftStore.isComplete"
        :roster="draftStore.roster"
        :total-points="0"
        @close="$emit('complete')"
      />
    </div>

    <!-- Right: Roster -->
    <div>
      <DraftRosterSlots :picks="draftStore.roster" />
    </div>
  </div>
</template>

<script setup lang="ts">
const draftStore = useDraftStore()

const emit = defineEmits<{
  pick: [playerId: string, playerName: string, position: string]
  complete: []
}>()

const handlePick = (playerId: string, playerName: string, position: string) => {
  emit('pick', playerId, playerName, position)
}
</script>
