<template>
  <PageHeader :title="t('admin.sso.title')">
    <template #below-title>
      <SupportText>
        {{ t('admin.sso.description', { origin: publicOrigin, gui: guiPath }) }}
      </SupportText>
      <SupportText class="admin-sso__warn">
        {{ t('admin.sso.redirectUriKeycloak') }}
      </SupportText>
      <SupportText class="admin-sso__warn">
        {{ t('admin.sso.redirectUriSameHost') }}
      </SupportText>
    </template>
  </PageHeader>

  <KCard>
    <div class="admin-sso__toolbar">
      <h3 class="admin-sso__list-title">
        {{ t('admin.sso.providersList') }}
      </h3>
      <KButton
        appearance="primary"
        @click="openCreateModal"
      >
        {{ t('admin.sso.addProviderButton') }}
      </KButton>
    </div>
    <div
      v-if="loading"
      class="admin-sso__state"
    >
      {{ t('admin.sso.loading') }}
    </div>
    <p
      v-else-if="errorMessage"
      class="admin-sso__error"
    >
      {{ errorMessage }}
    </p>
    <div
      v-else
      class="admin-sso__table-wrap"
    >
      <table class="admin-sso__table">
        <thead>
          <tr>
            <th>{{ t('admin.sso.fields.name') }}</th>
            <th>{{ t('admin.sso.fields.slug') }}</th>
            <th>{{ t('admin.sso.fields.issuerUrl') }}</th>
            <th>{{ t('admin.sso.fields.enabled') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="p in rows"
            :key="p.id"
          >
            <td>{{ p.name }}</td>
            <td><code>{{ p.slug }}</code></td>
            <td class="admin-sso__issuer-cell">
              {{ p.issuer_url }}
            </td>
            <td class="admin-sso__cell-switch">
              <KInputSwitch
                size="small"
                :model-value="p.enabled"
                @update:model-value="(v: boolean) => setProviderEnabled(p, v)"
              />
            </td>
            <td class="admin-sso__actions">
              <KButton
                size="small"
                appearance="secondary"
                @click="openEditModal(p)"
              >
                {{ t('global.buttons.edit') }}
              </KButton>
              <KButton
                size="small"
                appearance="danger"
                @click="removeProvider(p)"
              >
                {{ t('admin.sso.delete') }}
              </KButton>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </KCard>

  <!-- Create -->
  <KModal
    :visible="createModalVisible"
    :title="t('admin.sso.createModalTitle')"
    :max-width="'640px'"
    :action-button-text="t('global.buttons.create')"
    :cancel-button-text="t('global.buttons.cancel')"
    :action-button-disabled="savingCreate"
    @cancel="closeCreateModal"
    @proceed="saveCreateModal"
  >
    <div class="admin-sso__modal-body">
      <p class="admin-sso__modal-lead">
        {{ t('admin.sso.createLead') }}
      </p>

      <label class="admin-sso__field">
        <span class="admin-sso__label">{{ t('admin.sso.fields.slug') }}</span>
        <KInput
          v-model="createForm.slug"
          :placeholder="t('admin.sso.fields.slug')"
        />
      </label>

      <div class="admin-sso__field admin-sso__redirect-block">
        <span class="admin-sso__label">{{ t('admin.sso.redirectUriLabel') }}</span>
        <p class="admin-sso__redirect-hint">
          {{ t('admin.sso.redirectUriRegister') }}
        </p>
        <div class="admin-sso__redirect-row">
          <code class="admin-sso__redirect-url">{{ createRedirectPreview }}</code>
          <KButton
            size="small"
            appearance="secondary"
            :disabled="!createForm.slug.trim()"
            @click="copyCreateRedirectUri"
          >
            {{ t('admin.sso.copyRedirectUri') }}
          </KButton>
        </div>
      </div>

      <label class="admin-sso__field">
        <span class="admin-sso__label">{{ t('admin.sso.fields.name') }}</span>
        <KInput
          v-model="createForm.name"
          :placeholder="t('admin.sso.fields.name')"
        />
      </label>

      <label class="admin-sso__field">
        <span class="admin-sso__label">{{ t('admin.sso.fields.issuerUrl') }}</span>
        <KInput
          v-model="createForm.issuer_url"
          :placeholder="t('admin.sso.fields.issuerUrl')"
        />
      </label>

      <label class="admin-sso__field">
        <span class="admin-sso__label">{{ t('admin.sso.fields.clientId') }}</span>
        <KInput
          v-model="createForm.client_id"
          :placeholder="t('admin.sso.fields.clientId')"
        />
      </label>

      <label class="admin-sso__field">
        <span class="admin-sso__label">{{ t('admin.sso.fields.clientSecret') }}</span>
        <KInput
          v-model="createForm.client_secret"
          type="password"
          :placeholder="t('admin.sso.fields.clientSecret')"
        />
      </label>

      <label class="admin-sso__field">
        <span class="admin-sso__label">{{ t('admin.sso.fields.scopes') }}</span>
        <KInput
          v-model="createForm.scopes"
          :placeholder="t('admin.sso.fields.scopes')"
        />
      </label>

      <div class="admin-sso__field">
        <KInputSwitch
          v-model="createForm.enabled"
          :label="t('admin.sso.fields.enabled')"
        />
      </div>

      <label class="admin-sso__field">
        <span class="admin-sso__label">{{ t('admin.sso.fields.sortOrder') }}</span>
        <KInput
          :model-value="String(createForm.sort_order)"
          type="number"
          @update:model-value="(v: string) => { createForm.sort_order = parseInt(v, 10) || 0 }"
        />
      </label>
    </div>
  </KModal>

  <!-- Edit -->
  <KModal
    :visible="editModalVisible"
    :title="t('admin.sso.editTitle')"
    :max-width="'640px'"
    :action-button-text="t('global.buttons.save')"
    :cancel-button-text="t('global.buttons.cancel')"
    :action-button-disabled="savingModal"
    @cancel="closeEditModal"
    @proceed="saveEditModal"
  >
    <div
      v-if="editingRow"
      class="admin-sso__modal-body"
    >
      <p class="admin-sso__modal-lead">
        {{ t('admin.sso.editLead') }}
      </p>

      <label class="admin-sso__field">
        <span class="admin-sso__label">{{ t('admin.sso.fields.slug') }}</span>
        <code class="admin-sso__readonly">{{ editingRow.slug }}</code>
      </label>

      <div class="admin-sso__field admin-sso__redirect-block">
        <span class="admin-sso__label">{{ t('admin.sso.redirectUriLabel') }}</span>
        <p class="admin-sso__redirect-hint">
          {{ t('admin.sso.redirectUriRegister') }}
        </p>
        <div class="admin-sso__redirect-row">
          <code class="admin-sso__redirect-url">{{ callbackUrlForSlug(editingRow.slug) }}</code>
          <KButton
            size="small"
            appearance="secondary"
            @click="copyEditRedirectUri"
          >
            {{ t('admin.sso.copyRedirectUri') }}
          </KButton>
        </div>
      </div>

      <label class="admin-sso__field">
        <span class="admin-sso__label">{{ t('admin.sso.fields.name') }}</span>
        <KInput v-model="editDraft.name" />
      </label>

      <label class="admin-sso__field">
        <span class="admin-sso__label">{{ t('admin.sso.fields.issuerUrl') }}</span>
        <KInput v-model="editDraft.issuer_url" />
      </label>

      <label class="admin-sso__field">
        <span class="admin-sso__label">{{ t('admin.sso.fields.clientId') }}</span>
        <KInput v-model="editDraft.client_id" />
      </label>

      <label class="admin-sso__field">
        <span class="admin-sso__label">{{ t('admin.sso.fields.clientSecret') }}</span>
        <KInput
          v-model="editDraft.client_secret"
          type="password"
          :placeholder="t('admin.sso.secretPlaceholder')"
        />
      </label>

      <label class="admin-sso__field">
        <span class="admin-sso__label">{{ t('admin.sso.fields.scopes') }}</span>
        <KInput v-model="editDraft.scopes" />
      </label>

      <div class="admin-sso__field">
        <KInputSwitch
          v-model="editDraft.enabled"
          :label="t('admin.sso.fields.enabled')"
        />
      </div>

      <label class="admin-sso__field">
        <span class="admin-sso__label">{{ t('admin.sso.fields.sortOrder') }}</span>
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
import { computed, onMounted, reactive, ref } from 'vue'
import type { AxiosError } from 'axios'
import { KButton, KCard, KInput, KInputSwitch, KModal } from '@kong/kongponents'
import SupportText from '@/components/SupportText.vue'
import { useI18n } from '@/composables/useI18n'
import { useToaster } from '@/composables/useToaster'
import { apiService } from '@/services/apiService'
import { config } from 'config'

defineOptions({ name: 'AdminSSOProvidersPage' })

export type SSOProviderRow = {
  id: number
  slug: string
  name: string
  issuer_url: string
  client_id: string
  has_secret: boolean
  scopes: string
  enabled: boolean
  sort_order: number
}

const { t } = useI18n()
const toaster = useToaster()

const rows = ref<SSOProviderRow[]>([])
const loading = ref(true)
const errorMessage = ref('')

const createModalVisible = ref(false)
const savingCreate = ref(false)

const editModalVisible = ref(false)
const editingRow = ref<SSOProviderRow | null>(null)
const savingModal = ref(false)

const editDraft = reactive({
  name: '',
  issuer_url: '',
  client_id: '',
  client_secret: '',
  scopes: '',
  enabled: true,
  sort_order: 0,
})

const createForm = reactive({
  slug: '',
  name: '',
  issuer_url: '',
  client_id: '',
  client_secret: '',
  scopes: '',
  enabled: true,
  sort_order: 0,
})

const publicOrigin = computed(() => {
  if (typeof window === 'undefined') {
    return ''
  }
  return window.location.origin
})

const guiPath = computed(() => config.ADMIN_GUI_PATH.replace(/\/$/, '') || '')

const createRedirectPreview = computed(() => {
  const s = createForm.slug.trim()
  if (!s) {
    return t('admin.sso.redirectPreviewNeedSlug')
  }
  return callbackUrlForSlug(s)
})

/** Must match BFF buildOIDCRedirectURI (same origin + gui + /api/auth/oidc/{slug}/callback). */
function callbackUrlForSlug(slug: string): string {
  const gui = guiPath.value
  return `${publicOrigin.value}${gui}/api/auth/oidc/${encodeURIComponent(slug)}/callback`
}

function openCreateModal() {
  createForm.slug = ''
  createForm.name = ''
  createForm.issuer_url = ''
  createForm.client_id = ''
  createForm.client_secret = ''
  createForm.scopes = ''
  createForm.enabled = true
  createForm.sort_order = 0
  createModalVisible.value = true
}

function closeCreateModal() {
  createModalVisible.value = false
}

async function saveCreateModal() {
  if (!createForm.slug.trim() || !createForm.name.trim() || !createForm.issuer_url.trim()
    || !createForm.client_id.trim() || !createForm.client_secret.trim()) {
    toaster.open({ appearance: 'warning', message: t('admin.sso.createValidation') })
    return
  }
  savingCreate.value = true
  try {
    await apiService.bffPost('/api/admin/sso-providers', {
      slug: createForm.slug.trim(),
      name: createForm.name.trim(),
      issuer_url: createForm.issuer_url.trim(),
      client_id: createForm.client_id.trim(),
      client_secret: createForm.client_secret.trim(),
      scopes: createForm.scopes.trim() || undefined,
      enabled: createForm.enabled,
      sort_order: Number.isFinite(createForm.sort_order) ? createForm.sort_order : 0,
    })
    toaster.open({ appearance: 'success', message: t('admin.sso.created') })
    closeCreateModal()
    await load()
  } catch {
    toaster.open({ appearance: 'danger', message: t('global.error') })
  } finally {
    savingCreate.value = false
  }
}

async function copyCreateRedirectUri() {
  const s = createForm.slug.trim()
  if (!s) {
    return
  }
  const text = callbackUrlForSlug(s)
  try {
    await navigator.clipboard.writeText(text)
    toaster.open({ appearance: 'success', message: t('admin.sso.redirectUriCopied') })
  } catch {
    toaster.open({ appearance: 'danger', message: t('admin.sso.copyFailed') })
  }
}

function openEditModal(p: SSOProviderRow) {
  editingRow.value = p
  editDraft.name = p.name
  editDraft.issuer_url = p.issuer_url
  editDraft.client_id = p.client_id
  editDraft.client_secret = ''
  editDraft.scopes = p.scopes || ''
  editDraft.enabled = p.enabled
  editDraft.sort_order = p.sort_order ?? 0
  editModalVisible.value = true
}

function closeEditModal() {
  editModalVisible.value = false
  editingRow.value = null
  editDraft.client_secret = ''
}

async function copyEditRedirectUri() {
  if (!editingRow.value) {
    return
  }
  const text = callbackUrlForSlug(editingRow.value.slug)
  try {
    await navigator.clipboard.writeText(text)
    toaster.open({ appearance: 'success', message: t('admin.sso.redirectUriCopied') })
  } catch {
    toaster.open({ appearance: 'danger', message: t('admin.sso.copyFailed') })
  }
}

async function saveEditModal() {
  if (!editingRow.value) {
    return
  }
  savingModal.value = true
  try {
    const body: Record<string, unknown> = {
      name: editDraft.name.trim(),
      issuer_url: editDraft.issuer_url.trim(),
      client_id: editDraft.client_id.trim(),
      scopes: editDraft.scopes.trim(),
      enabled: editDraft.enabled,
      sort_order: Number.isFinite(editDraft.sort_order) ? Number(editDraft.sort_order) : 0,
    }
    const sec = editDraft.client_secret.trim()
    if (sec) {
      body.client_secret = sec
    }
    await apiService.bffPatch(`/api/admin/sso-providers/${editingRow.value.id}`, body)
    toaster.open({ appearance: 'success', message: t('admin.sso.saved') })
    closeEditModal()
    await load()
  } catch {
    toaster.open({ appearance: 'danger', message: t('global.error') })
  } finally {
    savingModal.value = false
  }
}

async function load() {
  loading.value = true
  errorMessage.value = ''
  try {
    const { data } = await apiService.bffGet<SSOProviderRow[]>('/api/admin/sso-providers')
    rows.value = Array.isArray(data) ? data : []
  } catch (e) {
    const err = e as AxiosError
    errorMessage.value = err.response?.status === 403
      ? t('admin.sso.error.forbidden')
      : t('admin.sso.error.load')
    rows.value = []
  } finally {
    loading.value = false
  }
}

async function setProviderEnabled(p: SSOProviderRow, enabled: boolean) {
  if (p.enabled === enabled) {
    return
  }
  try {
    await apiService.bffPatch(`/api/admin/sso-providers/${p.id}`, {
      enabled,
    })
    await load()
  } catch {
    toaster.open({ appearance: 'danger', message: t('global.error') })
  }
}

async function removeProvider(p: SSOProviderRow) {
  if (!window.confirm(t('admin.sso.confirmDelete', { name: p.name }))) {
    return
  }
  try {
    await apiService.bffDelete(`/api/admin/sso-providers/${p.id}`)
    toaster.open({ appearance: 'success', message: t('admin.sso.deleted') })
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
.admin-sso__warn {
  margin-top: 0.75rem;
}

.admin-sso__toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
}

.admin-sso__list-title {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
}

.admin-sso__state,
.admin-sso__error {
  margin: 0;
  font-size: 0.875rem;
}

.admin-sso__error {
  color: var(--kui-color-text-danger, #c20d0d);
}

.admin-sso__table-wrap {
  overflow-x: auto;
}

.admin-sso__table {
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

.admin-sso__issuer-cell {
  max-width: 280px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.admin-sso__actions {
  white-space: nowrap;
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.admin-sso__modal-body {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.admin-sso__modal-lead {
  margin: 0;
  font-size: 0.875rem;
  color: var(--kui-color-text-neutral, #6b6b6b);
}

.admin-sso__field {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.admin-sso__cell-switch {
  vertical-align: middle;
  white-space: nowrap;
}

.admin-sso__label {
  font-size: 0.8125rem;
  font-weight: 600;
}

.admin-sso__readonly {
  font-size: 0.875rem;
  padding: 0.25rem 0;
}

.admin-sso__redirect-block {
  padding: 0.75rem;
  border-radius: 4px;
  background: var(--kui-color-background-neutral-weakest, #f5f5f5);
}

.admin-sso__redirect-hint {
  margin: 0;
  font-size: 0.8125rem;
  color: var(--kui-color-text-neutral, #6b6b6b);
}

.admin-sso__redirect-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: flex-start;
}

.admin-sso__redirect-url {
  flex: 1 1 200px;
  font-size: 0.75rem;
  word-break: break-all;
  line-height: 1.4;
}
</style>
