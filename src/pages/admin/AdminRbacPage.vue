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
      {{ t('admin.rbac.roles.title') }}
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
      <div class="admin-rbac__create">
        <KInput
          v-model="newRoleName"
          :placeholder="t('admin.rbac.roles.newNamePlaceholder')"
          class="admin-rbac__create-input"
        />
        <KButton
          appearance="primary"
          :disabled="creating"
          @click="createRole"
        >
          {{ t('admin.rbac.roles.create') }}
        </KButton>
      </div>

      <div class="admin-rbac__table-wrap">
        <table class="admin-rbac__table">
          <thead>
            <tr>
              <th>{{ t('admin.rbac.headers.role') }}</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="g in groups"
              :key="g.id"
            >
              <td>
                <template v-if="renamingId === g.id">
                  <KInput
                    v-model="renameDraft"
                    class="admin-rbac__rename-input"
                  />
                </template>
                <template v-else>
                  <b>{{ g.name }}</b>
                  <KBadge
                    v-if="isSystemRole(g.name)"
                    appearance="neutral"
                    class="admin-rbac__sys-badge"
                  >
                    {{ t('admin.rbac.roles.systemRole') }}
                  </KBadge>
                </template>
              </td>
              <td class="admin-rbac__actions">
                <template v-if="renamingId === g.id">
                  <KButton
                    size="small"
                    appearance="primary"
                    @click="commitRename(g)"
                  >
                    {{ t('global.buttons.save') }}
                  </KButton>
                  <KButton
                    size="small"
                    appearance="tertiary"
                    @click="cancelRename"
                  >
                    {{ t('global.buttons.back') }}
                  </KButton>
                </template>
                <template v-else>
                  <KButton
                    size="small"
                    appearance="secondary"
                    @click="openPolicies(g.id)"
                  >
                    {{ t('admin.rbac.roles.policies') }}
                  </KButton>
                  <KButton
                    v-if="!isSystemRole(g.name)"
                    size="small"
                    appearance="tertiary"
                    @click="startRename(g)"
                  >
                    {{ t('admin.rbac.roles.rename') }}
                  </KButton>
                  <KButton
                    v-if="!isSystemRole(g.name)"
                    size="small"
                    appearance="danger"
                    @click="deleteRole(g)"
                  >
                    {{ t('admin.rbac.roles.delete') }}
                  </KButton>
                </template>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div
        v-if="policyGroupId != null"
        class="admin-rbac__policy-editor"
      >
        <h4 class="admin-rbac__policy-title">
          {{ t('admin.rbac.roles.policiesFor', { name: policyGroupName }) }}
        </h4>
        <p class="admin-rbac__hint">
          {{ t('admin.rbac.roles.policyHint') }}
        </p>
        <table class="admin-rbac__table admin-rbac__table--narrow">
          <thead>
            <tr>
              <th>{{ t('admin.rbac.headers.resource') }}</th>
              <th>{{ t('admin.rbac.headers.action') }}</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(row, idx) in policyRows"
              :key="idx"
            >
              <td>
                <KInput v-model="row.object" />
              </td>
              <td>
                <KInput v-model="row.action" />
              </td>
              <td>
                <KButton
                  size="small"
                  appearance="danger"
                  @click="removePolicyRow(idx)"
                >
                  ×
                </KButton>
              </td>
            </tr>
          </tbody>
        </table>
        <div class="admin-rbac__policy-actions">
          <KButton
            appearance="secondary"
            @click="addPolicyRow"
          >
            {{ t('admin.rbac.roles.addPolicyRow') }}
          </KButton>
          <KButton
            appearance="primary"
            :disabled="savingPolicies"
            @click="savePolicies"
          >
            {{ t('admin.rbac.roles.savePolicies') }}
          </KButton>
        </div>
      </div>

      <details class="admin-rbac__raw">
        <summary>{{ t('admin.rbac.roles.rawSnapshot') }}</summary>
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
                v-for="(row, i) in snapshotPolicies"
                :key="'p-' + i"
              >
                <td>{{ row[0] }}</td>
                <td><code>{{ row[1] }}</code></td>
                <td>{{ row[2] }}</td>
              </tr>
            </tbody>
          </table>
        </div>
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
                v-for="(row, i) in snapshotGrouping"
                :key="'g-' + i"
              >
                <td>{{ row[0] }}</td>
                <td>{{ row[1] }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </details>
    </template>
  </KCard>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import type { AxiosError } from 'axios'
import { KBadge, KButton, KCard, KInput } from '@kong/kongponents'
import SupportText from '@/components/SupportText.vue'
import { useI18n } from '@/composables/useI18n'
import { useToaster } from '@/composables/useToaster'
import { apiService } from '@/services/apiService'

defineOptions({ name: 'AdminRbacPage' })

interface GroupDto {
  id: number
  name: string
}

interface PolicyRow {
  object: string
  action: string
}

interface RbacPayload {
  policies: string[][]
  grouping: string[][]
}

const { t } = useI18n()
const toaster = useToaster()

const groups = ref<GroupDto[]>([])
const loading = ref(true)
const errorMessage = ref('')
const creating = ref(false)
const newRoleName = ref('')
const renamingId = ref<number | null>(null)
const renameDraft = ref('')
const policyGroupId = ref<number | null>(null)
const policyRows = ref<PolicyRow[]>([])
const savingPolicies = ref(false)
const snapshotPolicies = ref<string[][]>([])
const snapshotGrouping = ref<string[][]>([])

const policyGroupName = computed(() => {
  const id = policyGroupId.value
  if (id == null) {
    return ''
  }
  return groups.value.find((g) => g.id === id)?.name ?? ''
})

function isSystemRole(name: string) {
  return name === 'admin' || name === 'viewer'
}

async function loadAll() {
  loading.value = true
  errorMessage.value = ''
  try {
    const [gRes, snapRes] = await Promise.all([
      apiService.bffGet<GroupDto[]>('/api/admin/groups'),
      apiService.bffGet<RbacPayload>('/api/admin/rbac'),
    ])
    groups.value = Array.isArray(gRes.data) ? gRes.data : []
    snapshotPolicies.value = snapRes.data.policies ?? []
    snapshotGrouping.value = snapRes.data.grouping ?? []
  } catch (e) {
    const err = e as AxiosError
    errorMessage.value = err.response?.status === 403
      ? t('admin.rbac.error.forbidden')
      : t('admin.rbac.error.load')
  } finally {
    loading.value = false
  }
}

async function createRole() {
  const name = newRoleName.value.trim()
  if (!name) {
    return
  }
  creating.value = true
  try {
    await apiService.bffPost('/api/admin/groups', { name })
    newRoleName.value = ''
    toaster.open({ appearance: 'success', message: t('admin.rbac.roles.roleCreated') })
    await loadAll()
  } catch {
    toaster.open({ appearance: 'danger', message: t('global.error') })
  } finally {
    creating.value = false
  }
}

function startRename(g: GroupDto) {
  renamingId.value = g.id
  renameDraft.value = g.name
}

function cancelRename() {
  renamingId.value = null
  renameDraft.value = ''
}

async function commitRename(g: GroupDto) {
  const name = renameDraft.value.trim()
  if (!name || name === g.name) {
    cancelRename()
    return
  }
  try {
    await apiService.bffPatch(`/api/admin/groups/${g.id}`, { name })
    toaster.open({ appearance: 'success', message: t('admin.rbac.roles.roleRenamed') })
    cancelRename()
    await loadAll()
    if (policyGroupId.value === g.id) {
      await openPolicies(g.id)
    }
  } catch {
    toaster.open({ appearance: 'danger', message: t('global.error') })
  }
}

async function deleteRole(g: GroupDto) {
  if (!globalThis.confirm(t('admin.rbac.roles.confirmDelete', { name: g.name }))) {
    return
  }
  try {
    await apiService.bffDelete(`/api/admin/groups/${g.id}`)
    toaster.open({ appearance: 'success', message: t('admin.rbac.roles.roleDeleted') })
    if (policyGroupId.value === g.id) {
      policyGroupId.value = null
      policyRows.value = []
    }
    await loadAll()
  } catch (e) {
    const err = e as AxiosError
    if (err.response?.status === 409) {
      toaster.open({ appearance: 'warning', message: t('admin.rbac.roles.deleteBlockedUsers') })
    } else {
      toaster.open({ appearance: 'danger', message: t('global.error') })
    }
  }
}

async function openPolicies(id: number) {
  policyGroupId.value = id
  try {
    const { data } = await apiService.bffGet<{ policies: PolicyRow[] }>(`/api/admin/groups/${id}/policies`)
    const list = data.policies ?? []
    policyRows.value = list.length > 0 ? list.map((p) => ({ object: p.object, action: p.action })) : [{ object: '', action: '' }]
  } catch {
    policyRows.value = [{ object: '', action: '' }]
    toaster.open({ appearance: 'danger', message: t('global.error') })
  }
}

function addPolicyRow() {
  policyRows.value.push({ object: '', action: '' })
}

function removePolicyRow(idx: number) {
  policyRows.value.splice(idx, 1)
}

async function savePolicies() {
  const id = policyGroupId.value
  if (id == null) {
    return
  }
  const policies = policyRows.value
    .filter((r) => r.object.trim() && r.action.trim())
    .map((r) => ({ object: r.object.trim(), action: r.action.trim() }))
  savingPolicies.value = true
  try {
    await apiService.bffPut(`/api/admin/groups/${id}/policies`, { policies })
    toaster.open({ appearance: 'success', message: t('admin.rbac.roles.policiesSaved') })
    await loadAll()
  } catch {
    toaster.open({ appearance: 'danger', message: t('global.error') })
  } finally {
    savingPolicies.value = false
  }
}

onMounted(() => {
  void loadAll()
})
</script>

<style scoped lang="scss">
.admin-rbac__card {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.admin-rbac__section-title {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
}

.admin-rbac__state {
  padding: 0.5rem 0 1rem;
  color: var(--kui-color-text-neutral, #525252);
  font-size: 0.875rem;
}

.admin-rbac__error {
  color: var(--kui-color-text-danger, #c20d0d);
  font-size: 0.875rem;
  margin: 0 0 0.75rem;
}

.admin-rbac__create {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
  margin-bottom: 0.75rem;
}

.admin-rbac__create-input {
  max-width: 16rem;
}

.admin-rbac__rename-input {
  max-width: 14rem;
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
    vertical-align: middle;
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

.admin-rbac__table--narrow :deep(.k-input) {
  min-width: 8rem;
}

.admin-rbac__actions {
  white-space: nowrap;
  text-align: right;

  :deep(.k-button) {
    margin-left: 0.25rem;
  }
}

.admin-rbac__sys-badge {
  margin-left: 0.35rem;
}

.admin-rbac__policy-editor {
  margin-top: 1.25rem;
  padding-top: 1rem;
  border-top: 1px solid var(--kui-color-border-neutral, #e7e7e7);
}

.admin-rbac__policy-title {
  margin: 0 0 0.35rem;
  font-size: 0.9375rem;
}

.admin-rbac__hint {
  margin: 0 0 0.75rem;
  font-size: 0.8125rem;
  color: var(--kui-color-text-neutral, #525252);
}

.admin-rbac__policy-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 0.75rem;
}

.admin-rbac__raw {
  margin-top: 1.5rem;
  font-size: 0.875rem;

  summary {
    cursor: pointer;
    font-weight: 600;
    margin-bottom: 0.5rem;
  }
}
</style>
