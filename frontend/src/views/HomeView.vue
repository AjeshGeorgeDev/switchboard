<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { api } from '../api'
import SwitchboardLogo from '../components/SwitchboardLogo.vue'
import CatalogGrid from '../components/CatalogGrid.vue'
import SignInForm from '../components/SignInForm.vue'
import { type CatalogSection } from '../utils/catalog'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const apps = ref<any[]>([])
const sections = ref<CatalogSection[]>([])
const loading = ref(true)
const signInOpen = ref(false)

function redirectAfterAuth() {
  const target = typeof route.query.redirect === 'string' ? route.query.redirect : '/launcher'
  router.replace(target)
}

function clearSignInQuery() {
  if (route.query.signin !== '1') return
  const { signin, redirect, ...rest } = route.query
  router.replace({ query: rest })
}

function openSignIn() {
  signInOpen.value = true
  if (route.query.signin !== '1') {
    router.replace({ query: { ...route.query, signin: '1' } })
  }
}

function onSignInVisible(visible: boolean) {
  signInOpen.value = visible
  if (!visible) clearSignInQuery()
}

function onSignInSuccess() {
  signInOpen.value = false
  clearSignInQuery()
  redirectAfterAuth()
}

onMounted(async () => {
  try {
    const [appList, sectionList] = await Promise.all([
      api.get<any[]>('/api/catalog/public'),
      api.get<CatalogSection[]>('/api/catalog/sections'),
    ])
    apps.value = appList
    sections.value = sectionList
  } catch {
    apps.value = []
    sections.value = []
  } finally {
    loading.value = false
  }

  if (route.query.signin === '1') {
    if (auth.user) redirectAfterAuth()
    else signInOpen.value = true
  }
})

watch(
  () => route.query.signin,
  (signin) => {
    if (signin === '1' && !auth.user) signInOpen.value = true
    else if (signin !== '1') signInOpen.value = false
  },
)

watch(
  () => auth.user,
  (user) => {
    if (user && route.query.signin === '1') redirectAfterAuth()
  },
)
</script>

<template>
  <div class="relative flex min-h-screen flex-col bg-surface-50">
    <div class="pointer-events-none absolute inset-0 bg-[radial-gradient(ellipse_80%_50%_at_50%_-20%,color-mix(in_srgb,var(--p-primary-color)_16%,transparent),transparent)]" />

    <header class="relative z-10 border-b border-surface bg-surface-0/70 backdrop-blur-xl">
      <div class="flex h-16 w-full items-center justify-between gap-4 px-5 sm:px-8 lg:px-10">
        <button
          type="button"
          class="flex items-center gap-3 border-0 bg-transparent p-0 text-color"
          @click="router.push('/')"
        >
          <SwitchboardLogo />
          <span class="text-base font-bold tracking-tight sm:text-lg">Switchboard</span>
        </button>
        <Button
          v-if="auth.user"
          label="My Apps"
          icon="pi pi-th-large"
          severity="secondary"
          outlined
          @click="router.push('/launcher')"
        />
        <Button
          v-else
          label="Sign in"
          icon="pi pi-sign-in"
          severity="secondary"
          outlined
          @click="openSignIn"
        />
      </div>
    </header>

    <main class="relative z-10 w-full flex-1 px-5 py-5 sm:px-8 sm:py-6 lg:px-10">
      <div class="mb-5">
        <h1 class="text-3xl font-bold tracking-tight text-color sm:text-4xl">Applications</h1>
      </div>

      <div
        v-if="loading"
        class="flex items-center justify-center gap-3 py-32 text-muted-color"
      >
        <i class="pi pi-spin pi-spinner text-primary" />
        <span>Loading applications…</span>
      </div>
      <CatalogGrid
        v-else
        variant="public"
        :apps="apps"
        :sections="sections"
        empty-message="No public applications are available yet."
      />
    </main>

    <footer class="relative z-10 border-t border-surface bg-surface-0/60 backdrop-blur-sm">
      <div class="flex w-full items-center justify-between gap-4 px-5 py-4 text-xs text-muted-color sm:px-8 lg:px-10">
        <span>&copy; {{ new Date().getFullYear() }}</span>
        <span>MIT License · Switchboard</span>
      </div>
    </footer>

    <Dialog
      :visible="signInOpen"
      modal
      dismissable-mask
      header="Sign in"
      class="w-full max-w-md"
      :draggable="false"
      @update:visible="onSignInVisible"
    >
      <p class="mb-5 text-sm text-muted-color">Access your internal workspace and role-based apps.</p>
      <SignInForm @success="onSignInSuccess" />
    </Dialog>
  </div>
</template>
