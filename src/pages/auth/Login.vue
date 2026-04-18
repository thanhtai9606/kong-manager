<template>
  <div class="login-wrap">
    <div
      v-if="loginLogoSrc"
      class="login-brand"
    >
      <img
        :src="loginLogoSrc"
        alt=""
        class="login-brand__img"
        width="200"
        height="48"
      >
    </div>
    <KCard
      class="login-card"
      :title="t('auth.title')"
    >
      <form @submit.prevent="submit">
        <label class="field">
          <span class="label">{{ t('auth.username') }}</span>
          <KInput
            v-model="username"
            autocomplete="username"
            :placeholder="t('auth.username')"
            required
            data-testid="login-username"
          />
        </label>
        <label class="field">
          <span class="label">{{ t('auth.password') }}</span>
          <div class="password-field">
            <div class="password-field__input-wrap">
              <KInput
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                autocomplete="current-password"
                :placeholder="t('auth.password')"
                required
                data-testid="login-password"
              />
            </div>
            <KButton
              type="button"
              appearance="tertiary"
              class="password-field__toggle"
              :aria-label="showPassword ? t('auth.hidePassword') : t('auth.showPassword')"
              :title="showPassword ? t('auth.hidePassword') : t('auth.showPassword')"
              @click="showPassword = !showPassword"
            >
              <VisibilityOffIcon
                v-if="showPassword"
                decorative
                :size="20"
              />
              <VisibilityIcon
                v-else
                decorative
                :size="20"
              />
            </KButton>
          </div>
        </label>
        <p
          v-if="error"
          class="error"
          data-testid="login-error"
        >
          {{ error }}
        </p>
        <KButton
          type="submit"
          appearance="primary"
          :disabled="loading"
          data-testid="login-submit"
        >
          {{ t('auth.submit') }}
        </KButton>
      </form>

      <div
        v-if="ssoProviders.length"
        class="sso-block"
      >
        <p class="sso-label">
          {{ t('auth.ssoDivider') }}
        </p>
        <div class="sso-buttons">
          <KButton
            v-for="p in ssoProviders"
            :key="p.slug"
            appearance="secondary"
            :disabled="loading"
            @click="startSso(p.slug)"
          >
            {{ t('auth.ssoLoginWith', { name: p.name }) }}
          </KButton>
        </div>
      </div>
    </KCard>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { VisibilityIcon, VisibilityOffIcon } from '@kong/icons'
import { KButton, KCard, KInput } from '@kong/kongponents'
import { useI18n } from '@/composables/useI18n'
import { config } from 'config'
import { useAuthStore } from '@/stores/auth'
import defaultLoginLogo from '@/assets/logo.svg?url'

defineOptions({ name: 'AuthLoginPage' })

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const username = ref('')
const password = ref('')
const showPassword = ref(false)
const loading = ref(false)
const error = ref('')
const ssoProviders = ref<Array<{ slug: string, name: string }>>([])

const loginLogoSrc = computed(() => {
  const raw = config.LOGIN_LOGO_URL
  if (raw == null || String(raw).trim() === '') {
    return defaultLoginLogo
  }
  const u = String(raw).trim()
  if (/^https?:\/\//i.test(u)) {
    return u
  }
  if (u.startsWith('/')) {
    return `${typeof window !== 'undefined' ? window.location.origin : ''}${u}`
  }
  return defaultLoginLogo
})

function apiBase(): string {
  return config.ADMIN_GUI_PATH.replace(/\/$/, '') || ''
}

function startSso(slug: string) {
  window.location.assign(`${apiBase()}/api/auth/oidc/${encodeURIComponent(slug)}/login`)
}

async function loadSsoProviders() {
  try {
    const res = await fetch(`${apiBase()}/api/auth/sso/providers`)
    if (!res.ok) {
      return
    }
    const data = (await res.json()) as Array<{ slug: string, name: string }>
    ssoProviders.value = Array.isArray(data) ? data : []
  } catch {
    ssoProviders.value = []
  }
}

onMounted(async () => {
  const h = window.location.hash
  if (h.startsWith('#km_token=')) {
    const token = decodeURIComponent(h.slice('#km_token='.length))
    authStore.setSession(token)
    window.history.replaceState(null, '', `${window.location.pathname}${window.location.search}`)
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/'
    await router.replace(redirect)
    return
  }
  const ssoErr = typeof route.query.sso_error === 'string' ? route.query.sso_error : ''
  if (ssoErr) {
    error.value = t('auth.ssoError', { code: ssoErr })
    const rest = { ...route.query } as Record<string, string | string[] | undefined | null>
    delete rest.sso_error
    await router.replace({ path: route.path, query: rest })
  }
  await loadSsoProviders()
})

async function submit() {
  error.value = ''
  loading.value = true
  try {
    const res = await fetch(`${apiBase()}/api/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        username: username.value,
        password: password.value,
      }),
    })
    const text = await res.text()
    let payload: { token?: string, error?: string } = {}
    try {
      payload = text ? JSON.parse(text) as { token?: string, error?: string } : {}
    } catch {
      payload = {}
    }
    if (!res.ok) {
      if (payload.error === 'sso_required') {
        error.value = t('auth.ssoRequired')
        return
      }
      error.value = t('auth.invalid')
      return
    }
    if (!payload.token) {
      error.value = t('global.error')
      return
    }
    authStore.setSession(payload.token)
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/'
    await router.replace(redirect)
  } catch {
    error.value = t('global.error')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped lang="scss">
.login-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 2rem;
  box-sizing: border-box;
  gap: 1.25rem;
}

.login-brand {
  display: flex;
  justify-content: center;
  width: 100%;
  max-width: 420px;
}

.login-brand__img {
  max-height: 52px;
  width: auto;
  max-width: 100%;
  object-fit: contain;
}

.login-card {
  width: 100%;
  max-width: 420px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  margin-bottom: 1rem;
}

.label {
  font-size: 0.875rem;
  font-weight: 600;
}

.password-field {
  display: flex;
  align-items: stretch;
  gap: 0.25rem;
  width: 100%;
}

.password-field__input-wrap {
  flex: 1;
  min-width: 0;

  :deep(.k-input-wrapper) {
    width: 100%;
  }
}

.password-field__toggle {
  flex-shrink: 0;
  align-self: stretch;
  min-width: 2.5rem;
  padding-left: 0.35rem;
  padding-right: 0.35rem;
}

.error {
  color: var(--kui-color-text-danger, #c20d0d);
  margin-bottom: 0.75rem;
  font-size: 0.875rem;
}

.sso-block {
  margin-top: 1.5rem;
  padding-top: 1.25rem;
  border-top: 1px solid var(--kui-color-border-neutral, #e0e0e0);
}

.sso-label {
  margin: 0 0 0.75rem;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--kui-color-text-neutral, #6b6b6b);
}

.sso-buttons {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
</style>
