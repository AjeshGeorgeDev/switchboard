<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { useToast } from 'primevue/usetoast'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import { useThemeStore } from '../../../stores/theme'
import { THEME_PRESETS, type ThemePresetId } from '../../../themes/presets'

const toast = useToast()
const theme = useThemeStore()
const saving = ref(false)
const selected = ref<ThemePresetId>(theme.preset)

onMounted(() => {
  selected.value = theme.preset
})

onUnmounted(() => {
  theme.cancelPreview()
})

function pick(id: ThemePresetId) {
  selected.value = id
  theme.previewTheme(id)
}

async function save() {
  saving.value = true
  try {
    await theme.setTheme(selected.value)
    toast.add({
      severity: 'success',
      summary: 'Theme saved',
      detail: 'Appearance updated for everyone.',
      life: 3000,
    })
  } catch (e) {
    toast.add({
      severity: 'error',
      summary: 'Failed to save theme',
      detail: e instanceof Error ? e.message : 'Unknown error',
      life: 5000,
    })
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <section class="config-section">
    <header class="config-card-header">
      <div>
        <h2>Appearance</h2>
        <p class="config-card-lead">
          Pick a modern preset with gradients and accent colors. Applies org-wide for all users.
        </p>
      </div>
      <Button
        label="Save theme"
        icon="pi pi-check"
        :loading="saving"
        :disabled="selected === theme.preset"
        @click="save"
      />
    </header>

    <div class="preset-grid">
      <button
        v-for="preset in THEME_PRESETS"
        :key="preset.id"
        type="button"
        class="preset-card"
        :class="{ selected: selected === preset.id, active: theme.preset === preset.id }"
        @click="pick(preset.id)"
      >
        <div class="preset-preview" :style="{ background: preset.previewGradient }">
          <div class="preview-window">
            <span class="preview-dot" />
            <span class="preview-dot" />
            <span class="preview-dot" />
          </div>
          <div class="preview-bar" />
        </div>
        <div class="preset-meta">
          <div class="preset-title-row">
            <h3>{{ preset.name }}</h3>
            <Tag v-if="theme.preset === preset.id" value="Active" severity="success" />
          </div>
          <p>{{ preset.tagline }}</p>
          <div class="swatches">
            <span
              v-for="(color, i) in preset.swatches"
              :key="i"
              class="swatch"
              :style="{ background: color }"
            />
          </div>
        </div>
      </button>
    </div>

    <p class="hint">
      Click a card to live-preview. Press <strong>Save theme</strong> to keep it for everyone.
    </p>
  </section>
</template>

<style scoped>
.preset-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 1rem;
  margin-top: 0.5rem;
}

.preset-card {
  display: flex;
  flex-direction: column;
  text-align: left;
  border: 1px solid var(--sb-border);
  border-radius: var(--sb-radius);
  background: var(--sb-surface);
  padding: 0;
  overflow: hidden;
  cursor: pointer;
  transition: border-color 0.15s, box-shadow 0.15s, transform 0.15s;
}

.preset-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--sb-shadow);
}

.preset-card.selected {
  border-color: var(--sb-primary);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--sb-primary) 28%, transparent);
}

.preset-preview {
  height: 7.5rem;
  padding: 0.85rem;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.preview-window {
  display: flex;
  gap: 0.3rem;
  width: fit-content;
  padding: 0.35rem 0.5rem;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.18);
  backdrop-filter: blur(6px);
}

.preview-dot {
  width: 0.45rem;
  height: 0.45rem;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.75);
}

.preview-bar {
  height: 0.55rem;
  width: 55%;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.35);
}

.preset-meta {
  padding: 0.9rem 1rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.preset-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}

.preset-meta h3 {
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--sb-text);
}

.preset-meta p {
  font-size: 0.8rem;
  line-height: 1.4;
  color: var(--sb-muted);
}

.swatches {
  display: flex;
  gap: 0.35rem;
  margin-top: 0.25rem;
}

.swatch {
  width: 1rem;
  height: 1rem;
  border-radius: 999px;
  border: 1px solid rgba(15, 23, 42, 0.12);
}

.hint {
  margin-top: 1rem;
  font-size: 0.85rem;
  color: var(--sb-muted);
}
</style>
