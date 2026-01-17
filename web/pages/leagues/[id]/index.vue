<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <CommonLoadingSpinner v-if="loading" size="lg" />

    <template v-else-if="league">
      <!-- Header -->
      <div class="mb-8 flex items-start justify-between">
        <div>
          <h1 class="text-3xl font-bold text-gray-900">{{ league.name }}</h1>
          <p class="mt-2 text-gray-600">
            {{ league.member_count }} / {{ league.max_members }} members
            <span class="mx-2">·</span>
            <span class="capitalize">{{ league.scoring_type }}</span>
            <span class="mx-2">·</span>
            Invite: <span class="font-mono">{{ league.invite_code }}</span>
          </p>
        </div>
        <div class="flex space-x-3">
          <button
            v-if="!currentSeason"
            @click="createSeason"
            :disabled="creatingseason"
            class="btn-primary"
          >
            Start Season
          </button>
          <NuxtLink
            v-if="currentSeason"
            :to="`/leagues/${league.id}/draft`"
            class="btn-primary"
          >
            Draft Week {{ currentSeason.current_week }}
          </NuxtLink>
        </div>
      </div>

      <!-- Tabs -->
      <div class="border-b border-gray-200 mb-6">
        <nav class="flex space-x-8">
          <button
            @click="activeTab = 'standings'"
            :class="[
              activeTab === 'standings'
                ? 'border-primary-500 text-primary-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300',
              'whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm'
            ]"
          >
            Standings
          </button>
          <button
            @click="activeTab = 'members'"
            :class="[
              activeTab === 'members'
                ? 'border-primary-500 text-primary-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300',
              'whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm'
            ]"
          >
            Members
          </button>
          <button
            @click="activeTab = 'history'"
            :class="[
              activeTab === 'history'
                ? 'border-primary-500 text-primary-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300',
              'whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm'
            ]"
          >
            History
          </button>
        </nav>
      </div>

      <!-- Tab Content -->
      <div v-if="activeTab === 'standings'" class="card">
        <LeagueStandingsTable :standings="standings" />
      </div>

      <div v-else-if="activeTab === 'members'" class="card">
        <LeagueMembersList :members="members" />
      </div>

      <div v-else-if="activeTab === 'history'" class="card">
        <LeagueWeekHistory
          v-if="currentSeason"
          :league-id="league.id"
          :season="currentSeason"
        />
        <div v-else class="p-8 text-center text-gray-500">
          No season started yet
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: 'auth',
})

const route = useRoute()
const leagueStore = useLeagueStore()

const leagueId = route.params.id as string
const loading = ref(true)
const creatingseason = ref(false)
const activeTab = ref('standings')

const league = computed(() => leagueStore.currentLeague)
const members = computed(() => leagueStore.members)
const currentSeason = computed(() => leagueStore.currentSeason)
const standings = computed(() => leagueStore.standings)

const loadData = async () => {
  loading.value = true
  try {
    await leagueStore.fetchLeague(leagueId)
    await Promise.all([
      leagueStore.fetchMembers(leagueId),
      leagueStore.fetchActiveSeason(leagueId),
    ])

    if (currentSeason.value) {
      await leagueStore.fetchStandings(leagueId, currentSeason.value.id)
    }
  } finally {
    loading.value = false
  }
}

const createSeason = async () => {
  creatingseason.value = true
  try {
    await leagueStore.createSeason(leagueId)
    await leagueStore.fetchStandings(leagueId, currentSeason.value!.id)
  } finally {
    creatingseason.value = false
  }
}

onMounted(loadData)

onUnmounted(() => {
  leagueStore.clearCurrentLeague()
})
</script>
