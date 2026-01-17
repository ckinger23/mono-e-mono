<template>
  <div
    class="rounded-md p-4"
    :class="alertClasses"
    role="alert"
  >
    <div class="flex">
      <div class="flex-shrink-0">
        <component :is="icon" class="h-5 w-5" :class="iconClass" />
      </div>
      <div class="ml-3">
        <h3 v-if="title" class="text-sm font-medium" :class="titleClass">
          {{ title }}
        </h3>
        <div class="text-sm" :class="textClass">
          <slot>{{ message }}</slot>
        </div>
      </div>
      <div v-if="dismissible" class="ml-auto pl-3">
        <button
          type="button"
          class="inline-flex rounded-md p-1.5 focus:outline-none focus:ring-2 focus:ring-offset-2"
          :class="dismissClass"
          @click="$emit('dismiss')"
        >
          <span class="sr-only">Dismiss</span>
          <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  type?: 'info' | 'success' | 'warning' | 'error'
  title?: string
  message?: string
  dismissible?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  type: 'info',
  dismissible: false,
})

defineEmits<{
  dismiss: []
}>()

const alertClasses = computed(() => {
  switch (props.type) {
    case 'success':
      return 'bg-green-50'
    case 'warning':
      return 'bg-yellow-50'
    case 'error':
      return 'bg-red-50'
    default:
      return 'bg-blue-50'
  }
})

const iconClass = computed(() => {
  switch (props.type) {
    case 'success':
      return 'text-green-400'
    case 'warning':
      return 'text-yellow-400'
    case 'error':
      return 'text-red-400'
    default:
      return 'text-blue-400'
  }
})

const titleClass = computed(() => {
  switch (props.type) {
    case 'success':
      return 'text-green-800'
    case 'warning':
      return 'text-yellow-800'
    case 'error':
      return 'text-red-800'
    default:
      return 'text-blue-800'
  }
})

const textClass = computed(() => {
  switch (props.type) {
    case 'success':
      return 'text-green-700'
    case 'warning':
      return 'text-yellow-700'
    case 'error':
      return 'text-red-700'
    default:
      return 'text-blue-700'
  }
})

const dismissClass = computed(() => {
  switch (props.type) {
    case 'success':
      return 'bg-green-50 text-green-500 hover:bg-green-100 focus:ring-green-600'
    case 'warning':
      return 'bg-yellow-50 text-yellow-500 hover:bg-yellow-100 focus:ring-yellow-600'
    case 'error':
      return 'bg-red-50 text-red-500 hover:bg-red-100 focus:ring-red-600'
    default:
      return 'bg-blue-50 text-blue-500 hover:bg-blue-100 focus:ring-blue-600'
  }
})

// Simple icon components
const icon = computed(() => {
  return {
    template: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
      ${props.type === 'success' ? '<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />' : ''}
      ${props.type === 'warning' ? '<path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />' : ''}
      ${props.type === 'error' ? '<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />' : ''}
      ${props.type === 'info' ? '<path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />' : ''}
    </svg>`
  }
})
</script>
