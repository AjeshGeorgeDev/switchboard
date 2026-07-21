<script setup lang="ts">
withDefaults(defineProps<{
  app: {
    id: string
    name: string
    description?: string | { String?: string; Valid?: boolean }
    icon_url?: string | { String?: string; Valid?: boolean }
    access_type: string
    target_host: string
    target_port?: number | { Int32?: number; Valid?: boolean }
  }
  preview?: boolean
  featured?: boolean
  publicTile?: boolean
}>(), {
  preview: false,
  featured: false,
  publicTile: false,
})

function textField(value: unknown): string {
  if (!value) return ''
  if (typeof value === 'string') return value
  if (typeof value === 'object' && value !== null && 'String' in value) {
    return (value as { String?: string }).String || ''
  }
  return ''
}

function portNumber(value: unknown): number | null {
  if (typeof value === 'number') return value
  if (typeof value === 'object' && value !== null && 'Int32' in value) {
    return (value as { Int32?: number }).Int32 ?? null
  }
  return null
}

function launchUrl(app: {
  access_type: string
  target_host: string
  target_port?: unknown
}) {
  if (app.access_type === 'url') return app.target_host
  const port = portNumber(app.target_port)
  return `http://${app.target_host}${port ? `:${port}` : ''}`
}

function hue(name: string) {
  let hash = 0
  for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash)
  return Math.abs(hash) % 360
}

function targetLabel(app: {
  access_type: string
  target_host: string
  target_port?: unknown
}) {
  if (app.access_type === 'url') return app.target_host
  const port = portNumber(app.target_port)
  return `${app.target_host}${port ? `:${port}` : ''}`
}

</script>

<template>
  <!-- Public / homepage card — icon, name, link -->
  <component
    v-if="publicTile || featured"
    :is="preview ? 'div' : 'a'"
    class="group relative flex flex-col overflow-hidden rounded-border border border-surface bg-surface-0 p-4 text-left text-inherit no-underline shadow-sm transition-all duration-200 hover:-translate-y-0.5 hover:border-primary/20 hover:shadow-md"
    :class="preview ? 'cursor-default' : 'cursor-pointer'"
    :href="preview ? undefined : launchUrl(app)"
    :target="preview ? undefined : '_blank'"
    :rel="preview ? undefined : 'noopener'"
    :title="launchUrl(app)"
  >
    <div class="flex items-start justify-between gap-3">
      <div
        class="grid aspect-square h-11 w-11 shrink-0 place-items-center overflow-hidden rounded-xl bg-surface-100 text-base font-extrabold text-white shadow-sm"
        :style="!textField(app.icon_url) ? { background: `linear-gradient(145deg, hsl(${hue(app.name)}, 72%, 56%), hsl(${hue(app.name)}, 82%, 38%))` } : undefined"
      >
        <img
          v-if="textField(app.icon_url)"
          :src="textField(app.icon_url)"
          :alt="app.name"
          class="h-full w-full object-contain p-1"
        />
        <span v-else>{{ app.name.charAt(0).toUpperCase() }}</span>
      </div>
      <span
        v-if="!preview"
        class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-surface-100 text-muted-color opacity-0 transition-all duration-200 group-hover:bg-primary group-hover:text-primary-contrast group-hover:opacity-100"
      >
        <i class="pi pi-arrow-up-right text-sm" />
      </span>
    </div>

    <h3 class="mt-3 text-base font-bold text-color">{{ app.name }}</h3>

    <div class="mt-2 flex items-center gap-2">
      <span class="min-w-0 flex-1 truncate font-mono text-xs text-muted-color">
        {{ targetLabel(app) }}
      </span>
    </div>
  </component>

  <!-- Launcher / signed-in card — icon, name, category, link -->
  <component
    v-else
    :is="preview ? 'div' : 'a'"
    class="group relative flex min-h-[168px] flex-col overflow-hidden rounded-border border border-surface bg-surface-0 p-4 text-inherit no-underline shadow-sm transition-all duration-200 hover:-translate-y-0.5 hover:border-primary/20 hover:shadow-md"
    :class="preview ? 'cursor-default' : 'cursor-pointer'"
    :href="preview ? undefined : launchUrl(app)"
    :target="preview ? undefined : '_blank'"
    :rel="preview ? undefined : 'noopener'"
  >
    <div class="flex items-start justify-between gap-3">
      <div
        class="grid aspect-square h-11 w-11 shrink-0 place-items-center overflow-hidden rounded-xl bg-surface-100 text-base font-extrabold text-white shadow-sm"
        :style="!textField(app.icon_url) ? { background: `linear-gradient(145deg, hsl(${hue(app.name)}, 72%, 56%), hsl(${hue(app.name)}, 82%, 38%))` } : undefined"
      >
        <img
          v-if="textField(app.icon_url)"
          :src="textField(app.icon_url)"
          :alt="app.name"
          class="h-full w-full object-contain p-1"
        />
        <span v-else>{{ app.name.charAt(0).toUpperCase() }}</span>
      </div>
      <span
        v-if="!preview"
        class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-surface-100 text-muted-color transition-colors group-hover:bg-primary group-hover:text-primary-contrast"
      >
        <i class="pi pi-arrow-up-right text-sm" />
      </span>
    </div>

    <div class="mt-4 flex flex-1 flex-col gap-1.5">
      <h3 class="text-base font-bold text-color">{{ app.name }}</h3>
      <p class="line-clamp-2 text-sm leading-relaxed text-muted-color">
        {{ textField(app.description) || 'Internal application' }}
      </p>
    </div>

    <div class="mt-auto flex items-center gap-2 pt-3">
      <span class="rounded-md bg-surface-100 px-2 py-0.5 text-[0.68rem] font-semibold uppercase tracking-wide text-muted-color">
        {{ app.access_type === 'url' ? 'Web' : 'Service' }}
      </span>
      <span class="min-w-0 flex-1 truncate font-mono text-xs text-muted-color">
        {{ targetLabel(app) }}
      </span>
    </div>
  </component>
</template>
