<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <!-- Header -->
    <div class="mb-6 flex items-center justify-between">
      <div>
        <NuxtLink :to="`/leagues/${leagueId}`" class="text-primary-600 hover:text-primary-700 text-sm mb-2 inline-flex items-center">
          <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
          Back to League
        </NuxtLink>
        <h1 class="text-2xl font-bold text-gray-900">
          Week {{ week }} Draft
        </h1>
      </div>
    </div>

    <!-- Error Alert -->
    <CommonAlert
      v-if="error"
      type="error"
      :message="error"
      class="mb-6"
      dismissible
      @dismiss="error = ''"
    />

    <!-- Loading State -->
    <div v-if="loading" class="flex justify-center py-20">
      <CommonLoadingSpinner size="lg" text="Loading draft..." />
    </div>

    <!-- Draft Not Started -->
    <div v-else-if="!draftStarted && !draftStore.isComplete" class="card p-8 text-center">
      <h2 class="text-xl font-semibold text-gray-900 mb-4">Ready to Draft?</h2>
      <p class="text-gray-600 mb-6">
        You'll draft 7 players for your Week {{ week }} roster. Each pick, you'll get a random NFL team to choose from.
      </p>
      <button @click="startDraft" :disabled="starting" class="btn-primary">
        <CommonLoadingSpinner v-if="starting" size="sm" />
        <span v-else>Start Draft</span>
      </button>
    </div>

    <!-- Draft Board -->
    <DraftDraftBoard
      v-else
      @pick="handlePick"
      @complete="handleComplete"
    />
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: 'auth',
})

const route = useRoute()
const leagueStore = useLeagueStore()
const draftStore = useDraftStore()

const leagueId = route.params.id as string
const loading = ref(true)
const starting = ref(false)
const error = ref('')
const draftStarted = ref(false)

// Get week from season
const week = computed(() => leagueStore.currentSeason?.current_week || 1)

const initializeDraft = async () => {
  loading.value = true
  error.value = ''

  try {
    // Load league and season info
    await leagueStore.fetchLeague(leagueId)
    await leagueStore.fetchActiveSeason(leagueId)

    if (!leagueStore.currentSeason) {
      error.value = 'No active season found'
      return
    }

    // Check existing draft status
    const draft = await draftStore.fetchDraft(
      leagueId,
      leagueStore.currentSeason.id,
      week.value
    )

    if (draft.status === 'in_progress') {
      draftStarted.value = true
      // Load full draft state including players
      await draftStore.loadDraftState(draft.id)
    } else if (draft.status === 'complete') {
      draftStarted.value = true
    }
  } catch (err: any) {
    // Draft doesn't exist yet, which is fine
    if (err.statusCode !== 404) {
      error.value = err.data?.message || 'Failed to load draft'
    }
  } finally {
    loading.value = false
  }
}

const startDraft = async () => {
  starting.value = true
  error.value = ''

  try {
    await draftStore.startDraft(
      leagueId,
      leagueStore.currentSeason!.id,
      week.value
    )

    draftStarted.value = true
  } catch (err: any) {
    error.value = err.data?.message || 'Failed to start draft'
  } finally {
    starting.value = false
  }
}

const handlePick = async (playerId: string, playerName: string, position: string) => {
  if (!draftStore.draftState?.id) return

  draftStore.clearError()

  try {
    await draftStore.makePick(draftStore.draftState.id, playerId, playerName, position)
  } catch (err: any) {
    error.value = err.data?.message || 'Failed to make pick'
  }
}

const handleComplete = () => {
  navigateTo(`/leagues/${leagueId}`)
}

onMounted(initializeDraft)

onUnmounted(() => {
  draftStore.reset()
})

// Watch for errors from draft store
watch(() => draftStore.error, (newError) => {
  if (newError) {
    error.value = newError
  }
})
</script>
