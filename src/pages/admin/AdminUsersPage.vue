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
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="u in users"
            :key="u.id"
          >
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
          </tr>
        </tbody>
      </table>
    </div>
  </KCard>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import type { AxiosError } from 'axios'
import { KBadge, KCard } from '@kong/kongponents'
import dayjs from 'dayjs'
import SupportText from '@/components/SupportText.vue'
import { useI18n } from '@/composables/useI18n'
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

const users = ref<AdminUserRow[]>([])
const loading = ref(true)
const errorMessage = ref('')

function formatAt(iso: string) {
  return dayjs(iso).format('MMM DD, YYYY, h:mm A')
}

onMounted(async () => {
  loading.value = true
  errorMessage.value = ''
  try {
    const { data } = await apiService.bffGet<AdminUserRow[]>('/api/admin/users')
    users.value = Array.isArray(data) ? data : []
  } catch (e) {
    const err = e as AxiosError
    const status = err.response?.status
    if (status === 403) {
      errorMessage.value = t('admin.users.error.forbidden')
    } else {
      errorMessage.value = t('admin.users.error.load')
    }
    users.value = []
  } finally {
    loading.value = false
  }
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
</style>
