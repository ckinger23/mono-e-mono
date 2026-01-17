<template>
  <div class="overflow-x-auto">
    <table v-if="standings.length > 0" class="min-w-full divide-y divide-gray-200">
      <thead class="bg-gray-50">
        <tr>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
            Rank
          </th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
            Team
          </th>
          <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
            Points
          </th>
          <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
            Wins
          </th>
          <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
            Best Week
          </th>
          <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
            Weeks
          </th>
        </tr>
      </thead>
      <tbody class="bg-white divide-y divide-gray-200">
        <tr v-for="standing in standings" :key="standing.id" class="hover:bg-gray-50">
          <td class="px-6 py-4 whitespace-nowrap">
            <span
              :class="[
                'inline-flex items-center justify-center w-8 h-8 rounded-full text-sm font-bold',
                standing.current_rank === 1 ? 'bg-yellow-100 text-yellow-800' :
                standing.current_rank === 2 ? 'bg-gray-100 text-gray-800' :
                standing.current_rank === 3 ? 'bg-orange-100 text-orange-800' :
                'bg-gray-50 text-gray-600'
              ]"
            >
              {{ standing.current_rank || '-' }}
            </span>
          </td>
          <td class="px-6 py-4 whitespace-nowrap">
            <div class="flex items-center">
              <div class="flex-shrink-0 h-10 w-10">
                <div class="h-10 w-10 rounded-full bg-primary-100 flex items-center justify-center">
                  <span class="text-primary-600 font-medium">
                    {{ standing.team_name?.charAt(0) || '?' }}
                  </span>
                </div>
              </div>
              <div class="ml-4">
                <div class="text-sm font-medium text-gray-900">{{ standing.team_name }}</div>
                <div class="text-sm text-gray-500">{{ standing.display_name }}</div>
              </div>
            </div>
          </td>
          <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-semibold text-gray-900">
            {{ standing.total_points.toFixed(2) }}
          </td>
          <td class="px-6 py-4 whitespace-nowrap text-right text-sm text-gray-500">
            {{ standing.weekly_wins }}
          </td>
          <td class="px-6 py-4 whitespace-nowrap text-right text-sm text-gray-500">
            {{ standing.best_week.toFixed(2) }}
          </td>
          <td class="px-6 py-4 whitespace-nowrap text-right text-sm text-gray-500">
            {{ standing.weeks_played }}
          </td>
        </tr>
      </tbody>
    </table>

    <div v-else class="p-8 text-center text-gray-500">
      No standings yet. Start a season to begin tracking!
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Standing } from '~/types'

defineProps<{
  standings: Standing[]
}>()
</script>
