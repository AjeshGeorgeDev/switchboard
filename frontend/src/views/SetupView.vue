<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api'
import { useAuthStore } from '../stores/auth'
import { useSetupStore } from '../stores/setup'
import AuthShell from '../components/AuthShell.vue'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Button from 'primevue/button'

const router = useRouter()
const auth = useAuthStore()
const setup = useSetupStore()

const email = ref('')
const displayName = ref('')
const password = ref('')
const confirmPassword = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  if (password.value !== confirmPassword.value) {
    error.value = 'Passwords do not match'
    return
  }
  if (password.value.length < 8) {
    error.value = 'Password must be at least 8 characters'
    return
  }

  loading.value = true
  try {
    await api.post('/api/setup', {
      email: email.value,
      display_name: displayName.value,
      password: password.value,
    })
    setup.markComplete()
    await auth.fetchMe()
    router.push('/')
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : 'Setup failed'
    if (msg.includes('setup already complete')) {
      setup.markComplete()
      router.push({ path: '/', query: { signin: '1' } })
      return
    }
    error.value = msg.replace(/^\{.*"error":"([^"]+)".*\}$/, '$1') || 'Setup failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <AuthShell>
    <div class="auth-form">
      <h2>Welcome aboard</h2>
      <p class="subtitle">Create the first administrator account for this instance.</p>

      <form class="setup-form" @submit.prevent="submit">
        <div class="field">
          <label for="setup-email">Email</label>
          <InputText id="setup-email" v-model="email" type="email" class="w-full" autocomplete="email" />
        </div>
      <div class="field">
        <label>Display name</label>
        <InputText v-model="displayName" class="w-full" placeholder="Administrator" />
      </div>
      <div class="field">
        <label>Password</label>
        <Password v-model="password" :feedback="true" toggle-mask class="w-full" input-class="w-full" autocomplete="new-password" />
      </div>
      <div class="field">
        <label>Confirm password</label>
        <Password v-model="confirmPassword" :feedback="false" toggle-mask class="w-full" input-class="w-full" autocomplete="new-password" />
      </div>

        <p v-if="error" class="error">{{ error }}</p>
        <Button type="submit" label="Create administrator account" class="w-full" :loading="loading" />
      </form>
    </div>
  </AuthShell>
</template>

<style scoped>
.auth-form h2 {
  font-size: 1.5rem;
  font-weight: 800;
  margin-bottom: 0.35rem;
}

.subtitle {
  color: var(--sb-muted);
  margin-bottom: 1.5rem;
}

.setup-form {
  display: flex;
  flex-direction: column;
}

.error {
  color: #dc2626;
  margin-bottom: 1rem;
  font-size: 0.875rem;
}
</style>
