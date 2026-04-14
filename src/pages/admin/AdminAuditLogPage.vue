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
      <div class="admin-audit__table-panel">
        <table class="admin-audit__table">
          <thead>
            <tr
              v-for="headerGroup in table.getHeaderGroups()"
              :key="headerGroup.id"
            >
              <th
                v-for="header in headerGroup.headers"
                :key="header.id"
                scope="col"
              >
                <FlexRender
                  v-if="!header.isPlaceholder"
                  :render="header.column.columnDef.header"
                  :props="header.getContext()"
                />
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="row in table.getRowModel().rows"
              :key="row.id"
            >
              <td
                v-for="cell in row.getVisibleCells()"
                :key="cell.id"
                :class="{
                  'admin-audit__cell--details': cell.column.id === 'details',
                  'admin-audit__cell--time': cell.column.id === 'created_at',
                }"
              >
                <FlexRender
                  :render="cell.column.columnDef.cell"
                  :props="cell.getContext()"
                />
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div
        v-if="showPager"
        class="admin-audit__toolbar"
      >
        <AuditLogPagination
          :page-size="pageSize"
          :offset="offset"
          :total="total"
          :page-end="pageEnd"
          :current-page="currentPage"
          :total-pages="totalPages"
          :last-offset="lastOffset"
          @update:page-size="onPageSizeUpdate"
          @first="goFirst"
          @prev="pagePrev"
          @next="pageNext"
          @last="goLast"
          @refresh="reload"
        />
      </div>
    </div>
  </KCard>
</template>

<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue'
import type { AxiosError } from 'axios'
import { KCard } from '@kong/kongponents'
import {
  FlexRender,
  createColumnHelper,
  getCoreRowModel,
  useVueTable,
} from '@tanstack/vue-table'
import dayjs from 'dayjs'
import SupportText from '@/components/SupportText.vue'
import PageHeader from '@/components/PageHeader.vue'
import AuditLogPagination from '@/components/admin/AuditLogPagination.vue'
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
const pageSize = ref(50)
const offset = ref(0)
const loading = ref(true)
const errorMessage = ref('')

const pageEnd = computed(() => Math.min(offset.value + items.value.length, total.value))
const totalPages = computed(() => (total.value <= 0 ? 1 : Math.ceil(total.value / pageSize.value)))
const currentPage = computed(() =>
  total.value <= 0 ? 1 : Math.min(totalPages.value, Math.floor(offset.value / pageSize.value) + 1),
)
const showPager = computed(() => total.value > 0)
const lastOffset = computed(() =>
  total.value <= 0 ? 0 : Math.max(0, (totalPages.value - 1) * pageSize.value),
)

function formatAt(iso: string) {
  return dayjs(iso).format('MMM DD, YYYY, h:mm:ss A')
}

function truncateDetails(s: string | undefined) {
  if (!s) {
    return '—'
  }
  return s.length > 120 ? `${s.slice(0, 117)}…` : s
}

const columnHelper = createColumnHelper<AuditLogRow>()

const columns = computed(() => [
  columnHelper.accessor('created_at', {
    header: () => t('admin.auditLog.headers.time'),
    cell: info => formatAt(info.getValue()),
  }),
  columnHelper.accessor('actor_username', {
    header: () => t('admin.auditLog.headers.actor'),
    cell: info => info.getValue(),
  }),
  columnHelper.accessor('action', {
    header: () => t('admin.auditLog.headers.action'),
    cell: info =>
      h(
        'code',
        { class: 'admin-audit__action-code' },
        info.getValue(),
      ),
  }),
  columnHelper.accessor('resource', {
    header: () => t('admin.auditLog.headers.resource'),
    cell: info => info.getValue() || '—',
  }),
  columnHelper.display({
    id: 'details',
    header: () => t('admin.auditLog.headers.details'),
    cell: ({ row }) => truncateDetails(row.original.details),
  }),
  columnHelper.accessor('client_ip', {
    header: () => t('admin.auditLog.headers.ip'),
    cell: info => info.getValue() || '—',
  }),
])

const table = useVueTable({
  get data() {
    return items.value
  },
  get columns() {
    return columns.value
  },
  getCoreRowModel: getCoreRowModel(),
})

async function reload() {
  loading.value = true
  errorMessage.value = ''
  try {
    const { data } = await apiService.bffGet<{ items: AuditLogRow[], total: number }>(
      `/api/admin/audit-logs?limit=${pageSize.value}&offset=${offset.value}`,
    )
    items.value = Array.isArray(data.items) ? data.items : []
    total.value = typeof data.total === 'number' ? data.total : items.value.length
    if (total.value > 0 && offset.value >= total.value) {
      offset.value = lastOffset.value
      const { data: data2 } = await apiService.bffGet<{ items: AuditLogRow[], total: number }>(
        `/api/admin/audit-logs?limit=${pageSize.value}&offset=${offset.value}`,
      )
      items.value = Array.isArray(data2.items) ? data2.items : []
    }
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
  if (offset.value + pageSize.value >= total.value) {
    return
  }
  offset.value += pageSize.value
  void reload()
}

function pagePrev() {
  offset.value = Math.max(0, offset.value - pageSize.value)
  void reload()
}

function goFirst() {
  if (offset.value === 0) {
    return
  }
  offset.value = 0
  void reload()
}

function goLast() {
  const lo = lastOffset.value
  if (offset.value === lo) {
    return
  }
  offset.value = lo
  void reload()
}

function onPageSizeUpdate(n: number) {
  pageSize.value = n
  offset.value = 0
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

.admin-audit__table-panel {
  border: 1px solid var(--kui-color-border-neutral, #e0e0e0);
  border-radius: var(--kui-border-radius-30, 8px);
  background: var(--kui-color-background, #fff);
  box-shadow:
    0 1px 2px rgb(0 0 0 / 0.04),
    0 2px 8px rgb(0 0 0 / 0.06);
  overflow: hidden;
}

.admin-audit__table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.8125rem;
  line-height: 1.45;

  th,
  td {
    text-align: left;
    padding: 0.65rem 0.85rem;
    border-bottom: 1px solid var(--kui-color-border-neutral-weaker, #ececec);
    vertical-align: top;
  }

  thead th {
    position: sticky;
    top: 0;
    z-index: 1;
    font-weight: 600;
    font-size: 0.75rem;
    letter-spacing: 0.02em;
    text-transform: uppercase;
    color: var(--kui-color-text-neutral-stronger, #3d3d3d);
    background: linear-gradient(
      180deg,
      var(--kui-color-background-neutral-weakest, #fafafa) 0%,
      var(--kui-color-background-neutral-weak, #f0f0f0) 100%
    );
    border-bottom: 2px solid var(--kui-color-border-neutral, #d9d9d9);
    box-shadow: 0 1px 0 rgb(0 0 0 / 0.03);
  }

  tbody tr {
    transition: background-color 0.12s ease;
  }

  tbody tr:nth-child(even) td {
    background: var(--kui-color-background-neutral-weakest, rgb(0 0 0 / 0.02));
  }

  tbody tr:nth-child(odd) td {
    background: var(--kui-color-background, #fff);
  }

  tbody tr:hover td {
    background: var(--kui-color-background-primary-weakest, rgb(59 130 246 / 0.08)) !important;
  }

  tbody tr:last-child td {
    border-bottom: none;
  }
}

.admin-audit__cell--time {
  font-variant-numeric: tabular-nums;
  font-family: ui-monospace, 'Cascadia Code', 'SF Mono', Menlo, monospace;
  font-size: 0.8rem;
  color: var(--kui-color-text-neutral, #525252);
  white-space: nowrap;
}

.admin-audit__cell--details {
  word-break: break-word;
  max-width: 28rem;
  color: var(--kui-color-text-neutral, #525252);
}

/* Rendered inside TanStack FlexRender — needs :deep */
.admin-audit__table :deep(.admin-audit__action-code) {
  display: inline-block;
  margin: 0;
  padding: 0.15rem 0.45rem;
  font-family: ui-monospace, 'Cascadia Code', 'SF Mono', Menlo, monospace;
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--kui-color-text-primary, #1155cc);
  background: rgb(17 85 204 / 0.08);
  border: 1px solid rgb(17 85 204 / 0.22);
  border-radius: var(--kui-border-radius-20, 4px);
}

.admin-audit__error {
  color: var(--kui-color-text-danger, #c20d0d);
  font-size: 0.875rem;
  margin: 0 0 0.5rem;
}

.admin-audit__toolbar {
  margin-top: 0.75rem;
}
</style>
