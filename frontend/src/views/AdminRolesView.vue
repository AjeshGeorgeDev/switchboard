<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useToast } from 'primevue/usetoast'
import { api } from '../api'
import PageHeader from '../components/PageHeader.vue'
import TableEmptyState from '../components/TableEmptyState.vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'

const toast = useToast()
const roles = ref<any[]>([])

onMounted(load)

async function load() {
  try {
    roles.value = await api.get('/api/admin/roles')
  } catch (e) {
    toast.add({
      severity: 'error',
      summary: 'Failed to load roles',
      detail: e instanceof Error ? e.message : 'Unknown error',
      life: 5000,
    })
  }
}
</script>

<template>
  <div class="page">
    <PageHeader
      title="Roles"
      subtitle="System roles used for launcher and admin access. Assign them to users from the Users page."
    />

    <DataTable :value="roles" class="surface-card table-card">
      <Column field="name" header="Name" />
      <Column field="description" header="Description" />
      <template #empty>
        <TableEmptyState
          title="No roles"
          message="Roles are created during setup. If this list is empty, check your database migrations."
        />
      </template>
    </DataTable>
  </div>
</template>

<style scoped>
.table-card {
  overflow: hidden;
}
</style>
