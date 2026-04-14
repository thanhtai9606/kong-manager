<template>
  <PageHeader :title="t('admin.sso.title')">
    <template #below-title>
      <SupportText>
        {{ t('admin.sso.description', { origin: publicOrigin, gui: guiPath }) }}
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
            <td>
              <template v-if="editingId === p.id">
                <KInput v-model="editDraft.issuer_url" />
              </template>
              <template v-else>
                {{ p.issuer_url }}
              </template>
            </td>
            <td>
              <KCheckbox
                :model-value="p.enabled"
                @change="() => toggleEnabled(p)"
              />
            </td>
            <td class="admin-sso__actions">
              <template v-if="editingId === p.id">
                <KButton
                  size="small"
                  appearance="primary"
                  @click="saveEdit(p)"
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
                  @click="startEdit(p)"
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
              </template>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </KCard>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import type { AxiosError } from 'axios'
import { KButton, KCard, KCheckbox, KInput } from '@kong/kongponents'
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
const editingId = ref<number | null>(null)
const editDraft = reactive({ issuer_url: '' })

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

function startEdit(p: SSOProviderRow) {
  editingId.value = p.id
  editDraft.issuer_url = p.issuer_url
}

function cancelEdit() {
  editingId.value = null
}

async function saveEdit(p: SSOProviderRow) {
  try {
    await apiService.bffPatch(`/api/admin/sso-providers/${p.id}`, {
      issuer_url: editDraft.issuer_url.trim(),
    })
    editingId.value = null
    toaster.open({ appearance: 'success', message: t('admin.sso.saved') })
    await load()
  } catch {
    toaster.open({ appearance: 'danger', message: t('global.error') })
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

.admin-sso__actions {
  white-space: nowrap;
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}
</style>
