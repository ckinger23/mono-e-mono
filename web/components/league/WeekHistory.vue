<template>
  <div>
    <!-- Week Selector -->
    <div class="px-6 py-4 border-b border-gray-200">
      <div class="flex items-center space-x-4">
        <label class="text-sm font-medium text-gray-700">Week:</label>
        <select v-model="selectedWeek" class="input w-32">
          <option v-for="week in availableWeeks" :key="week" :value="week">
            Week {{ week }}
          </option>
        </select>
      </div>
    </div>

    <!-- Results -->
    <div v-if="loading" class="p-8">
      <CommonLoadingSpinner />
    </div>

    <div v-else-if="results.length > 0" class="divide-y divide-gray-200">
      <div
        v-for="(result, index) in results"
        :key="result.id"
        class="px-6 py-4 flex items-center justify-between"
      >
        <div class="flex items-center">
          <span
            :class="[
              'inline-flex items-center justify-center w-8 h-8 rounded-full text-sm font-bold mr-4',
              index === 0 ? 'bg-yellow-100 text-yellow-800' : 'bg-gray-50 text-gray-600'
            ]"
          >
            {{ index + 1 }}
          </span>
          <div>
            <div class="text-sm font-medium text-gray-900">{{ result.team_name }}</div>
            <div class="text-sm text-gray-500">{{ result.display_name }}</div>
          </div>
        </div>
        <div class="text-right">
          <div class="text-lg font-semibold text-gray-900">
            {{ result.total_points.toFixed(2) }}
          </div>
          <div v-if="result.is_weekly_winner" class="text-xs text-yellow-600 font-medium">
            Weekly Winner
          </div>
        </div>
      </div>
    </div>

    <div v-else class="p-8 text-center text-gray-500">
      No results for this week yet
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Season, WeeklyResult } from '~/types'

const props = defineProps<{
  leagueId: string
  season: Season
}>()

const leagueStore = useLeagueStore()

const selectedWeek = ref(props.season.current_week)
const loading = ref(false)
const results = ref<WeeklyResult[]>([])

const availableWeeks = computed(() => {
  return Array.from({ length: props.season.current_week }, (_, i) => i + 1)
})

const loadResults = async () => {
  loading.value = true
  try {
    results.value = await leagueStore.fetchWeeklyResults(
      props.leagueId,
      props.season.id,
      selectedWeek.value
    )
  } finally {
    loading.value = false
  }
}

watch(selectedWeek, loadResults)
onMounted(loadResults)
</script>
