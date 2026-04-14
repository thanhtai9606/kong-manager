<template>
  <PageHeader :title="t('admin.pages.auditLog.title')">
    <template #below-title>
      <SupportText>
        {{ t('admin.auditLog.description') }}
      </SupportText>
      <p
        v-if="!loading && !errorMessage"
        class="admin-audit__summary"
      >
        {{ t('admin.auditLog.summary', { count: total }) }}
      </p>
    </template>
  </PageHeader>

  <KCard>
    <div
      v-if="loading"
      class="admin-audit__state"
    >
      {{ t('admin.auditLog.loading') }}
    </div>
    <p
      v-else-if="errorMessage"
      class="admin-audit__error"
    >
      {{ errorMessage }}
    </p>
    <div
      v-else-if="items.length === 0"
      class="admin-audit__state"
    >
      {{ t('admin.auditLog.empty') }}
    </div>
    <div
      v-else
      class="admin-audit__table-wrap"
    >
      <table class="admin-audit__table">
        <thead>
          <tr>
            <th>{{ t('admin.auditLog.headers.time') }}</th>
            <th>{{ t('admin.auditLog.headers.actor') }}</th>
            <th>{{ t('admin.auditLog.headers.action') }}</th>
            <th>{{ t('admin.auditLog.headers.resource') }}</th>
            <th>{{ t('admin.auditLog.headers.details') }}</th>
            <th>{{ t('admin.auditLog.headers.ip') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="row in items"
            :key="row.id"
          >
            <td>{{ formatAt(row.created_at) }}</td>
            <td>{{ row.actor_username }}</td>
            <td><code>{{ row.action }}</code></td>
            <td>{{ row.resource || '—' }}</td>
            <td class="admin-audit__details">
              {{ truncateDetails(row.details) }}
            </td>
            <td>{{ row.client_ip || '—' }}</td>
          </tr>
        </tbody>
      </table>
      <div
        v-if="total > limit"
        class="admin-audit__pager"
      >
        <KButton
          appearance="secondary"
          size="small"
          :disabled="offset === 0"
          @click="pagePrev"
        >
          {{ t('admin.auditLog.prev') }}
        </KButton>
        <span class="admin-audit__pager-meta">{{ offset + 1 }}–{{ pageEnd }} / {{ total }}</span>
        <KButton
          appearance="secondary"
          size="small"
          :disabled="offset + limit >= total"
          @click="pageNext"
        >
          {{ t('admin.auditLog.next') }}
        </KButton>
        <KButton
          appearance="secondary"
          size="small"
          @click="reload"
        >
          {{ t('admin.auditLog.refresh') }}
        </KButton>
      </div>
    </div>
  </KCard>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import type { AxiosError } from 'axios'
import { KButton, KCard } from '@kong/kongponents'
import dayjs from 'dayjs'
import SupportText from '@/components/SupportText.vue'
import PageHeader from '@/components/PageHeader.vue'
import { useI18n } from '@/composables/useI18n'
import { apiService } from '@/services/apiService'

defineOptions({ name: 'AdminAuditLogPage' })

interface AuditLogRow {
  id: number
  actor_username: string
  action: string
  resource: string
  details?: string
  client_ip?: string
  created_at: string
}

const { t } = useI18n()

const items = ref<AuditLogRow[]>([])
const total = ref(0)
const limit = 50
const offset = ref(0)
const loading = ref(true)
const errorMessage = ref('')

const pageEnd = computed(() => Math.min(offset.value + items.value.length, total.value))

function formatAt(iso: string) {
  return dayjs(iso).format('MMM DD, YYYY, h:mm:ss A')
}

function truncateDetails(s: string | undefined) {
  if (!s) {
    return '—'
  }
  return s.length > 120 ? `${s.slice(0, 117)}…` : s
}

async function reload() {
  loading.value = true
  errorMessage.value = ''
  try {
    const { data } = await apiService.bffGet<{ items: AuditLogRow[], total: number }>(
      `/api/admin/audit-logs?limit=${limit}&offset=${offset.value}`,
    )
    items.value = Array.isArray(data.items) ? data.items : []
    total.value = typeof data.total === 'number' ? data.total : items.value.length
  } catch (e) {
    const err = e as AxiosError
    errorMessage.value = err.response?.status === 403
      ? t('admin.auditLog.error.forbidden')
      : t('admin.auditLog.error.load')
    items.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function pageNext() {
  offset.value += limit
  void reload()
}

function pagePrev() {
  offset.value = Math.max(0, offset.value - limit)
  void reload()
}

onMounted(() => {
  void reload()
})
</script>

<style scoped lang="scss">
.admin-audit__summary {
  margin: 0.5rem 0 0;
  font-size: 0.875rem;
  color: var(--kui-color-text-neutral, #525252);
}

.admin-audit__state {
  padding: 1rem 0;
  color: var(--kui-color-text-neutral, #525252);
  font-size: 0.875rem;
}

.admin-audit__table-wrap {
  overflow-x: auto;
}

.admin-audit__table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.8125rem;

  th,
  td {
    text-align: left;
    padding: 0.55rem 0.65rem;
    border-bottom: 1px solid var(--kui-color-border-neutral, #e7e7e7);
    vertical-align: top;
  }

  th {
    font-weight: 600;
    color: var(--kui-color-text-neutral, #525252);
  }
}

.admin-audit__details {
  word-break: break-word;
  max-width: 28rem;
}

.admin-audit__error {
  color: var(--kui-color-text-danger, #c20d0d);
  font-size: 0.875rem;
  margin: 0 0 0.5rem;
}

.admin-audit__pager {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-top: 1rem;
  flex-wrap: wrap;
}

.admin-audit__pager-meta {
  font-size: 0.8125rem;
  color: var(--kui-color-text-neutral, #525252);
}
</style>
