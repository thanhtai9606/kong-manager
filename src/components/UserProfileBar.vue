<template>
  <div
    class="user-profile-bar"
    data-testid="user-profile-bar"
  >
    <KDropdown>
      <KButton
        appearance="tertiary"
        class="user-profile-bar__trigger"
        data-testid="user-profile-menu-trigger"
      >
        <span
          class="user-profile-bar__avatar"
          aria-hidden="true"
        >{{ initials }}</span>
        <span class="user-profile-bar__name">{{ displayName }}</span>
        <ChevronDownIcon class="user-profile-bar__chevron" />
      </KButton>
      <template #items>
        <KDropdownItem
          v-if="config.AUTH_REQUIRED"
          data-testid="user-menu-admin"
          @click="goAdmin"
        >
          {{ t('auth.menu.admin') }}
        </KDropdownItem>
        <KDropdownItem
          data-testid="user-sign-out"
          @click="signOut"
        >
          {{ t('auth.menu.signOut') }}
        </KDropdownItem>
      </template>
    </KDropdown>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { KButton, KDropdown, KDropdownItem } from '@kong/kongponents'
import { ChevronDownIcon } from '@kong/icons'
import { config } from 'config'
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

function goAdmin() {
  void router.push({ name: 'admin-home' })
}

async function signOut() {
  authStore.clearSession()
  await router.replace({ name: 'login' })
}
</script>

<style scoped lang="scss">
.user-profile-bar {
  display: flex;
  align-items: center;
  max-width: 320px;
}

.user-profile-bar__trigger {
  display: inline-flex !important;
  align-items: center;
  gap: 0.5rem;
  max-width: 100%;
  padding-inline: 0.35rem 0.25rem !important;
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
  display: inline-flex;
  align-items: center;
  justify-content: center;
  letter-spacing: 0.02em;
}

.user-profile-bar__name {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--kui-color-text, #1a1a1a);
  min-width: 0;
  max-width: 11rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-profile-bar__chevron {
  flex-shrink: 0;
  width: 1rem;
  height: 1rem;
}
</style>
