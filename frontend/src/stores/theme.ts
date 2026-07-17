import { ref } from 'vue'
import { defineStore } from 'pinia'
import { api } from '../api'
import {
  applyThemePreset,
  DEFAULT_THEME,
  isThemePresetId,
  type ThemePresetId,
} from '../themes/presets'

export const useThemeStore = defineStore('theme', () => {
  const preset = ref<ThemePresetId>(DEFAULT_THEME)
  const loaded = ref(false)

  async function fetchTheme() {
    try {
      const data = await api.get<{ theme_preset: string }>('/api/settings/theme')
      if (isThemePresetId(data.theme_preset)) {
        preset.value = data.theme_preset
      }
    } catch {
      preset.value = DEFAULT_THEME
    } finally {
      applyThemePreset(preset.value)
      loaded.value = true
    }
  }

  async function setTheme(id: ThemePresetId) {
    const data = await api.put<{ theme_preset: string }>('/api/admin/settings/theme', {
      theme_preset: id,
    })
    const next = isThemePresetId(data.theme_preset) ? data.theme_preset : id
    preset.value = next
    applyThemePreset(next)
  }

  function previewTheme(id: ThemePresetId) {
    applyThemePreset(id)
  }

  function cancelPreview() {
    applyThemePreset(preset.value)
  }

  return { preset, loaded, fetchTheme, setTheme, previewTheme, cancelPreview }
})
