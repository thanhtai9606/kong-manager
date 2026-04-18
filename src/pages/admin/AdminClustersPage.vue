<template>
  <PageHeader :title="t('admin.pages.clusters.title')">
    <template #below-title>
      <SupportText>
        {{ t('admin.clusters.description') }}
      </SupportText>
    </template>
  </PageHeader>

  <KCard class="admin-clusters__card">
    <h3 class="admin-clusters__h3">
      {{ t('admin.clusters.addTitle') }}
    </h3>
    <div class="admin-clusters__form">
      <KInput
        v-model="createForm.name"
        :placeholder="t('admin.clusters.fields.name')"
      />
      <KInput
        v-model="createForm.slug"
        :placeholder="t('admin.clusters.fields.slug')"
      />
      <KInput
        v-model="createForm.admin_base_url"
        :placeholder="t('admin.clusters.fields.adminBaseUrl')"
      />
      <KInput
        v-model="createForm.admin_token"
        type="password"
        :placeholder="t('admin.clusters.fields.adminTokenOptional')"
      />
      <KButton
        appearance="primary"
        :disabled="saving"
        @click="createCluster"
      >
        {{ t('global.buttons.create') }}
      </KButton>
    </div>
  </KCard>

  <KCard>
    <div
      v-if="loading"
      class="admin-clusters__state"
    >
      {{ t('admin.clusters.loading') }}
    </div>
    <p
      v-else-if="errorMessage"
      class="admin-clusters__error"
    >
      {{ errorMessage }}
    </p>
    <div
      v-else
      class="admin-clusters__table-wrap"
    >
      <table class="admin-clusters__table">
        <thead>
          <tr>
            <th>{{ t('admin.clusters.fields.name') }}</th>
            <th>{{ t('admin.clusters.fields.slug') }}</th>
            <th>{{ t('admin.clusters.fields.adminBaseUrl') }}</th>
            <th>{{ t('admin.clusters.fields.enabled') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="c in rows"
            :key="c.id"
          >
            <td>{{ c.name }}</td>
            <td><code>{{ c.slug }}</code></td>
            <td>
              <template v-if="editingId === c.id">
                <KInput v-model="editDraft.url" />
              </template>
              <template v-else>
                {{ c.admin_base_url }}
              </template>
            </td>
            <td class="admin-clusters__cell-switch">
              <KInputSwitch
                size="small"
                :model-value="c.enabled"
                @update:model-value="(v: boolean) => setClusterEnabled(c, v)"
              />
            </td>
            <td class="admin-clusters__actions">
              <template v-if="editingId === c.id">
                <KButton
                  size="small"
                  appearance="primary"
                  @click="saveUrl(c)"
                >
                  {{ t('global.buttons.save') }}
                </KButton>
                <KButton
                  size="small"
                  appearance="tertiary"
                  @click="cancelEdit"
                >
                  {{ t('global.buttons.back') }}
                </KButton>
              </template>
              <template v-else>
                <KButton
                  size="small"
                  appearance="secondary"
                  @click="startEdit(c)"
                >
                  {{ t('global.buttons.edit') }}
                </KButton>
                <KButton
                  v-if="c.slug !== 'default'"
                  size="small"
                  appearance="danger"
                  @click="removeCluster(c)"
                >
                  {{ t('admin.clusters.delete') }}
                </KButton>
              </template>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </KCard>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import type { AxiosError } from 'axios'
import { KButton, KCard, KInput, KInputSwitch } from '@kong/kongponents'
import SupportText from '@/components/SupportText.vue'
import { useI18n } from '@/composables/useI18n'
import { useToaster } from '@/composables/useToaster'
import { apiService } from '@/services/apiService'
import { useKongClusterStore, type KongClusterRow } from '@/stores/kongCluster'

defineOptions({ name: 'AdminClustersPage' })

const { t } = useI18n()
const toaster = useToaster()
const kongClusterStore = useKongClusterStore()

const rows = ref<KongClusterRow[]>([])
const loading = ref(true)
const saving = ref(false)
const errorMessage = ref('')
const editingId = ref<number | null>(null)
const editDraft = reactive({ url: '' })

const createForm = reactive({
  name: '',
  slug: '',
  admin_base_url: '',
  admin_token: '',
})

async function load() {
  loading.value = true
  errorMessage.value = ''
  try {
    const { data } = await apiService.bffGet<KongClusterRow[]>('/api/admin/kong-clusters')
    rows.value = Array.isArray(data) ? data : []
  } catch (e) {
    const err = e as AxiosError
    errorMessage.value = err.response?.status === 403
      ? t('admin.clusters.error.forbidden')
      : t('admin.clusters.error.load')
    rows.value = []
  } finally {
    loading.value = false
  }
}

async function createCluster() {
  saving.value = true
  try {
    await apiService.bffPost('/api/admin/kong-clusters', {
      name: createForm.name.trim(),
      slug: createForm.slug.trim(),
      admin_base_url: createForm.admin_base_url.trim(),
      admin_token: createForm.admin_token.trim() || undefined,
    })
    createForm.name = ''
    createForm.slug = ''
    createForm.admin_base_url = ''
    createForm.admin_token = ''
    toaster.open({ appearance: 'success', message: t('admin.clusters.created') })
    await load()
    await kongClusterStore.loadClusters()
  } catch {
    toaster.open({ appearance: 'danger', message: t('global.error') })
  } finally {
    saving.value = false
  }
}

function startEdit(c: KongClusterRow) {
  editingId.value = c.id
  editDraft.url = c.admin_base_url
}

function cancelEdit() {
  editingId.value = null
  editDraft.url = ''
}

async function saveUrl(c: KongClusterRow) {
  try {
    await apiService.bffPatch(`/api/admin/kong-clusters/${c.id}`, {
      admin_base_url: editDraft.url.trim(),
    })
    toaster.open({ appearance: 'success', message: t('admin.clusters.saved') })
    cancelEdit()
    await load()
    await kongClusterStore.loadClusters()
  } catch {
    toaster.open({ appearance: 'danger', message: t('global.error') })
  }
}

async function setClusterEnabled(c: KongClusterRow, enabled: boolean) {
  if (c.enabled === enabled) {
    return
  }
  try {
    await apiService.bffPatch(`/api/admin/kong-clusters/${c.id}`, { enabled })
    await load()
    await kongClusterStore.loadClusters()
  } catch {
    toaster.open({ appearance: 'danger', message: t('global.error') })
  }
}

async function removeCluster(c: KongClusterRow) {
  if (!globalThis.confirm(t('admin.clusters.confirmDelete', { name: c.name }))) {
    return
  }
  try {
    await apiService.bffDelete(`/api/admin/kong-clusters/${c.id}`)
    toaster.open({ appearance: 'success', message: t('admin.clusters.deleted') })
    await load()
    await kongClusterStore.loadClusters()
  } catch (e) {
    const err = e as AxiosError
    if (err.response?.status === 400) {
      toaster.open({ appearance: 'warning', message: t('admin.clusters.deleteBlocked') })
    } else {
      toaster.open({ appearance: 'danger', message: t('global.error') })
    }
  }
}

onMounted(() => {
  void load()
})
</script>

<style scoped lang="scss">
.admin-clusters__card {
  margin-bottom: 1rem;
}

.admin-clusters__h3 {
  margin: 0 0 0.75rem;
  font-size: 1rem;
}

.admin-clusters__form {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;

  :deep(.k-input) {
    min-width: 10rem;
    flex: 1;
  }
}

.admin-clusters__state {
  padding: 0.5rem 0;
  color: var(--kui-color-text-neutral, #525252);
}

.admin-clusters__error {
  color: var(--kui-color-text-danger, #c20d0d);
  font-size: 0.875rem;
}

.admin-clusters__table-wrap {
  overflow-x: auto;
}

.admin-clusters__table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;

  th,
  td {
    text-align: left;
    padding: 0.5rem 0.65rem;
    border-bottom: 1px solid var(--kui-color-border-neutral, #e7e7e7);
    vertical-align: middle;
  }

  code {
    font-size: 0.8125rem;
  }
}

.admin-clusters__cell-switch {
  vertical-align: middle;
  white-space: nowrap;
}

.admin-clusters__actions {
  white-space: nowrap;
  text-align: right;

  :deep(.k-button) {
    margin-left: 0.25rem;
  }
}
</style>
