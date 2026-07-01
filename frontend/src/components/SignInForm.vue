<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { api } from '../api'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Button from 'primevue/button'
import Divider from 'primevue/divider'
import Checkbox from 'primevue/checkbox'

const props = withDefaults(defineProps<{
  idPrefix?: string
}>(), {
  idPrefix: 'signin',
})

const emit = defineEmits<{
  success: []
}>()

const auth = useAuthStore()
const email = ref('')
const password = ref('')
const rememberMe = ref(localStorage.getItem('remember_me') !== 'false')
const providers = ref<{ name: string; display_name: string }[]>([])
const error = ref('')
const submitting = ref(false)

onMounted(async () => {
  try {
    providers.value = await api.get('/api/auth/providers')
  } catch {
    providers.value = []
  }
})

async function submit() {
  error.value = ''
  submitting.value = true
  try {
    localStorage.setItem('remember_me', rememberMe.value ? 'true' : 'false')
    await auth.login(email.value, password.value, rememberMe.value)
    emit('success')
  } catch {
    error.value = 'Invalid credentials'
  } finally {
    submitting.value = false
  }
}

function oidcLogin(name: string) {
  window.location.href = `/api/auth/oidc/${name}/login`
}
</script>

<template>
  <div>
    <div v-if="providers.length" class="mb-4 grid gap-3">
      <Button
        v-for="p in providers"
        :key="p.name"
        :label="p.display_name"
        severity="secondary"
        outlined
        class="w-full"
        @click="oidcLogin(p.name)"
      />
      <Divider align="center"><span class="text-sm text-muted-color">or continue with password</span></Divider>
    </div>

    <form class="flex flex-col gap-4" @submit.prevent="submit">
      <div class="field">
        <label :for="`${idPrefix}-email`">Email</label>
        <InputText
          :id="`${idPrefix}-email`"
          v-model="email"
          type="email"
          class="w-full"
          autocomplete="email"
        />
      </div>
      <div class="field">
        <label :for="`${idPrefix}-password`">Password</label>
        <Password
          :id="`${idPrefix}-password`"
          v-model="password"
          :feedback="false"
          toggle-mask
          class="w-full"
          input-class="w-full"
          autocomplete="current-password"
          @keyup.enter="submit"
        />
      </div>

      <div class="field-row remember">
        <Checkbox v-model="rememberMe" :input-id="`${idPrefix}-remember`" binary />
        <label :for="`${idPrefix}-remember`">Keep me signed in</label>
      </div>

      <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
      <Button type="submit" label="Sign in" class="w-full" :loading="submitting" />
    </form>
  </div>
</template>

<style scoped>
.remember label {
  font-weight: 500;
  color: var(--sb-muted);
}
</style>
