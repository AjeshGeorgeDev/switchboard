<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useToast } from 'primevue/usetoast'
import { api } from '../api'
import PageHeader from '../components/PageHeader.vue'
import TableEmptyState from '../components/TableEmptyState.vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import MultiSelect from 'primevue/multiselect'
import Tag from 'primevue/tag'
import Message from 'primevue/message'

const toast = useToast()
const roles = ref<any[]>([])
const users = ref<any[]>([])
const invitations = ref<any[]>([])
const showCreateDialog = ref(false)
const showInviteDialog = ref(false)
const showUserDialog = ref(false)
const showInviteLinkDialog = ref(false)
const inviteLink = ref('')
const inviteEmailSent = ref(false)
const saving = ref(false)
const error = ref('')
const createForm = ref({
  email: '', display_name: '', password: '', role_ids: [] as string[],
})
const inviteForm = ref({
  email: '', display_name: '', role_ids: [] as string[],
})
const selectedUser = ref<any>(null)
const userRoleIds = ref<string[]>([])
const showHistoryDialog = ref(false)
const loginHistory = ref<any[]>([])
const historyLoading = ref(false)

onMounted(load)

async function load() {
  try {
    const [roleList, userList, inviteList] = await Promise.all([
      api.get<any[]>('/api/admin/roles'),
      api.get<any[]>('/api/admin/users'),
      api.get<any[]>('/api/admin/invitations'),
    ])
    roles.value = roleList
    users.value = userList
    invitations.value = inviteList
  } catch (e) {
    toast.add({
      severity: 'error',
      summary: 'Failed to load users',
      detail: e instanceof Error ? e.message : 'Unknown error',
      life: 5000,
    })
  }
}

function defaultRoleIds() {
  const viewer = roles.value.find(r => r.name === 'viewer')
  return viewer ? [viewer.id] : []
}

function openCreateUser() {
  error.value = ''
  createForm.value = {
    email: '', display_name: '', password: '', role_ids: defaultRoleIds(),
  }
  showCreateDialog.value = true
}

function openInviteUser() {
  error.value = ''
  inviteForm.value = {
    email: '', display_name: '', role_ids: defaultRoleIds(),
  }
  showInviteDialog.value = true
}

async function createUser() {
  error.value = ''
  saving.value = true
  try {
    await api.post('/api/admin/users', createForm.value)
    showCreateDialog.value = false
    await load()
    toast.add({ severity: 'success', summary: 'User created', life: 3000 })
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Create failed'
  } finally {
    saving.value = false
  }
}

async function inviteUser() {
  error.value = ''
  saving.value = true
  try {
    const result = await api.post<any>('/api/admin/users/invite', inviteForm.value)
    showInviteDialog.value = false
    inviteLink.value = result.invite_url
    inviteEmailSent.value = result.email_sent
    showInviteLinkDialog.value = true
    await load()
    toast.add({
      severity: 'success',
      summary: 'Invitation created',
      detail: result.email_sent ? 'Email sent to the user.' : 'Copy the invite link and send it manually.',
      life: 5000,
    })
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Invite failed'
  } finally {
    saving.value = false
  }
}

async function editUserRoles(user: any) {
  selectedUser.value = user
  userRoleIds.value = user.role_ids?.length ? [...user.role_ids] : []
  if (!userRoleIds.value.length) {
    try {
      const userRoles = await api.get<any[]>(`/api/admin/users/${user.id}/roles`)
      userRoleIds.value = userRoles.map(r => r.id)
    } catch {
      userRoleIds.value = []
    }
  }
  showUserDialog.value = true
}

async function saveUserRoles() {
  await api.put(`/api/admin/users/${selectedUser.value.id}/roles`, { role_ids: userRoleIds.value })
  showUserDialog.value = false
  selectedUser.value = null
  await load()
  toast.add({ severity: 'success', summary: 'Roles updated', life: 3000 })
}

async function forceLogout(user: any) {
  await api.post(`/api/admin/users/${user.id}/force-logout`, {})
  toast.add({ severity: 'success', summary: 'Sessions revoked', life: 3000 })
}

async function copyInviteLink() {
  await navigator.clipboard.writeText(inviteLink.value)
  toast.add({ severity: 'info', summary: 'Invite link copied', life: 2000 })
}

function formatDate(value: unknown) {
  if (!value) return '—'
  const d = new Date(String(value))
  return Number.isNaN(d.getTime()) ? '—' : d.toLocaleString()
}

async function viewLoginHistory(user: any) {
  selectedUser.value = user
  showHistoryDialog.value = true
  historyLoading.value = true
  loginHistory.value = []
  try {
    loginHistory.value = await api.get<any[]>(`/api/admin/users/${user.id}/sessions`)
  } catch (e) {
    toast.add({
      severity: 'error',
      summary: 'Failed to load login history',
      detail: e instanceof Error ? e.message : 'Unknown error',
      life: 5000,
    })
  } finally {
    historyLoading.value = false
  }
}

function roleNames(roleIds: string[]) {
  return roleIds
    .map(id => roles.value.find(r => r.id === id)?.name)
    .filter(Boolean)
}
</script>

<template>
  <div class="page">
    <PageHeader title="Users" subtitle="Create accounts, send invitations, and manage access.">
      <template #actions>
        <Button label="Invite User" icon="pi pi-envelope" severity="secondary" @click="openInviteUser" class="mr-2"/>
        <Button label="Create User" icon="pi pi-user-plus" @click="openCreateUser" />
      </template>
    </PageHeader>

    <DataTable :value="users" class="surface-card table-card mb-4">
      <Column field="email" header="Email" />
      <Column header="Roles">
        <template #body="{ data }">
          <div class="role-tags">
            <Tag v-for="role in data.roles" :key="role" :value="role" severity="info" />
            <span v-if="!data.roles?.length" class="muted">No roles</span>
          </div>
        </template>
      </Column>
      <Column field="auth_type" header="Auth Type" />
      <Column header="Last Login">
        <template #body="{ data }">
          {{ formatDate(data.last_login_at) }}
        </template>
      </Column>
      <Column field="is_active" header="Active" />
      <Column header="Actions">
        <template #body="{ data }">
          <Button label="History" text @click="viewLoginHistory(data)" />
          <Button label="Roles" text @click="editUserRoles(data)" />
          <Button label="Force Logout" text severity="danger" @click="forceLogout(data)" />
        </template>
      </Column>
      <template #empty>
        <TableEmptyState
          title="No users"
          message="Create a user account or send an invitation to get started."
        />
      </template>
    </DataTable>

    <h2 class="section-title">Pending Invitations</h2>
    <DataTable :value="invitations" class="surface-card table-card">
      <Column field="email" header="Email" />
      <Column header="Roles">
        <template #body="{ data }">
          <div class="role-tags">
            <Tag v-for="role in roleNames(data.role_ids)" :key="role" :value="role" severity="secondary" />
          </div>
        </template>
      </Column>
      <Column field="expires_at" header="Expires" />
      <template #empty>
        <TableEmptyState
          title="No pending invitations"
          message="Invitations you send will appear here until they are accepted or expire."
        />
      </template>
    </DataTable>

    <Dialog v-model:visible="showCreateDialog" header="Create User" modal class="user-dialog">
      <Message v-if="error" severity="error" class="mb-3">{{ error }}</Message>
      <div class="field"><label>Email</label><InputText v-model="createForm.email" type="email" class="w-full" /></div>
      <div class="field"><label>Display Name</label><InputText v-model="createForm.display_name" class="w-full" /></div>
      <div class="field"><label>Password</label><Password v-model="createForm.password" toggle-mask class="w-full" input-class="w-full" /></div>
      <div class="field"><label>Roles</label>
        <MultiSelect v-model="createForm.role_ids" :options="roles" option-label="name" option-value="id" display="chip" class="w-full" />
      </div>
      <template #footer>
        <Button label="Cancel" text severity="secondary" @click="showCreateDialog = false" />
        <Button label="Create" :loading="saving" @click="createUser" />
      </template>
    </Dialog>

    <Dialog v-model:visible="showInviteDialog" header="Invite User" modal class="user-dialog">
      <Message v-if="error" severity="error" class="mb-3">{{ error }}</Message>
      <p class="hint">An invite link is generated. Email is sent automatically when SMTP is configured.</p>
      <div class="field"><label>Email</label><InputText v-model="inviteForm.email" type="email" class="w-full" /></div>
      <div class="field"><label>Display Name</label><InputText v-model="inviteForm.display_name" class="w-full" /></div>
      <div class="field"><label>Roles</label>
        <MultiSelect v-model="inviteForm.role_ids" :options="roles" option-label="name" option-value="id" display="chip" class="w-full" />
      </div>
      <template #footer>
        <Button label="Cancel" text severity="secondary" @click="showInviteDialog = false" />
        <Button label="Send Invite" icon="pi pi-envelope" :loading="saving" @click="inviteUser" />
      </template>
    </Dialog>

    <Dialog v-model:visible="showInviteLinkDialog" header="Invitation Link" modal>
      <p v-if="inviteEmailSent" class="hint success">Invitation email sent.</p>
      <p v-else class="hint">SMTP is not configured — copy this link and send it to the user.</p>
      <InputText :model-value="inviteLink" readonly class="w-full" />
      <template #footer>
        <Button label="Copy Link" icon="pi pi-copy" @click="copyInviteLink" />
        <Button label="Done" @click="showInviteLinkDialog = false" />
      </template>
    </Dialog>

    <Dialog v-model:visible="showUserDialog" :header="`Roles for ${selectedUser?.email}`" modal>
      <MultiSelect v-model="userRoleIds" :options="roles" option-label="name" option-value="id" display="chip" class="w-full mb-3" />
      <Button label="Save" @click="saveUserRoles" />
    </Dialog>

    <Dialog
      v-model:visible="showHistoryDialog"
      :header="`Login history — ${selectedUser?.email}`"
      modal
      class="history-dialog"
    >
      <div v-if="historyLoading" class="history-loading">
        <i class="pi pi-spin pi-spinner" />
        <span>Loading…</span>
      </div>
      <DataTable v-else :value="loginHistory" class="table-card">
        <Column header="Signed in">
          <template #body="{ data }">{{ formatDate(data.issued_at) }}</template>
        </Column>
        <Column field="ip_address" header="IP" />
        <Column field="user_agent" header="User Agent" />
        <Column header="Status">
          <template #body="{ data }">
            <Tag
              :value="data.revoked ? 'Ended' : 'Active'"
              :severity="data.revoked ? 'secondary' : 'success'"
            />
          </template>
        </Column>
        <template #empty>
          <TableEmptyState
            title="No login history"
            message="Sign-in events will appear here after this user logs in."
            icon="pi-history"
          />
        </template>
      </DataTable>
    </Dialog>
  </div>
</template>

<style scoped>
.section-title {
  font-size: 1rem;
  font-weight: 700;
  margin-bottom: 0.75rem;
}

.table-card { overflow: hidden; }
.mb-4 { margin-bottom: 1.5rem; }
.mb-3 { margin-bottom: 1rem; }
.w-full { width: 100%; }

.role-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.muted {
  color: var(--sb-muted);
  font-size: 0.85rem;
}

.hint {
  color: var(--sb-muted);
  font-size: 0.875rem;
  margin-bottom: 1rem;
}

.hint.success {
  color: #15803d;
}

.history-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 2rem;
  color: var(--sb-muted);
}
</style>

<style>
.user-dialog,
.history-dialog {
  width: min(720px, 92vw);
}
</style>
