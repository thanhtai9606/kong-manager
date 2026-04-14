<template>
  <PageHeader :title="t('admin.pages.rbac.title')">
    <template #below-title>
      <SupportText>
        {{ t('admin.rbac.description') }}
      </SupportText>
    </template>
  </PageHeader>

  <KCard class="admin-rbac__card">
    <h3 class="admin-rbac__section-title">
      {{ t('admin.rbac.policiesTitle') }}
    </h3>
    <div
      v-if="loading"
      class="admin-rbac__state"
    >
      {{ t('admin.rbac.loading') }}
    </div>
    <p
      v-else-if="errorMessage"
      class="admin-rbac__error"
    >
      {{ errorMessage }}
    </p>
    <template v-else>
      <div class="admin-rbac__table-wrap">
        <table class="admin-rbac__table">
          <thead>
            <tr>
              <th>{{ t('admin.rbac.headers.role') }}</th>
              <th>{{ t('admin.rbac.headers.resource') }}</th>
              <th>{{ t('admin.rbac.headers.action') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(row, i) in policies"
              :key="'p-' + i"
            >
              <td>{{ row[0] }}</td>
              <td><code>{{ row[1] }}</code></td>
              <td>{{ row[2] }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <h3 class="admin-rbac__section-title admin-rbac__section-title--second">
        {{ t('admin.rbac.groupingTitle') }}
      </h3>
      <div class="admin-rbac__table-wrap">
        <table class="admin-rbac__table">
          <thead>
            <tr>
              <th>{{ t('admin.rbac.headers.user') }}</th>
              <th>{{ t('admin.rbac.headers.group') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(row, i) in grouping"
              :key="'g-' + i"
            >
              <td>{{ row[0] }}</td>
              <td>{{ row[1] }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
  </KCard>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import type { AxiosError } from 'axios'
import { KCard } from '@kong/kongponents'
import SupportText from '@/components/SupportText.vue'
import { useI18n } from '@/composables/useI18n'
import { apiService } from '@/services/apiService'

defineOptions({ name: 'AdminRbacPage' })

interface RbacPayload {
  policies: string[][]
  grouping: string[][]
}

const { t } = useI18n()

const policies = ref<string[][]>([])
const grouping = ref<string[][]>([])
const loading = ref(true)
const errorMessage = ref('')

onMounted(async () => {
  loading.value = true
  errorMessage.value = ''
  try {
    const { data } = await apiService.bffGet<RbacPayload>('/api/admin/rbac')
    policies.value = data.policies ?? []
    grouping.value = data.grouping ?? []
  } catch (e) {
    const err = e as AxiosError
    const status = err.response?.status
    if (status === 403) {
      errorMessage.value = t('admin.rbac.error.forbidden')
    } else {
      errorMessage.value = t('admin.rbac.error.load')
    }
  } finally {
    loading.value = false
  }
})
</script>

<style scoped lang="scss">
.admin-rbac__card {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.admin-rbac__section-title {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
}

.admin-rbac__section-title--second {
  margin-top: 1.25rem;
}

.admin-rbac__state {
  padding: 0.5rem 0 1rem;
  color: var(--kui-color-text-neutral, #525252);
  font-size: 0.875rem;
}

.admin-rbac__table-wrap {
  overflow-x: auto;
}

.admin-rbac__table {
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

  code {
    font-size: 0.8125rem;
    word-break: break-all;
  }
}

.admin-rbac__error {
  color: var(--kui-color-text-danger, #c20d0d);
  font-size: 0.875rem;
  margin: 0 0 0.75rem;
}
</style>
