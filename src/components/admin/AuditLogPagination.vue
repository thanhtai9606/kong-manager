<template>
  <div class="audit-log-pagination">
    <label class="audit-log-pagination__size">
      <span>{{ t('admin.auditLog.pageSize') }}</span>
      <select
        class="audit-log-pagination__select"
        :value="pageSize"
        @change="onSizeChange"
      >
        <option
          v-for="n in pageSizeOptions"
          :key="n"
          :value="n"
        >
          {{ n }}
        </option>
      </select>
    </label>
    <span class="audit-log-pagination__meta">
      {{ t('admin.auditLog.range', { start: rangeStart, end: pageEnd, total }) }}
    </span>
    <span class="audit-log-pagination__meta">
      {{ t('admin.auditLog.pageOf', { current: currentPage, total: totalPages }) }}
    </span>
    <KButton
      appearance="secondary"
      size="small"
      :disabled="offset === 0"
      @click="emit('first')"
    >
      {{ t('admin.auditLog.first') }}
    </KButton>
    <KButton
      appearance="secondary"
      size="small"
      :disabled="offset === 0"
      @click="emit('prev')"
    >
      {{ t('admin.auditLog.prev') }}
    </KButton>
    <KButton
      appearance="secondary"
      size="small"
      :disabled="offset + pageSize >= total"
      @click="emit('next')"
    >
      {{ t('admin.auditLog.next') }}
    </KButton>
    <KButton
      appearance="secondary"
      size="small"
      :disabled="offset >= lastOffset"
      @click="emit('last')"
    >
      {{ t('admin.auditLog.last') }}
    </KButton>
    <KButton
      appearance="secondary"
      size="small"
      @click="emit('refresh')"
    >
      {{ t('admin.auditLog.refresh') }}
    </KButton>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { KButton } from '@kong/kongponents'
import { useI18n } from '@/composables/useI18n'

defineOptions({ name: 'AuditLogPagination' })

const props = defineProps<{
  pageSize: number
  offset: number
  total: number
  pageEnd: number
  currentPage: number
  totalPages: number
  lastOffset: number
}>()

const emit = defineEmits<{
  first: []
  prev: []
  next: []
  last: []
  refresh: []
  'update:pageSize': [n: number]
}>()

const { t } = useI18n()

const pageSizeOptions = [25, 50, 100, 200] as const

const rangeStart = computed(() => (props.total === 0 ? 0 : props.offset + 1))

function onSizeChange(ev: Event) {
  const n = Number((ev.target as HTMLSelectElement).value)
  if (Number.isFinite(n) && n > 0) {
    emit('update:pageSize', n)
  }
}
</script>

<style scoped lang="scss">
.audit-log-pagination {
  display: flex;
  align-items: center;
  gap: 0.5rem 0.75rem;
  flex-wrap: wrap;
}

.audit-log-pagination__size {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.8125rem;
  color: var(--kui-color-text-neutral, #525252);

  span {
    white-space: nowrap;
  }
}

.audit-log-pagination__select {
  min-width: 4.5rem;
  padding: 0.35rem 0.5rem;
  font-size: 0.8125rem;
  border: 1px solid var(--kui-color-border-neutral, #e7e7e7);
  border-radius: 4px;
  background: var(--kui-color-background, #fff);
}

.audit-log-pagination__meta {
  font-size: 0.8125rem;
  color: var(--kui-color-text-neutral, #525252);
  white-space: nowrap;
}
</style>
