<script setup lang="ts">
import SwitchboardLogo from './SwitchboardLogo.vue'
</script>

<template>
  <div class="auth-shell">
    <div class="auth-brand">
      <SwitchboardLogo size="lg" variant="glass" />
      <h1>Switchboard</h1>
      <p>Your internal app launcher and security command center.</p>
      <ul class="brand-points">
        <li><i class="pi pi-check" /> Role-based app catalog</li>
        <li><i class="pi pi-check" /> CVE & deployment visibility</li>
        <li><i class="pi pi-check" /> Teams, email & in-app alerts</li>
      </ul>
    </div>
    <div class="auth-panel surface-card">
      <slot />
    </div>
  </div>
</template>

<style scoped>
.auth-shell {
  min-height: 100vh;
  display: grid;
  grid-template-columns: 1fr 1fr;
  background: var(--sb-auth-gradient);
}

.auth-brand {
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 4rem;
  color: #f8fafc;
}

.auth-brand :deep(.switchboard-logo) {
  margin-bottom: 1.5rem;
}

.auth-brand h1 {
  font-size: 2.5rem;
  font-weight: 800;
  margin-bottom: 0.75rem;
}

.auth-brand > p {
  font-size: 1.05rem;
  line-height: 1.6;
  color: #cbd5e1;
  max-width: 28rem;
}

.brand-points {
  list-style: none;
  padding: 0;
  margin: 2rem 0 0;
  display: grid;
  gap: 0.75rem;
}

.brand-points li {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  color: #e2e8f0;
  font-size: 0.95rem;
}

.brand-points i {
  color: #a5b4fc;
  font-size: 0.8rem;
}

.auth-panel {
  margin: 1.5rem;
  padding: 2rem;
  align-self: center;
  width: calc(100% - 3rem);
  max-width: 440px;
  justify-self: center;
}

@media (max-width: 900px) {
  .auth-shell {
    grid-template-columns: 1fr;
  }

  .auth-brand {
    padding: 2.5rem 1.5rem 1rem;
  }

  .auth-brand h1 {
    font-size: 2rem;
  }

  .brand-points {
    display: none;
  }

  .auth-panel {
    margin: 0 1rem 1.5rem;
    width: auto;
  }
}
</style>
