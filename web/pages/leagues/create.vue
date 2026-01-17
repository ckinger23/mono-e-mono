<template>
  <div class="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div class="mb-8">
      <h1 class="text-3xl font-bold text-gray-900">Create League</h1>
      <p class="mt-2 text-gray-600">Start a new league and invite your friends</p>
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
            <label for="name" class="label">League Name</label>
            <input
              id="name"
              v-model="name"
              type="text"
              required
              class="input"
              placeholder="e.g., Sunday Champions"
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

          <div>
            <label for="maxMembers" class="label">Max Members</label>
            <select id="maxMembers" v-model="maxMembers" class="input">
              <option :value="4">4 members</option>
              <option :value="6">6 members</option>
              <option :value="8">8 members</option>
              <option :value="10">10 members</option>
              <option :value="12">12 members</option>
            </select>
          </div>

          <div>
            <label for="scoringType" class="label">Scoring Type</label>
            <select id="scoringType" v-model="scoringType" class="input">
              <option value="ppr">PPR (1 point per reception)</option>
              <option value="half_ppr">Half PPR (0.5 points per reception)</option>
              <option value="standard">Standard (no reception points)</option>
            </select>
          </div>

          <div class="flex space-x-4">
            <button
              type="submit"
              :disabled="loading"
              class="btn-primary flex-1"
            >
              <CommonLoadingSpinner v-if="loading" size="sm" />
              <span v-else>Create League</span>
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

const name = ref('')
const teamName = ref('')
const maxMembers = ref(10)
const scoringType = ref('ppr')
const loading = ref(false)
const error = ref('')

const handleSubmit = async () => {
  loading.value = true
  error.value = ''

  try {
    const league = await leagueStore.createLeague(
      name.value,
      teamName.value,
      maxMembers.value,
      scoringType.value
    )
    navigateTo(`/leagues/${league.id}`)
  } catch (err: any) {
    error.value = err.data?.message || 'Failed to create league'
  } finally {
    loading.value = false
  }
}
</script>
