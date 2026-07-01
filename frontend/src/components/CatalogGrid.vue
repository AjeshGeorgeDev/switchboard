<script setup lang="ts">
import { ref, computed } from 'vue'
import AppCard from './AppCard.vue'
import EmptyPanel from './EmptyPanel.vue'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import InputText from 'primevue/inputtext'
import {
  type CatalogSection,
  groupAppsBySection,
  shouldShowSectionHeaders,
} from '../utils/catalog'

const props = withDefaults(defineProps<{
  apps: any[]
  sections?: CatalogSection[]
  preview?: boolean
  variant?: 'default' | 'featured' | 'public'
  emptyMessage?: string
}>(), {
  sections: () => [],
  preview: false,
  variant: 'default',
  emptyMessage: 'Nothing has been assigned to this role yet.',
})

const search = ref('')

const filtered = computed(() =>
  props.apps.filter(a => a.name.toLowerCase().includes(search.value.toLowerCase()))
)

const grouped = computed(() => groupAppsBySection(filtered.value, props.sections))
const showSections = computed(() => shouldShowSectionHeaders(grouped.value))
const isPublic = computed(() => props.variant === 'public' || props.variant === 'featured')

const gridClass = computed(() =>
  isPublic.value
    ? 'grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5'
    : 'grid-cols-[repeat(auto-fill,minmax(260px,1fr))]',
)
</script>

<template>
  <div class="flex flex-col gap-8">
    <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <IconField :class="isPublic ? 'w-full sm:max-w-sm' : 'max-w-[420px]'">
        <InputIcon class="pi pi-search text-muted-color" />
        <InputText
          v-model="search"
          placeholder="Filter applications…"
          class="w-full"
          :class="isPublic ? '!rounded-full !border-surface !bg-surface-0 !py-2.5 !shadow-sm' : ''"
        />
      </IconField>
      <p v-if="apps.length && isPublic" class="text-sm text-muted-color">
        {{ filtered.length }} {{ filtered.length === 1 ? 'app' : 'apps' }}
      </p>
      <span
        v-else-if="apps.length"
        class="inline-flex items-center gap-1.5 rounded-full bg-surface-100 px-3 py-1 text-sm font-medium text-color"
      >
        <i class="pi pi-box text-xs text-primary" />
        {{ filtered.length }} of {{ apps.length }}
      </span>
    </div>

    <template v-if="filtered.length">
      <template v-if="showSections">
        <section
          v-for="group in grouped"
          :key="group.id ?? 'other'"
          class="flex flex-col gap-4"
        >
          <h2 class="text-sm font-semibold tracking-wide text-muted-color uppercase">
            {{ group.name }}
          </h2>
          <div class="grid gap-4" :class="gridClass">
            <AppCard
              v-for="app in group.apps"
              :key="app.id"
              :app="app"
              :preview="preview"
              :public-tile="isPublic"
            />
          </div>
        </section>
      </template>

      <div v-else class="grid gap-4" :class="gridClass">
        <AppCard
          v-for="app in filtered"
          :key="app.id"
          :app="app"
          :preview="preview"
          :public-tile="isPublic"
        />
      </div>
    </template>

    <EmptyPanel
      v-else
      :compact="!isPublic"
      :title="search ? 'No matches' : 'No applications'"
      :message="search ? 'Try a different search term.' : emptyMessage"
      :icon="search ? 'pi-search' : 'pi-th-large'"
    />
  </div>
</template>
