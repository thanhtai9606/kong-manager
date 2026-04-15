<template>
  <div
    class="km-app-root"
    :key="localeStore.remountKey"
  >
    <div
      v-if="isAdminShell"
      class="admin-root"
    >
      <router-view />
    </div>
    <AppLayout
      v-else-if="showAppShell"
      :sidebar-top-items="sidebarItems"
    >
      <template #navbar-right>
        <KongClusterSwitch />
        <LanguageSwitcher />
        <UserProfileBar v-if="config.AUTH_REQUIRED" />
        <GithubStar
          v-else
          url="https://github.com/kong/kong"
        />
      </template>
      <template #sidebar-header>
        <NavbarLogo />
      </template>
      <router-view :key="clusterRouteKey" />
      <MakeAWish />
    </AppLayout>
    <div
      v-else
      class="auth-only-view"
    >
      <LanguageSwitcher class="auth-only-view__locale" />
      <router-view />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { storeToRefs } from 'pinia'
import { AppLayout, type SidebarPrimaryItem } from '@kong-ui-public/app-layout'
import { GithubStar } from '@kong-ui-public/misc-widgets'
import { config } from 'config'
import { useAuthStore } from '@/stores/auth'
import { useInfoStore } from '@/stores/info'
import { useKongClusterStore } from '@/stores/kongCluster'
import { useLocaleStore } from '@/stores/locale'
import NavbarLogo from '@/components/NavbarLogo.vue'
import KongClusterSwitch from '@/components/KongClusterSwitch.vue'
import LanguageSwitcher from '@/components/LanguageSwitcher.vue'
import MakeAWish from '@/components/MakeAWish.vue'
import UserProfileBar from '@/components/UserProfileBar.vue'
import { useI18n } from '@/composables/useI18n'

const route = useRoute()
const { t } = useI18n()
const localeStore = useLocaleStore()
const authStore = useAuthStore()
const infoStore = useInfoStore()
const kongClusterStore = useKongClusterStore()
const { isHybridMode } = storeToRefs(infoStore)

/** Remount gateway pages when Kong cluster changes so lists/detail refetch against the new admin base URL. */
const clusterRouteKey = computed(
  () => `${kongClusterStore.selectedSlug}::${route.fullPath}`,
)

async function loadClustersIfAuth() {
  if (!config.AUTH_REQUIRED || !authStore.isAuthenticated) {
    return
  }
  await kongClusterStore.loadClusters()
  await infoStore.getInfo({ silent: true, force: true })
}

onMounted(() => {
  void loadClustersIfAuth()
})

watch(
  () => authStore.isAuthenticated,
  (ok) => {
    if (ok && config.AUTH_REQUIRED) {
      void loadClustersIfAuth()
    }
  },
)

const isAdminShell = computed(() =>
  route.matched.some((r) => r.meta?.shell === 'admin'),
)

/** With BFF auth, unauthenticated users only see the login route (no sidebar/header). */
const showAppShell = computed(() => {
  if (!config.AUTH_REQUIRED) {
    return true
  }
  return authStore.isAuthenticated
})

const sidebarItems = computed<SidebarPrimaryItem[]>(() => [
  {
    name: t('app.nav.overview'),
    to: { name: 'overview' },
    key: 'Overview',
    active: route.name === 'overview',
  },
  {
    name: t('app.nav.gatewayServices'),
    to: { name: 'service-list' },
    key: 'Gateway Services',
    active: route.meta?.entity === 'service',
  },
  {
    name: t('app.nav.routes'),
    to: { name: 'route-list' },
    key: 'Routes',
    active: route.meta?.entity === 'route',
  },
  {
    name: t('app.nav.consumers'),
    to: { name: 'consumer-list' },
    key: 'Consumers',
    active: route.meta?.entity === 'consumer',
  },
  {
    name: t('app.nav.plugins'),
    to: { name: 'plugin-list' },
    key: 'Plugins',
    active: route.meta?.entity === 'plugin',
  },
  {
    name: t('app.nav.upstreams'),
    to: { name: 'upstream-list' },
    key: 'Upstreams',
    active: route.meta?.entity === 'upstream',
  },
  {
    name: t('app.nav.certificates'),
    to: { name: 'certificate-list' },
    key: 'Certificates',
    active: route.meta?.entity === 'certificate',
  },
  {
    name: t('app.nav.caCertificates'),
    to: { name: 'ca-certificate-list' },
    key: 'CA Certificates',
    active: route.meta?.entity === 'ca-certificate',
  },
  {
    name: t('app.nav.snis'),
    to: { name: 'sni-list' },
    key: 'SNIs',
    active: route.meta?.entity === 'sni',
  },
  {
    name: t('app.nav.vaults'),
    to: { name: 'vault-list' },
    key: 'Vaults',
    active: route.meta?.entity === 'vault',
  },
  {
    name: t('app.nav.keys'),
    to: { name: 'key-list' },
    key: 'Keys',
    active: route.meta?.entity === 'key',
  },
  {
    name: t('app.nav.keySets'),
    to: { name: 'key-set-list' },
    key: 'Key Sets',
    active: route.meta?.entity === 'key-set',
  },
  ...(
    isHybridMode.value
      ? [
        // {
        //   name: 'Data Plane Nodes',
        //   to: { name: 'data-plane-nodes' },
        //   key: 'Data Plane Nodes',
        //   active: route.meta?.entity === 'data-plane-node',
        // },
      ]
      : []
  ),
])
</script>

<style scoped lang="scss">
.app-title {
  color: #fff;
  margin: 0;
  font-size: 20px;
}

:deep(.kong-ui-app-layout-content-inner) {
  position: relative;
  min-height: 100%;
  padding: 32px 40px 80px !important;
}

:deep(.json-content.k-code-block) {
  border-top-left-radius: $kui-border-radius-0 !important;
  border-top-right-radius: $kui-border-radius-0 !important;
}

.auth-only-view {
  position: relative;
  min-height: 100vh;
  background: var(--kui-color-background, #f7f7f7);
}

.auth-only-view__locale {
  position: absolute;
  top: 1rem;
  right: 1rem;
  z-index: 2;
}

.admin-root {
  min-height: 100vh;
}
</style>
