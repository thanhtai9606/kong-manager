<template>
  <PageHeader :title="t('admin.sso.title')">
    <template #below-title>
      <SupportText>
        {{ t('admin.sso.description', { origin: publicOrigin, gui: guiPath }) }}
      </SupportText>
      <SupportText class="admin-sso__warn">
        {{ t('admin.sso.redirectUriKeycloak') }}
      </SupportText>
    </template>
  </PageHeader>

  <KCard class="admin-sso__card">
    <h3 class="admin-sso__h3">
      {{ t('admin.sso.addTitle') }}
    </h3>
    <div class="admin-sso__form">
      <KInput
        v-model="createForm.slug"
        :placeholder="t('admin.sso.fields.slug')"
      />
      <KInput
        v-model="createForm.name"
        :placeholder="t('admin.sso.fields.name')"
      />
      <KInput
        v-model="createForm.issuer_url"
        :placeholder="t('admin.sso.fields.issuerUrl')"
      />
      <KInput
        v-model="createForm.client_id"
        :placeholder="t('admin.sso.fields.clientId')"
      />
      <KInput
        v-model="createForm.client_secret"
        type="password"
        :placeholder="t('admin.sso.fields.clientSecret')"
      />
      <KInput
        v-model="createForm.scopes"
        :placeholder="t('admin.sso.fields.scopes')"
      />
      <KButton
        appearance="primary"
        :disabled="saving"
        @click="createProvider"
      >
        {{ t('global.buttons.create') }}
      </KButton>
    </div>
  </KCard>

  <KCard>
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
            <td>
              <KCheckbox
                :model-value="p.enabled"
                @change="() => toggleEnabled(p)"
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
            @click="copyRedirectUri"
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

      <label class="admin-sso__field admin-sso__field--inline">
        <KCheckbox
          :model-value="editDraft.enabled"
          @update:model-value="(v: boolean) => { editDraft.enabled = v }"
        />
        <span>{{ t('admin.sso.fields.enabled') }}</span>
      </label>

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
import { KButton, KCard, KCheckbox, KInput, KModal } from '@kong/kongponents'
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
const saving = ref(false)
const errorMessage = ref('')

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

const publicOrigin = computed(() => {
  if (typeof window === 'undefined') {
    return ''
  }
  return window.location.origin
})

const guiPath = computed(() => config.ADMIN_GUI_PATH.replace(/\/$/, '') || '')

const createForm = reactive({
  slug: '',
  name: '',
  issuer_url: '',
  client_id: '',
  client_secret: '',
  scopes: '',
})

/** Must match BFF buildOIDCRedirectURI (same origin + gui + /api/auth/oidc/{slug}/callback). */
function callbackUrlForSlug(slug: string): string {
  const gui = guiPath.value
  return `${publicOrigin.value}${gui}/api/auth/oidc/${encodeURIComponent(slug)}/callback`
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

async function copyRedirectUri() {
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

async function createProvider() {
  saving.value = true
  try {
    await apiService.bffPost('/api/admin/sso-providers', {
      slug: createForm.slug.trim(),
      name: createForm.name.trim(),
      issuer_url: createForm.issuer_url.trim(),
      client_id: createForm.client_id.trim(),
      client_secret: createForm.client_secret.trim(),
      scopes: createForm.scopes.trim() || undefined,
    })
    createForm.slug = ''
    createForm.name = ''
    createForm.issuer_url = ''
    createForm.client_id = ''
    createForm.client_secret = ''
    createForm.scopes = ''
    toaster.open({ appearance: 'success', message: t('admin.sso.created') })
    await load()
  } catch {
    toaster.open({ appearance: 'danger', message: t('global.error') })
  } finally {
    saving.value = false
  }
}

async function toggleEnabled(p: SSOProviderRow) {
  try {
    await apiService.bffPatch(`/api/admin/sso-providers/${p.id}`, {
      enabled: !p.enabled,
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

.admin-sso__card {
  margin-bottom: 1.5rem;
}

.admin-sso__h3 {
  margin: 0 0 1rem;
  font-size: 1rem;
  font-weight: 600;
}

.admin-sso__form {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: flex-end;
}

.admin-sso__form :deep(.k-input) {
  min-width: 200px;
  flex: 1 1 200px;
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

.admin-sso__field--inline {
  flex-direction: row;
  align-items: center;
  gap: 0.5rem;
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
