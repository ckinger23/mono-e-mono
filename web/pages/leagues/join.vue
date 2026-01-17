<template>
  <div class="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div class="mb-8">
      <h1 class="text-3xl font-bold text-gray-900">Join League</h1>
      <p class="mt-2 text-gray-600">Enter an invite code to join an existing league</p>
    </div>

    <div class="card">
      <div class="card-body">
        <form @submit.prevent="handleSubmit" class="space-y-6">
          <CommonAlert
            v-if="error"
            type="error"
            :message="error"
            dismissible
            @dismiss="error = ''"
          />

          <div>
            <label for="inviteCode" class="label">Invite Code</label>
            <input
              id="inviteCode"
              v-model="inviteCode"
              type="text"
              required
              class="input font-mono uppercase"
              placeholder="e.g., ABC12345"
              maxlength="8"
            />
          </div>

          <div>
            <label for="teamName" class="label">Your Team Name</label>
            <input
              id="teamName"
              v-model="teamName"
              type="text"
              required
              class="input"
              placeholder="e.g., The Underdogs"
            />
          </div>

          <div class="flex space-x-4">
            <button
              type="submit"
              :disabled="loading"
              class="btn-primary flex-1"
            >
              <CommonLoadingSpinner v-if="loading" size="sm" />
              <span v-else>Join League</span>
            </button>
            <NuxtLink to="/dashboard" class="btn-secondary">
              Cancel
            </NuxtLink>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: 'auth',
})

const leagueStore = useLeagueStore()

const inviteCode = ref('')
const teamName = ref('')
const loading = ref(false)
const error = ref('')

const handleSubmit = async () => {
  loading.value = true
  error.value = ''

  try {
    const member = await leagueStore.joinLeague(inviteCode.value, teamName.value)
    navigateTo(`/leagues/${member.league_id}`)
  } catch (err: any) {
    error.value = err.data?.message || 'Failed to join league'
  } finally {
    loading.value = false
  }
}
</script>
