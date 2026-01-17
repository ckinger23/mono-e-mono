<template>
  <div class="bg-white">
    <!-- Hero section -->
    <div class="relative overflow-hidden">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24 sm:py-32">
        <div class="text-center">
          <h1 class="text-4xl sm:text-6xl font-extrabold text-gray-900 tracking-tight">
            <span class="block">Weekly Draft</span>
            <span class="block text-primary-600">Fantasy Football</span>
          </h1>
          <p class="mt-6 max-w-2xl mx-auto text-xl text-gray-500">
            Draft a fresh roster every week. No trades, no waivers - just pure drafting skill.
            Compete for the highest weekly score and become the season champion.
          </p>
          <div class="mt-10 flex justify-center gap-4">
            <NuxtLink to="/register" class="btn-primary text-lg px-8 py-3">
              Get Started
            </NuxtLink>
            <NuxtLink to="/login" class="btn-secondary text-lg px-8 py-3">
              Sign In
            </NuxtLink>
          </div>
        </div>
      </div>
    </div>

    <!-- Features section -->
    <div class="bg-gray-50 py-24">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="text-center mb-16">
          <h2 class="text-3xl font-bold text-gray-900">How It Works</h2>
          <p class="mt-4 text-lg text-gray-600">A fresh take on fantasy football</p>
        </div>

        <div class="grid md:grid-cols-3 gap-8">
          <div class="bg-white p-8 rounded-xl shadow-sm">
            <div class="w-12 h-12 bg-primary-100 rounded-lg flex items-center justify-center mb-4">
              <svg class="w-6 h-6 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
              </svg>
            </div>
            <h3 class="text-xl font-semibold text-gray-900 mb-2">Create or Join a League</h3>
            <p class="text-gray-600">
              Start a league with friends or join an existing one using an invite code.
            </p>
          </div>

          <div class="bg-white p-8 rounded-xl shadow-sm">
            <div class="w-12 h-12 bg-primary-100 rounded-lg flex items-center justify-center mb-4">
              <svg class="w-6 h-6 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
              </svg>
            </div>
            <h3 class="text-xl font-semibold text-gray-900 mb-2">Draft Every Week</h3>
            <p class="text-gray-600">
              Each week, draft a fresh 7-player roster. Random team draws keep it exciting!
            </p>
          </div>

          <div class="bg-white p-8 rounded-xl shadow-sm">
            <div class="w-12 h-12 bg-primary-100 rounded-lg flex items-center justify-center mb-4">
              <svg class="w-6 h-6 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
            </div>
            <h3 class="text-xl font-semibold text-gray-900 mb-2">Climb the Standings</h3>
            <p class="text-gray-600">
              Accumulate points all season. The highest total wins the championship!
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- CTA section -->
    <div class="bg-primary-600 py-16">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
        <h2 class="text-3xl font-bold text-white mb-4">Ready to Draft?</h2>
        <p class="text-primary-100 text-lg mb-8">
          Join thousands of players in the ultimate weekly fantasy experience.
        </p>
        <NuxtLink to="/register" class="bg-white text-primary-600 px-8 py-3 rounded-md font-semibold hover:bg-primary-50 transition-colors">
          Create Free Account
        </NuxtLink>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'default',
})
</script>
