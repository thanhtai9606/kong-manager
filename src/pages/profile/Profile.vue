<template>
  <PageHeader :title="t('profile.title')">
    <template #below-title>
      <SupportText>
        {{ t('profile.description') }}
      </SupportText>
    </template>
  </PageHeader>

  <KCard>
    <div
      v-if="!authStore.isAuthenticated"
      class="profile__empty"
    >
      {{ t('profile.notSignedIn') }}
    </div>
    <template v-else>
      <dl class="profile__dl">
        <div
          v-for="row in tokenRows"
          :key="row.key"
          class="profile__row"
        >
          <dt class="profile__dt">
            {{ row.label }}
          </dt>
          <dd class="profile__dd">
            {{ row.value }}
          </dd>
        </div>
        <div
          v-if="me?.email"
          class="profile__row"
        >
          <dt class="profile__dt">
            {{ t('profile.fields.email') }}
          </dt>
          <dd class="profile__dd">
            {{ me.email }}
          </dd>
        </div>
        <div class="profile__row profile__row--roles">
          <dt class="profile__dt">
            {{ t('profile.fields.roles') }}
          </dt>
          <dd class="profile__dd profile__roles">
            <span v-if="meLoading">{{ t('profile.loading') }}</span>
            <span
              v-else-if="meError"
              class="profile__roles-error"
            >{{ meError }}</span>
            <template v-else-if="me?.groups?.length">
              <KBadge
                v-for="g in me.groups"
                :key="g.id"
                class="profile__badge"
                appearance="info"
              >
                {{ g.name }}
              </KBadge>
            </template>
            <span v-else-if="me">—</span>
            <span v-else>—</span>
          </dd>
        </div>
      </dl>
      <div class="profile__actions">
        <KButton
          v-if="config.AUTH_REQUIRED"
          appearance="primary"
          @click="goAdmin"
        >
          {{ t('profile.openAdmin') }}
        </KButton>
        <KButton
          appearance="danger"
          @click="signOut"
        >
          {{ t('auth.menu.signOut') }}
        </KButton>
      </div>
    </template>
  </KCard>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import type { AxiosError } from 'axios'
import { useRouter } from 'vue-router'
import { KBadge, KButton, KCard } from '@kong/kongponents'
import dayjs from 'dayjs'
import PageHeader from '@/components/PageHeader.vue'
import SupportText from '@/components/SupportText.vue'
import { config } from 'config'
import { useI18n } from '@/composables/useI18n'
import { apiService } from '@/services/apiService'
import { useAuthStore } from '@/stores/auth'
import { jwtExpiresAt, jwtSubject, jwtUserId } from '@/utils/jwt'

defineOptions({ name: 'ProfilePage' })

const { t } = useI18n()
const router = useRouter()
const authStore = useAuthStore()

type MeGroup = { id: number, name: string }
type MeResponse = {
  id: number
  username: string
  email?: string
  groups: MeGroup[]
}

const me = ref<MeResponse | null>(null)
const meLoading = ref(false)
const meError = ref('')

const tokenRows = computed(() => {
  const token = authStore.token
  const username = jwtSubject(token)
  const uid = jwtUserId(token)
  const exp = jwtExpiresAt(token)
  const out: Array<{ key: string, label: string, value: string }> = []
  if (username) {
    out.push({
      key: 'username',
      label: t('profile.fields.username'),
      value: username,
    })
  }
  if (uid != null) {
    out.push({
      key: 'uid',
      label: t('profile.fields.userId'),
      value: String(uid),
    })
  }
  if (exp) {
    out.push({
      key: 'exp',
      label: t('profile.fields.sessionExpires'),
      value: dayjs(exp).format('YYYY-MM-DD HH:mm:ss Z'),
    })
  }
  if (out.length === 0) {
    out.push({
      key: 'token',
      label: t('profile.fields.session'),
      value: t('profile.sessionUnknown'),
    })
  }
  return out
})

async function loadMe() {
  if (!authStore.isAuthenticated || !authStore.token) {
    me.value = null
    meError.value = ''
    return
  }
  meLoading.value = true
  meError.value = ''
  try {
    const { data } = await apiService.bffGet<MeResponse>('/api/auth/me')
    me.value = data
  } catch (e) {
    me.value = null
    const err = e as AxiosError
    meError.value = err.response?.status === 401
      ? t('profile.loadErrorAuth')
      : t('profile.loadError')
  } finally {
    meLoading.value = false
  }
}

onMounted(() => {
  void loadMe()
})

watch(
  () => authStore.token,
  () => {
    void loadMe()
  },
)

function goAdmin() {
  void router.push({ name: 'admin-users' })
}

async function signOut() {
  authStore.clearSession()
  me.value = null
  await router.replace({ name: 'login' })
}
</script>

<style scoped lang="scss">
.profile__empty {
  margin: 0;
  font-size: 0.875rem;
  color: var(--kui-color-text-neutral, #525252);
}

.profile__dl {
  margin: 0 0 1.25rem;
}

.profile__row {
  display: grid;
  grid-template-columns: minmax(8rem, 14rem) 1fr;
  gap: 0.5rem 1.5rem;
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--kui-color-border-neutral, #e7e7e7);
  font-size: 0.875rem;

  &:last-of-type {
    border-bottom: none;
  }
}

.profile__dt {
  margin: 0;
  font-weight: 600;
  color: var(--kui-color-text-neutral, #525252);
}

.profile__dd {
  margin: 0;
  word-break: break-word;
}

.profile__row--roles {
  align-items: flex-start;
}

.profile__roles {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  align-items: center;
}

.profile__badge {
  margin: 0;
}

.profile__roles-error {
  color: var(--kui-color-text-danger, #c20d0d);
  font-size: 0.8125rem;
}

.profile__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}
</style>
