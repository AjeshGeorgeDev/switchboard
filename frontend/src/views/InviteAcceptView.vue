<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../api'
import { useAuthStore } from '../stores/auth'
import AuthShell from '../components/AuthShell.vue'
import Password from 'primevue/password'
import Button from 'primevue/button'
import Message from 'primevue/message'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const token = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const error = ref('')
const loading = ref(false)
const ready = ref(false)

onMounted(async () => {
  token.value = String(route.query.token || '')
  if (!token.value) {
    error.value = 'Invitation link is missing or invalid.'
    return
  }
  try {
    const data = await api.get<{ email: string }>(`/api/auth/invite?token=${encodeURIComponent(token.value)}`)
    email.value = data.email
    ready.value = true
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Invitation is invalid or expired.'
  }
})

async function submit() {
  error.value = ''
  if (password.value !== confirmPassword.value) {
    error.value = 'Passwords do not match.'
    return
  }
  if (password.value.length < 8) {
    error.value = 'Password must be at least 8 characters.'
    return
  }
  loading.value = true
  try {
    await api.post('/api/auth/invite/accept', {
      token: token.value,
      password: password.value,
    })
    await auth.fetchMe()
    router.push('/launcher')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Could not accept invitation.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <AuthShell>
    <div class="invite-form">
      <h2>Accept invitation</h2>
      <p class="subtitle">Set a password to join Switchboard.</p>

      <Message v-if="error" severity="error" class="mb-3">{{ error }}</Message>

      <form v-if="ready" @submit.prevent="submit">
        <div class="readonly mb-3">
          <span>Email</span>
          <strong>{{ email }}</strong>
        </div>

        <div class="field">
          <label for="password">Password</label>
          <Password id="password" v-model="password" toggle-mask class="w-full" input-class="w-full" />
        </div>
        <div class="field">
          <label for="confirm">Confirm password</label>
          <Password id="confirm" v-model="confirmPassword" toggle-mask class="w-full" input-class="w-full" />
        </div>
        <Button type="submit" label="Create account" class="w-full" :loading="loading" />
      </form>
    </div>
  </AuthShell>
</template>

<style scoped>
.invite-form h2 {
  font-size: 1.5rem;
  margin-bottom: 0.35rem;
}

.subtitle {
  color: var(--sb-muted);
  margin-bottom: 1.25rem;
}

.readonly {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.readonly span {
  font-size: 0.8rem;
  color: var(--sb-muted);
  font-weight: 600;
}

.mb-3 { margin-bottom: 1rem; }
</style>
