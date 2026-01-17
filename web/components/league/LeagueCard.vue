<template>
  <NuxtLink :to="`/leagues/${league.id}`" class="block py-4 hover:bg-gray-50 -mx-4 px-4 transition-colors">
    <div class="flex items-center justify-between">
      <div>
        <h3 class="text-lg font-medium text-gray-900">{{ league.name }}</h3>
        <div class="mt-1 flex items-center space-x-4 text-sm text-gray-500">
          <span>{{ league.member_count }} / {{ league.max_members }} members</span>
          <span class="capitalize">{{ league.scoring_type }}</span>
        </div>
      </div>
      <div class="flex items-center">
        <span class="text-sm text-gray-500 mr-4">
          Code: <span class="font-mono">{{ league.invite_code }}</span>
        </span>
        <svg class="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
      </div>
    </div>
  </NuxtLink>
</template>

<script setup lang="ts">
import type { League } from '~/types'

defineProps<{
  league: League
}>()
</script>
