<template>
  <div class="divide-y divide-gray-200">
    <div
      v-for="member in members"
      :key="member.id"
      class="px-6 py-4 flex items-center justify-between"
    >
      <div class="flex items-center">
        <div class="flex-shrink-0 h-10 w-10">
          <img
            v-if="member.avatar_url"
            :src="member.avatar_url"
            :alt="member.display_name"
            class="h-10 w-10 rounded-full"
          />
          <div v-else class="h-10 w-10 rounded-full bg-primary-100 flex items-center justify-center">
            <span class="text-primary-600 font-medium">
              {{ member.display_name?.charAt(0) || '?' }}
            </span>
          </div>
        </div>
        <div class="ml-4">
          <div class="text-sm font-medium text-gray-900">{{ member.team_name }}</div>
          <div class="text-sm text-gray-500">{{ member.display_name }}</div>
        </div>
      </div>
      <div class="text-sm text-gray-500">
        Joined {{ formatDate(member.joined_at) }}
      </div>
    </div>

    <div v-if="members.length === 0" class="p-8 text-center text-gray-500">
      No members yet
    </div>
  </div>
</template>

<script setup lang="ts">
import type { LeagueMember } from '~/types'

defineProps<{
  members: LeagueMember[]
}>()

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}
</script>
