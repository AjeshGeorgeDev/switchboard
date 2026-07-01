<script setup lang="ts">
defineProps<{
  size?: 'sm' | 'md' | 'lg'
  variant?: 'gradient' | 'glass'
}>()
</script>

<template>
  <span
    class="switchboard-logo"
    :class="[`size-${size ?? 'sm'}`, `variant-${variant ?? 'gradient'}`]"
    aria-hidden="true"
  >
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path
        d="M12 5.5L6.5 15.5"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
      />
      <path
        d="M12 5.5L17.5 15.5"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
      />
      <path
        d="M7.5 15.5H16.5"
        stroke="currentColor"
        stroke-width="1.75"
        stroke-linecap="round"
        opacity="0.55"
      />
      <circle cx="12" cy="5.5" r="2.75" fill="currentColor" />
      <circle cx="6.5" cy="15.5" r="2.25" fill="currentColor" />
      <circle cx="17.5" cy="15.5" r="2.25" fill="currentColor" />
    </svg>
  </span>
</template>

<style scoped>
.switchboard-logo {
  display: grid;
  place-items: center;
  flex-shrink: 0;
  color: white;
}

.switchboard-logo svg {
  width: 62%;
  height: 62%;
}

.size-sm {
  width: 2rem;
  height: 2rem;
  border-radius: 0.65rem;
}

.size-md {
  width: 2.5rem;
  height: 2.5rem;
  border-radius: 0.75rem;
}

.size-lg {
  width: 3.25rem;
  height: 3.25rem;
  border-radius: 1rem;
}

.variant-gradient {
  background: linear-gradient(135deg, #4f46e5, #7c3aed);
}

.variant-glass {
  background: rgba(255, 255, 255, 0.12);
  border: 1px solid rgba(255, 255, 255, 0.16);
}
</style>
