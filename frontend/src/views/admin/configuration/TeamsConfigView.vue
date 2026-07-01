<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useToast } from 'primevue/usetoast'
import { api } from '../../../api'
import TableEmptyState from '../../../components/TableEmptyState.vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Checkbox from 'primevue/checkbox'

const toast = useToast()
const teamsWebhooks = ref<any[]>([])
const showDialog = ref(false)
const form = ref({ name: '', webhook_url: '', event_types: ['deployment_report'], is_active: true })

onMounted(load)

async function load() {
  try {
    teamsWebhooks.value = await api.get<any[]>('/api/admin/teams-webhooks')
  } catch (e) {
    toast.add({
      severity: 'error',
      summary: 'Failed to load Teams webhooks',
      detail: e instanceof Error ? e.message : 'Unknown error',
      life: 5000,
    })
  }
}

async function save() {
  await api.post('/api/admin/teams-webhooks', form.value)
  showDialog.value = false
  form.value = { name: '', webhook_url: '', event_types: ['deployment_report'], is_active: true }
  await load()
}

async function remove(row: any) {
  await api.delete(`/api/admin/teams-webhooks/${row.id}`)
  await load()
}
</script>

<template>
  <section class="config-section">
    <header class="config-card-header">
      <div>
        <h2>Microsoft Teams</h2>
        <p class="config-card-lead">
          Send deployment reports and critical CVE alerts to Teams channels via incoming webhooks.
        </p>
      </div>
      <Button label="Add webhook" icon="pi pi-plus" @click="showDialog = true" />
    </header>

    <ol class="setup-steps compact">
      <li>In Teams, open the target channel → <em>⋯ → Connectors</em> (or <em>Workflows</em>).</li>
      <li>Add an <strong>Incoming Webhook</strong> connector and copy the generated URL.</li>
      <li>Paste it below. Switchboard sends alerts for deployment reports and critical CVEs.</li>
    </ol>

    <DataTable :value="teamsWebhooks" class="table-card">
      <Column field="name" header="Name" />
      <Column field="webhook_url" header="Webhook URL" />
      <Column field="event_types" header="Events" />
      <Column field="is_active" header="Active" />
      <Column header="">
        <template #body="{ data }">
          <Button icon="pi pi-trash" text severity="danger" @click="remove(data)" />
        </template>
      </Column>
      <template #empty>
        <TableEmptyState
          title="No Teams webhooks"
          message="Add an incoming webhook to receive security alerts in Teams."
          icon="pi-microsoft"
        />
      </template>
    </DataTable>

    <Dialog v-model:visible="showDialog" header="Add Teams webhook" modal>
      <div class="field"><label>Name</label><InputText v-model="form.name" class="w-full" placeholder="Security alerts" /></div>
      <div class="field">
        <label>Webhook URL</label>
        <InputText v-model="form.webhook_url" class="w-full" placeholder="https://outlook.office.com/webhook/..." />
      </div>
      <div class="field-row">
        <Checkbox v-model="form.is_active" input-id="teams-active" binary />
        <label for="teams-active">Active</label>
      </div>
      <Button label="Save" @click="save" />
    </Dialog>
  </section>
</template>
