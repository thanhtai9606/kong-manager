<template>
  <PageHeader :title="t('admin.pages.users.title')">
    <template #below-title>
      <SupportText>
        {{ t('admin.users.description') }}
      </SupportText>
    </template>
  </PageHeader>

  <KCard class="admin-users__create-card">
    <h2 class="admin-users__create-title">
      {{ t('admin.users.createTitle') }}
    </h2>
    <div class="admin-users__create-fields">
      <label class="admin-users__field">
        <span>{{ t('auth.username') }}</span>
        <KInput
          v-model="newUsername"
          autocomplete="off"
          data-testid="admin-new-username"
        />
      </label>
      <label class="admin-users__field">
        <span>{{ t('auth.password') }}</span>
        <KInput
          v-model="newPassword"
          type="password"
          autocomplete="new-password"
          data-testid="admin-new-password"
        />
      </label>
    </div>
    <p class="admin-users__create-hint">
      {{ t('admin.users.createHint') }}
    </p>
    <div
      v-if="allGroups.length"
      class="admin-users__checks admin-users__create-groups"
    >
      <KCheckbox
        v-for="g in allGroups"
        :key="`new-${g.id}`"
        :model-value="newGroupIds.has(g.id)"
        :label="g.name"
        @update:model-value="(v: boolean) => setNewGroup(g.id, v)"
      />
    </div>
    <KButton
      appearance="primary"
      :disabled="creating || !newUsername.trim() || newPassword.length < 8"
      data-testid="admin-create-user"
      @click="createUser"
    >
      {{ t('admin.users.createSubmit') }}
    </KButton>
  </KCard>

  <KCard>
    <div
      v-if="loading"
      class="admin-users__state"
    >
      {{ t('admin.users.loading') }}
    </div>
    <p
      v-else-if="errorMessage"
      class="admin-users__error"
    >
      {{ errorMessage }}
    </p>
    <div
      v-else-if="users.length === 0"
      class="admin-users__state"
    >
      {{ t('admin.users.empty') }}
    </div>
    <div
      v-else
      class="admin-users__table-wrap"
    >
      <table class="admin-users__table">
        <thead>
          <tr>
            <th>{{ t('admin.users.headers.username') }}</th>
            <th>{{ t('admin.users.headers.groups') }}</th>
            <th>{{ t('admin.users.headers.createdAt') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <template
            v-for="u in users"
            :key="u.id"
          >
            <tr>
              <td>
                <b>{{ u.username }}</b>
              </td>
              <td>
                <template v-if="!u.groups?.length">
                  —
                </template>
                <template v-else>
                  <KBadge
                    v-for="g in u.groups"
                    :key="g.id"
                    class="admin-users__badge"
                    appearance="info"
                  >
                    {{ g.name }}
                  </KBadge>
                </template>
              </td>
              <td>{{ formatAt(u.created_at) }}</td>
              <td class="admin-users__cell-actions">
                <KButton
                  size="small"
                  appearance="secondary"
                  @click="openEdit(u)"
                >
                  {{ t('admin.users.editUser') }}
                </KButton>
                <KButton
                  v-if="authStore.username !== u.username"
                  size="small"
                  appearance="danger"
                  class="admin-users__btn-delete"
                  @click="confirmDelete(u)"
                >
                  {{ t('admin.users.delete') }}
                </KButton>
              </td>
            </tr>
            <tr
              v-if="editingUserId === u.id"
              class="admin-users__edit-row"
            >
              <td colspan="4">
                <div class="admin-users__edit-grid">
                  <label class="admin-users__field">
                    <span>{{ t('auth.username') }}</span>
                    <KInput
                      v-model="draftUsername"
                      autocomplete="off"
                    />
                  </label>
                  <label class="admin-users__field">
                    <span>{{ t('admin.users.newPassword') }}</span>
                    <KInput
                      v-model="draftPassword"
                      type="password"
                      autocomplete="new-password"
                      :placeholder="t('admin.users.passwordPlaceholder')"
                    />
                  </label>
                </div>
                <p class="admin-users__edit-hint">
                  {{ t('admin.users.editHint') }}
                </p>
                <div class="admin-users__checks">
                  <KCheckbox
                    v-for="g in allGroups"
                    :key="g.id"
                    :model-value="draftGroupIds.has(g.id)"
                    :label="g.name"
                    @update:model-value="(v: boolean) => setDraftGroup(g.id, v)"
                  />
                </div>
                <div class="admin-users__edit-actions">
                  <KButton
                    appearance="primary"
                    class="admin-users__save"
                    :disabled="savingUserId === u.id"
                    @click="saveUserEdit(u)"
                  >
                    {{ t('admin.users.saveChanges') }}
                  </KButton>
                  <KButton
                    appearance="tertiary"
                    @click="closeEdit"
                  >
                    {{ t('global.buttons.cancel') }}
                  </KButton>
                </div>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>
  </KCard>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import type { AxiosError } from 'axios'
import { KBadge, KButton, KCard, KCheckbox, KInput } from '@kong/kongponents'
import dayjs from 'dayjs'
import SupportText from '@/components/SupportText.vue'
import PageHeader from '@/components/PageHeader.vue'
import { useI18n } from '@/composables/useI18n'
import { useToaster } from '@/composables/useToaster'
import { useAuthStore } from '@/stores/auth'
import { apiService } from '@/services/apiService'

defineOptions({ name: 'AdminUsersPage' })

interface AdminGroup {
  id: number
  name: string
}

interface AdminUserRow {
  id: number
  username: string
  created_at: string
  groups?: AdminGroup[]
}

const { t } = useI18n()
const toaster = useToaster()
const authStore = useAuthStore()

const users = ref<AdminUserRow[]>([])
const allGroups = ref<AdminGroup[]>([])
const loading = ref(true)
const errorMessage = ref('')
const editingUserId = ref<number | null>(null)
const draftUsername = ref('')
const draftPassword = ref('')
const draftGroupIds = ref<Set<number>>(new Set())
const savingUserId = ref<number | null>(null)

const newUsername = ref('')
const newPassword = ref('')
const newGroupIds = ref<Set<number>>(new Set())
const creating = ref(false)

function formatAt(iso: string) {
  return dayjs(iso).format('MMM DD, YYYY, h:mm A')
}

function axiosErrBody(err: unknown): string | null {
  const e = err as AxiosError
  const d = e.response?.data
  if (typeof d === 'string' && d.trim()) {
    return d.trim()
  }
  return null
}

function setNewGroup(groupId: number, checked: boolean) {
  const next = new Set(newGroupIds.value)
  if (checked) {
    next.add(groupId)
  } else {
    next.delete(groupId)
  }
  newGroupIds.value = next
}

function setDraftGroup(groupId: number, checked: boolean) {
  const next = new Set(draftGroupIds.value)
  if (checked) {
    next.add(groupId)
  } else {
    next.delete(groupId)
  }
  draftGroupIds.value = next
}

function openEdit(u: AdminUserRow) {
  if (editingUserId.value === u.id) {
    closeEdit()
    return
  }
  editingUserId.value = u.id
  draftUsername.value = u.username
  draftPassword.value = ''
  draftGroupIds.value = new Set((u.groups ?? []).map((g) => g.id))
}

function closeEdit() {
  editingUserId.value = null
}

async function saveUserEdit(u: AdminUserRow) {
  savingUserId.value = u.id
  try {
    const un = draftUsername.value.trim()
    if (!un) {
      toaster.open({ appearance: 'danger', message: t('admin.users.error.usernameRequired') })
      return
    }
    const pw = draftPassword.value
    if (pw.length > 0 && pw.length < 8) {
      toaster.open({ appearance: 'danger', message: t('admin.users.error.passwordMin') })
      return
    }
    const patchBody: Record<string, string> = {}
    if (un !== u.username) {
      patchBody.username = un
    }
    if (pw.length >= 8) {
      patchBody.password = pw
    }
    if (Object.keys(patchBody).length > 0) {
      await apiService.bffPatch<AdminUserRow>(`/api/admin/users/${u.id}`, patchBody)
    }
    const { data } = await apiService.bffPut<AdminUserRow>(`/api/admin/users/${u.id}/groups`, {
      group_ids: [...draftGroupIds.value],
    })
    const idx = users.value.findIndex((x) => x.id === u.id)
    if (idx >= 0) {
      users.value[idx] = data
    }
    closeEdit()
    toaster.open({ appearance: 'success', message: t('admin.users.saved') })
  } catch (err) {
    const msg = axiosErrBody(err) ?? t('global.error')
    toaster.open({ appearance: 'danger', message: msg })
  } finally {
    savingUserId.value = null
  }
}

function confirmDelete(u: AdminUserRow) {
  if (!window.confirm(t('admin.users.confirmDelete', { name: u.username }))) {
    return
  }
  void deleteUser(u)
}

async function deleteUser(u: AdminUserRow) {
  try {
    await apiService.bffDelete(`/api/admin/users/${u.id}`)
    users.value = users.value.filter((x) => x.id !== u.id)
    if (editingUserId.value === u.id) {
      closeEdit()
    }
    toaster.open({ appearance: 'success', message: t('admin.users.deleted') })
  } catch (err) {
    const msg = axiosErrBody(err) ?? t('global.error')
    toaster.open({ appearance: 'danger', message: msg })
  }
}

async function createUser() {
  creating.value = true
  try {
    await apiService.bffPost<AdminUserRow>('/api/admin/users', {
      username: newUsername.value.trim(),
      password: newPassword.value,
      group_ids: [...newGroupIds.value],
    })
    newUsername.value = ''
    newPassword.value = ''
    newGroupIds.value = new Set()
    toaster.open({ appearance: 'success', message: t('admin.users.created') })
    await load()
  } catch (err) {
    const msg = axiosErrBody(err) ?? t('admin.users.createError')
    toaster.open({ appearance: 'danger', message: msg })
  } finally {
    creating.value = false
  }
}

async function load() {
  loading.value = true
  errorMessage.value = ''
  try {
    const [uRes, gRes] = await Promise.all([
      apiService.bffGet<AdminUserRow[]>('/api/admin/users'),
      apiService.bffGet<AdminGroup[]>('/api/admin/groups'),
    ])
    users.value = Array.isArray(uRes.data) ? uRes.data : []
    allGroups.value = Array.isArray(gRes.data) ? gRes.data : []
  } catch (e) {
    const err = e as AxiosError
    errorMessage.value = err.response?.status === 403
      ? t('admin.users.error.forbidden')
      : t('admin.users.error.load')
    users.value = []
    allGroups.value = []
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void load()
})
</script>

<style scoped lang="scss">
.admin-users__state {
  padding: 1rem 0;
  color: var(--kui-color-text-neutral, #525252);
  font-size: 0.875rem;
}

.admin-users__table-wrap {
  overflow-x: auto;
}

.admin-users__table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;

  th,
  td {
    text-align: left;
    padding: 0.65rem 0.75rem;
    border-bottom: 1px solid var(--kui-color-border-neutral, #e7e7e7);
  }

  th {
    font-weight: 600;
    color: var(--kui-color-text-neutral, #525252);
  }
}

.admin-users__badge {
  margin-right: 0.35rem;
  margin-bottom: 0.25rem;
}

.admin-users__error {
  color: var(--kui-color-text-danger, #c20d0d);
  font-size: 0.875rem;
  margin: 0 0 0.5rem;
}

.admin-users__cell-actions {
  text-align: right;
  white-space: nowrap;

  :deep(.k-button) {
    margin-left: 0.35rem;
  }
}

.admin-users__btn-delete {
  margin-left: 0.25rem;
}

.admin-users__edit-row td {
  background: var(--kui-color-background-neutral-weakest, #fafafa);
  padding-top: 1rem;
  padding-bottom: 1rem;
}

.admin-users__edit-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  margin-bottom: 0.5rem;
}

.admin-users__edit-hint {
  font-size: 0.8125rem;
  color: var(--kui-color-text-neutral, #525252);
  margin: 0 0 0.75rem;
}

.admin-users__edit-actions {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  margin-top: 0.5rem;
}

.admin-users__checks {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem 1.25rem;
  margin-bottom: 0.75rem;
}

.admin-users__save {
  margin-top: 0.25rem;
}

.admin-users__create-card {
  margin-bottom: 1.25rem;
}

.admin-users__create-title {
  font-size: 1rem;
  font-weight: 600;
  margin: 0 0 0.75rem;
}

.admin-users__create-fields {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  margin-bottom: 0.5rem;
}

.admin-users__field {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  min-width: 12rem;

  span {
    font-size: 0.8125rem;
    font-weight: 600;
  }
}

.admin-users__create-hint {
  font-size: 0.8125rem;
  color: var(--kui-color-text-neutral, #525252);
  margin: 0 0 0.75rem;
}

.admin-users__create-groups {
  margin-bottom: 0.75rem;
}
</style>
