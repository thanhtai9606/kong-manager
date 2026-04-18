<template>
  <PageHeader :title="t('admin.notifications.title')">
    <template #below-title>
      <SupportText>
        {{ t('admin.notifications.description') }}
      </SupportText>
    </template>
  </PageHeader>

  <KCard>
    <div class="admin-notify__toolbar">
      <h3 class="admin-notify__list-title">
        {{ t('admin.notifications.channelsList') }}
      </h3>
      <KButton
        appearance="primary"
        @click="openCreateModal"
      >
        {{ t('admin.notifications.addChannel') }}
      </KButton>
    </div>
    <div
      v-if="loading"
      class="admin-notify__state"
    >
      {{ t('admin.notifications.loading') }}
    </div>
    <p
      v-else-if="errorMessage"
      class="admin-notify__error"
    >
      {{ errorMessage }}
    </p>
    <div
      v-else
      class="admin-notify__table-wrap"
    >
      <table class="admin-notify__table">
        <thead>
          <tr>
            <th>{{ t('admin.notifications.fields.name') }}</th>
            <th>{{ t('admin.notifications.fields.type') }}</th>
            <th>{{ t('admin.notifications.fields.hasSecret') }}</th>
            <th>{{ t('admin.notifications.fields.enabled') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="row in rows"
            :key="row.id"
          >
            <td>{{ row.name }}</td>
            <td><code>{{ row.type }}</code></td>
            <td>{{ row.has_secret ? t('admin.notifications.yes') : t('admin.notifications.no') }}</td>
            <td class="admin-notify__cell-switch">
              <KInputSwitch
                size="small"
                :model-value="row.enabled"
                @update:model-value="(v: boolean) => setChannelEnabled(row, v)"
              />
            </td>
            <td class="admin-notify__actions">
              <KButton
                size="small"
                appearance="secondary"
                :disabled="!row.enabled"
                @click="sendTest(row)"
              >
                {{ t('admin.notifications.test') }}
              </KButton>
              <KButton
                size="small"
                appearance="secondary"
                @click="openEditModal(row)"
              >
                {{ t('global.buttons.edit') }}
              </KButton>
              <KButton
                size="small"
                appearance="danger"
                @click="removeChannel(row)"
              >
                {{ t('admin.notifications.delete') }}
              </KButton>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </KCard>

  <KModal
    :visible="createModalVisible"
    :title="t('admin.notifications.createModalTitle')"
    :max-width="'640px'"
    :action-button-text="t('global.buttons.create')"
    :cancel-button-text="t('global.buttons.cancel')"
    :action-button-disabled="savingCreate"
    @cancel="closeCreateModal"
    @proceed="saveCreateModal"
  >
    <div class="admin-notify__modal-body">
      <p class="admin-notify__hint">
        {{ typeHint(createForm.type) }}
      </p>
      <label class="admin-notify__field">
        <span class="admin-notify__label">{{ t('admin.notifications.fields.name') }}</span>
        <KInput v-model="createForm.name" />
      </label>
      <label class="admin-notify__field">
        <span class="admin-notify__label">{{ t('admin.notifications.fields.type') }}</span>
        <select
          v-model="createForm.type"
          class="admin-notify__select"
        >
          <option value="slack">
            slack
          </option>
          <option value="teams">
            teams
          </option>
          <option value="telegram">
            telegram
          </option>
          <option value="email">
            email
          </option>
        </select>
      </label>
      <label class="admin-notify__field">
        <span class="admin-notify__label">{{ t('admin.notifications.fields.secret') }}</span>
        <KInput
          v-model="createForm.secret"
          type="password"
          :placeholder="t('admin.notifications.secretPlaceholder')"
        />
      </label>
      <label class="admin-notify__field">
        <span class="admin-notify__label">{{ t('admin.notifications.fields.configJson') }}</span>
        <textarea
          v-model="createForm.configText"
          class="admin-notify__textarea"
          rows="8"
          spellcheck="false"
        />
      </label>
      <div class="admin-notify__field">
        <KInputSwitch
          v-model="createForm.enabled"
          :label="t('admin.notifications.fields.enabled')"
        />
      </div>
      <label class="admin-notify__field">
        <span class="admin-notify__label">{{ t('admin.notifications.fields.sortOrder') }}</span>
        <KInput
          :model-value="String(createForm.sort_order)"
          type="number"
          @update:model-value="(v: string) => { createForm.sort_order = parseInt(v, 10) || 0 }"
        />
      </label>
    </div>
  </KModal>

  <KModal
    :visible="editModalVisible"
    :title="t('admin.notifications.editTitle')"
    :max-width="'640px'"
    :action-button-text="t('global.buttons.save')"
    :cancel-button-text="t('global.buttons.cancel')"
    :action-button-disabled="savingEdit"
    @cancel="closeEditModal"
    @proceed="saveEditModal"
  >
    <div
      v-if="editingRow"
      class="admin-notify__modal-body"
    >
      <p class="admin-notify__hint">
        {{ typeHint(editDraft.type) }}
      </p>
      <label class="admin-notify__field">
        <span class="admin-notify__label">{{ t('admin.notifications.fields.name') }}</span>
        <KInput v-model="editDraft.name" />
      </label>
      <label class="admin-notify__field">
        <span class="admin-notify__label">{{ t('admin.notifications.fields.type') }}</span>
        <select
          v-model="editDraft.type"
          class="admin-notify__select"
        >
          <option value="slack">
            slack
          </option>
          <option value="teams">
            teams
          </option>
          <option value="telegram">
            telegram
          </option>
          <option value="email">
            email
          </option>
        </select>
      </label>
      <label class="admin-notify__field">
        <span class="admin-notify__label">{{ t('admin.notifications.fields.secret') }}</span>
        <KInput
          v-model="editDraft.secret"
          type="password"
          :placeholder="t('admin.notifications.secretEditPlaceholder')"
        />
      </label>
      <label class="admin-notify__field">
        <span class="admin-notify__label">{{ t('admin.notifications.fields.configJson') }}</span>
        <textarea
          v-model="editDraft.configText"
          class="admin-notify__textarea"
          rows="8"
          spellcheck="false"
        />
      </label>
      <div class="admin-notify__field">
        <KInputSwitch
          v-model="editDraft.enabled"
          :label="t('admin.notifications.fields.enabled')"
        />
      </div>
      <label class="admin-notify__field">
        <span class="admin-notify__label">{{ t('admin.notifications.fields.sortOrder') }}</span>
        <KInput
          :model-value="String(editDraft.sort_order)"
          type="number"
          @update:model-value="(v: string) => { editDraft.sort_order = parseInt(v, 10) || 0 }"
        />
      </label>
    </div>
  </KModal>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import type { AxiosError } from 'axios'
import { KButton, KCard, KInput, KInputSwitch, KModal } from '@kong/kongponents'
import PageHeader from '@/components/PageHeader.vue'
import SupportText from '@/components/SupportText.vue'
import { useI18n } from '@/composables/useI18n'
import { useToaster } from '@/composables/useToaster'
import { apiService } from '@/services/apiService'

defineOptions({ name: 'AdminNotificationsPage' })

export type NotificationChannelRow = {
  id: number
  name: string
  type: string
  config: Record<string, unknown>
  has_secret: boolean
  enabled: boolean
  sort_order: number
  created_at: string
  updated_at: string
}

const { t } = useI18n()
const toaster = useToaster()

const rows = ref<NotificationChannelRow[]>([])
const loading = ref(true)
const errorMessage = ref('')

const createModalVisible = ref(false)
const savingCreate = ref(false)
const editModalVisible = ref(false)
const savingEdit = ref(false)
const editingRow = ref<NotificationChannelRow | null>(null)

const createForm = reactive({
  name: '',
  type: 'slack',
  secret: '',
  configText: '{}',
  enabled: true,
  sort_order: 0,
})

const editDraft = reactive({
  name: '',
  type: 'slack',
  secret: '',
  configText: '{}',
  enabled: true,
  sort_order: 0,
})

function typeHint(type: string): string {
  switch (type) {
    case 'slack':
    case 'teams':
      return t('admin.notifications.hints.webhook')
    case 'telegram':
      return t('admin.notifications.hints.telegram')
    case 'email':
      return t('admin.notifications.hints.email')
    default:
      return ''
  }
}

function configToText(c: Record<string, unknown> | undefined): string {
  if (!c || Object.keys(c).length === 0) {
    return '{}'
  }
  return JSON.stringify(c, null, 2)
}

function parseConfigText(text: string): Record<string, unknown> {
  const raw = text.trim() === '' ? '{}' : text.trim()
  const v = JSON.parse(raw) as unknown
  if (typeof v !== 'object' || v === null || Array.isArray(v)) {
    throw new Error('config must be a JSON object')
  }
  return v as Record<string, unknown>
}

function openCreateModal() {
  createForm.name = ''
  createForm.type = 'slack'
  createForm.secret = ''
  createForm.configText = '{}'
  createForm.enabled = true
  createForm.sort_order = 0
  createModalVisible.value = true
}

function closeCreateModal() {
  createModalVisible.value = false
}

async function saveCreateModal() {
  if (!createForm.name.trim() || !createForm.secret.trim()) {
    toaster.open({ appearance: 'warning', message: t('admin.notifications.validationRequired') })
    return
  }
  let config: Record<string, unknown>
  try {
    config = parseConfigText(createForm.configText)
  } catch {
    toaster.open({ appearance: 'warning', message: t('admin.notifications.invalidConfigJson') })
    return
  }
  savingCreate.value = true
  try {
    await apiService.bffPost('/api/admin/notification-channels', {
      name: createForm.name.trim(),
      type: createForm.type,
      secret: createForm.secret.trim(),
      config,
      enabled: createForm.enabled,
      sort_order: Number.isFinite(createForm.sort_order) ? createForm.sort_order : 0,
    })
    toaster.open({ appearance: 'success', message: t('admin.notifications.created') })
    closeCreateModal()
    await load()
  } catch {
    toaster.open({ appearance: 'danger', message: t('global.error') })
  } finally {
    savingCreate.value = false
  }
}

function openEditModal(row: NotificationChannelRow) {
  editingRow.value = row
  editDraft.name = row.name
  editDraft.type = row.type
  editDraft.secret = ''
  editDraft.configText = configToText(row.config)
  editDraft.enabled = row.enabled
  editDraft.sort_order = row.sort_order ?? 0
  editModalVisible.value = true
}

function closeEditModal() {
  editModalVisible.value = false
  editingRow.value = null
  editDraft.secret = ''
}

async function saveEditModal() {
  if (!editingRow.value) {
    return
  }
  if (!editDraft.name.trim()) {
    toaster.open({ appearance: 'warning', message: t('admin.notifications.validationName') })
    return
  }
  let config: Record<string, unknown>
  try {
    config = parseConfigText(editDraft.configText)
  } catch {
    toaster.open({ appearance: 'warning', message: t('admin.notifications.invalidConfigJson') })
    return
  }
  savingEdit.value = true
  try {
    const body: Record<string, unknown> = {
      name: editDraft.name.trim(),
      type: editDraft.type,
      config,
      enabled: editDraft.enabled,
      sort_order: Number.isFinite(editDraft.sort_order) ? Number(editDraft.sort_order) : 0,
    }
    const sec = editDraft.secret.trim()
    if (sec) {
      body.secret = sec
    }
    await apiService.bffPatch(`/api/admin/notification-channels/${editingRow.value.id}`, body)
    toaster.open({ appearance: 'success', message: t('admin.notifications.saved') })
    closeEditModal()
    await load()
  } catch {
    toaster.open({ appearance: 'danger', message: t('global.error') })
  } finally {
    savingEdit.value = false
  }
}

async function load() {
  loading.value = true
  errorMessage.value = ''
  try {
    const { data } = await apiService.bffGet<NotificationChannelRow[]>('/api/admin/notification-channels')
    rows.value = Array.isArray(data) ? data : []
  } catch (e) {
    const err = e as AxiosError
    errorMessage.value = err.response?.status === 403
      ? t('admin.notifications.error.forbidden')
      : t('admin.notifications.error.load')
    rows.value = []
  } finally {
    loading.value = false
  }
}

async function setChannelEnabled(row: NotificationChannelRow, enabled: boolean) {
  if (row.enabled === enabled) {
    return
  }
  try {
    await apiService.bffPatch(`/api/admin/notification-channels/${row.id}`, {
      enabled,
    })
    await load()
  } catch {
    toaster.open({ appearance: 'danger', message: t('global.error') })
  }
}

async function sendTest(row: NotificationChannelRow) {
  try {
    const { data } = await apiService.bffPost<{ ok: boolean, error?: string }>(
      `/api/admin/notification-channels/${row.id}/test`,
    )
    if (data.ok) {
      toaster.open({ appearance: 'success', message: t('admin.notifications.testOk') })
    } else if (data.error) {
      toaster.open({ appearance: 'danger', message: data.error })
    } else {
      toaster.open({ appearance: 'danger', message: t('global.error') })
    }
  } catch (e) {
    const err = e as AxiosError<{ ok?: boolean, error?: string }>
    const msg = err.response?.data?.error ?? t('global.error')
    toaster.open({ appearance: 'danger', message: msg })
  }
}

async function removeChannel(row: NotificationChannelRow) {
  if (!window.confirm(t('admin.notifications.confirmDelete', { name: row.name }))) {
    return
  }
  try {
    await apiService.bffDelete(`/api/admin/notification-channels/${row.id}`)
    toaster.open({ appearance: 'success', message: t('admin.notifications.deleted') })
    await load()
  } catch {
    toaster.open({ appearance: 'danger', message: t('global.error') })
  }
}

onMounted(() => {
  load()
})
</script>

<style scoped lang="scss">
.admin-notify__toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
}

.admin-notify__list-title {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
}

.admin-notify__state,
.admin-notify__error {
  margin: 0;
  font-size: 0.875rem;
}

.admin-notify__error {
  color: var(--kui-color-text-danger, #c20d0d);
}

.admin-notify__table-wrap {
  overflow-x: auto;
}

.admin-notify__table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;

  th,
  td {
    text-align: left;
    padding: 0.5rem 0.75rem;
    border-bottom: 1px solid var(--kui-color-border-neutral, #e0e0e0);
  }

  th {
    font-weight: 600;
  }
}

.admin-notify__actions {
  white-space: nowrap;
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.admin-notify__modal-body {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.admin-notify__hint {
  margin: 0;
  font-size: 0.8125rem;
  color: var(--kui-color-text-neutral, #6b6b6b);
  line-height: 1.45;
}

.admin-notify__field {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.admin-notify__cell-switch {
  vertical-align: middle;
  white-space: nowrap;
}

.admin-notify__label {
  font-size: 0.8125rem;
  font-weight: 600;
}

.admin-notify__textarea {
  width: 100%;
  box-sizing: border-box;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.8125rem;
  padding: 0.5rem 0.65rem;
  border-radius: 4px;
  border: 1px solid var(--kui-color-border-neutral, #e0e0e0);
  background: var(--kui-color-background, #fff);
  color: inherit;
}

.admin-notify__select {
  max-width: 100%;
  padding: 0.45rem 0.65rem;
  border-radius: 4px;
  border: 1px solid var(--kui-color-border-neutral, #e0e0e0);
  background: var(--kui-color-background, #fff);
  color: inherit;
  font-size: 0.875rem;
}
</style>
