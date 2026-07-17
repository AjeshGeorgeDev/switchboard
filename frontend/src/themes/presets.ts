export type ThemePresetId =
  | 'indigo-pulse'
  | 'harbor-mist'
  | 'signal-green'
  | 'ember-glow'
  | 'midnight-sky'
  | 'violet-flare'

export interface ThemePreset {
  id: ThemePresetId
  name: string
  tagline: string
  swatches: [string, string, string]
  previewGradient: string
}

export const THEME_PRESETS: ThemePreset[] = [
  {
    id: 'indigo-pulse',
    name: 'Indigo Pulse',
    tagline: 'Clean indigo accents on soft neutrals.',
    swatches: ['#4f46e5', '#ffffff', '#f4f5f9'],
    previewGradient: 'linear-gradient(135deg, #eef2ff 0%, #4f46e5 100%)',
  },
  {
    id: 'harbor-mist',
    name: 'Harbor Mist',
    tagline: 'Cool teal accents — calm and crisp.',
    swatches: ['#0e7490', '#ffffff', '#f3f7f8'],
    previewGradient: 'linear-gradient(135deg, #ecfeff 0%, #0e7490 100%)',
  },
  {
    id: 'signal-green',
    name: 'Signal Green',
    tagline: 'Soft emerald for an ops-ready feel.',
    swatches: ['#059669', '#ffffff', '#f3f7f5'],
    previewGradient: 'linear-gradient(135deg, #ecfdf5 0%, #059669 100%)',
  },
  {
    id: 'ember-glow',
    name: 'Ember Glow',
    tagline: 'Warm orange accents, still easy on the eyes.',
    swatches: ['#ea580c', '#ffffff', '#f7f4f1'],
    previewGradient: 'linear-gradient(135deg, #fff7ed 0%, #ea580c 100%)',
  },
  {
    id: 'midnight-sky',
    name: 'Midnight Sky',
    tagline: 'Classic blue — sharp and familiar.',
    swatches: ['#2563eb', '#ffffff', '#f3f5f9'],
    previewGradient: 'linear-gradient(135deg, #eff6ff 0%, #2563eb 100%)',
  },
  {
    id: 'violet-flare',
    name: 'Violet Flare',
    tagline: 'Muted violet with a quiet purple accent.',
    swatches: ['#7c3aed', '#ffffff', '#f6f4f8'],
    previewGradient: 'linear-gradient(135deg, #f5f3ff 0%, #7c3aed 100%)',
  },
]

export const DEFAULT_THEME: ThemePresetId = 'indigo-pulse'

export function isThemePresetId(value: string): value is ThemePresetId {
  return THEME_PRESETS.some((p) => p.id === value)
}

export function applyThemePreset(id: ThemePresetId) {
  document.documentElement.setAttribute('data-theme', id)
}
