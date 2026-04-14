<template>
  <PageHeader :title="t('admin.pages.users.title')">
    <template #below-title>
      <SupportText>
        {{ t('admin.users.description') }}
      </SupportText>
    </template>
  </PageHeader>

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
                  @click="toggleEdit(u.id)"
                >
                  {{ t('admin.users.editRoles') }}
                </KButton>
              </td>
            </tr>
            <tr
              v-if="editingUserId === u.id"
              class="admin-users__edit-row"
            >
              <td colspan="4">
                <div class="admin-users__checks">
                  <KCheckbox
                    v-for="g in allGroups"
                    :key="g.id"
                    :model-value="draftGroupIds.has(g.id)"
                    :label="g.name"
                    @change="toggleDraftGroup(g.id)"
                  />
                </div>
                <KButton
                  appearance="primary"
                  class="admin-users__save"
                  :disabled="savingUserId === u.id"
                  @click="saveUserGroups(u)"
                >
                  {{ t('admin.users.saveRoles') }}
                </KButton>
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
import { KBadge, KButton, KCard, KCheckbox } from '@kong/kongponents'
import dayjs from 'dayjs'
import SupportText from '@/components/SupportText.vue'
import { useI18n } from '@/composables/useI18n'
import { useToaster } from '@/composables/useToaster'
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

const users = ref<AdminUserRow[]>([])
const allGroups = ref<AdminGroup[]>([])
const loading = ref(true)
const errorMessage = ref('')
const editingUserId = ref<number | null>(null)
const draftGroupIds = ref<Set<number>>(new Set())
const savingUserId = ref<number | null>(null)

function formatAt(iso: string) {
  return dayjs(iso).format('MMM DD, YYYY, h:mm A')
}

function toggleEdit(userId: number) {
  if (editingUserId.value === userId) {
    editingUserId.value = null
    return
  }
  const u = users.value.find((x) => x.id === userId)
  editingUserId.value = userId
  draftGroupIds.value = new Set((u?.groups ?? []).map((g) => g.id))
}

function toggleDraftGroup(groupId: number) {
  const next = new Set(draftGroupIds.value)
  if (next.has(groupId)) {
    next.delete(groupId)
  } else {
    next.add(groupId)
  }
  draftGroupIds.value = next
}

async function saveUserGroups(u: AdminUserRow) {
  savingUserId.value = u.id
  try {
    const { data } = await apiService.bffPut<AdminUserRow>(`/api/admin/users/${u.id}/groups`, {
      group_ids: [...draftGroupIds.value],
    })
    const idx = users.value.findIndex((x) => x.id === u.id)
    if (idx >= 0) {
      users.value[idx] = data
    }
    editingUserId.value = null
    toaster.open({ appearance: 'success', message: t('admin.users.rolesSaved') })
  } catch {
    toaster.open({ appearance: 'danger', message: t('global.error') })
  } finally {
    savingUserId.value = null
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
}

.admin-users__edit-row td {
  background: var(--kui-color-background-neutral-weakest, #fafafa);
  padding-top: 1rem;
  padding-bottom: 1rem;
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
</style>
