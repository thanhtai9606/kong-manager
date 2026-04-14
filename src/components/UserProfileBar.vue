<template>
  <div
    class="user-profile-bar"
    data-testid="user-profile-bar"
  >
    <div
      class="user-profile-bar__avatar"
      aria-hidden="true"
    >
      {{ initials }}
    </div>
    <div class="user-profile-bar__meta">
      <span class="user-profile-bar__name">{{ displayName }}</span>
      <KButton
        appearance="tertiary"
        class="user-profile-bar__signout"
        data-testid="user-sign-out"
        @click="signOut"
      >
        {{ t('auth.signOut') }}
      </KButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { KButton } from '@kong/kongponents'
import { useI18n } from '@/composables/useI18n'
import { useAuthStore } from '@/stores/auth'

defineOptions({ name: 'UserProfileBar' })

const { t } = useI18n()
const router = useRouter()
const authStore = useAuthStore()

const displayName = computed(() => authStore.username ?? '—')

const initials = computed(() => {
  const u = authStore.username
  if (!u) {
    return '?'
  }
  const parts = u.trim().split(/[\s._-]+/).filter(Boolean)
  if (parts.length >= 2 && parts[0]?.[0] && parts[1]?.[0]) {
    return `${parts[0][0]}${parts[1][0]}`.toUpperCase()
  }
  return u.slice(0, 2).toUpperCase()
})

async function signOut() {
  authStore.clearSession()
  await router.replace({ name: 'login' })
}
</script>

<style scoped lang="scss">
.user-profile-bar {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  max-width: 280px;
}

.user-profile-bar__avatar {
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--kui-color-background-brand, #1155cb);
  color: #fff;
  font-size: 0.75rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  letter-spacing: 0.02em;
}

.user-profile-bar__meta {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0;
  min-width: 0;
}

.user-profile-bar__name {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--kui-color-text, #1a1a1a);
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-profile-bar__signout {
  padding: 0 !important;
  min-height: auto !important;
  font-size: 0.75rem !important;
}
</style>
