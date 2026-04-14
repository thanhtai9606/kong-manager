<template>
  <div class="login-wrap">
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
          <KInput
            v-model="password"
            type="password"
            autocomplete="current-password"
            :placeholder="t('auth.password')"
            required
            data-testid="login-password"
          />
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
    </KCard>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { KButton, KCard, KInput } from '@kong/kongponents'
import { useI18n } from '@/composables/useI18n'
import { config } from 'config'
import { useAuthStore } from '@/stores/auth'

defineOptions({ name: 'AuthLoginPage' })

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const username = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

function apiBase(): string {
  return config.ADMIN_GUI_PATH.replace(/\/$/, '') || ''
}

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
    if (!res.ok) {
      error.value = t('auth.invalid')
      return
    }
    const data = (await res.json()) as { token?: string }
    if (!data.token) {
      error.value = t('global.error')
      return
    }
    authStore.setSession(data.token)
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
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 2rem;
  box-sizing: border-box;
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

.error {
  color: var(--kui-color-text-danger, #c20d0d);
  margin-bottom: 0.75rem;
  font-size: 0.875rem;
}
</style>
