<template>
  <AppLayout :sidebar-top-items="sidebarItems">
    <template #navbar-right>
      <KongClusterSwitch />
      <UserProfileBar v-if="config.AUTH_REQUIRED" />
    </template>
    <template #sidebar-header>
      <NavbarLogo />
    </template>
    <router-view />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { AppLayout, type SidebarPrimaryItem } from '@kong-ui-public/app-layout'
import { config } from 'config'
import NavbarLogo from '@/components/NavbarLogo.vue'
import KongClusterSwitch from '@/components/KongClusterSwitch.vue'
import UserProfileBar from '@/components/UserProfileBar.vue'
import { useI18n } from '@/composables/useI18n'

defineOptions({ name: 'AdminLayout' })

const route = useRoute()
const { t } = useI18n()

const sidebarItems = computed<SidebarPrimaryItem[]>(() => [
  {
    name: t('admin.nav.backToGateway'),
    to: { name: 'overview' },
    key: 'admin-back-gateway',
    active: false,
  },
  {
    name: t('admin.nav.home'),
    to: { name: 'admin-home' },
    key: 'admin-home-nav',
    active: route.name === 'admin-home',
  },
  {
    name: t('admin.nav.users'),
    to: { name: 'admin-users' },
    key: 'admin-users',
    active: route.name === 'admin-users',
  },
  {
    name: t('admin.nav.rbac'),
    to: { name: 'admin-rbac' },
    key: 'admin-rbac',
    active: route.name === 'admin-rbac',
  },
  {
    name: t('admin.nav.clusters'),
    to: { name: 'admin-clusters' },
    key: 'admin-clusters',
    active: route.name === 'admin-clusters',
  },
  {
    name: t('admin.nav.notifications'),
    to: { name: 'admin-notifications' },
    key: 'admin-notifications',
    active: route.name === 'admin-notifications',
  },
])
</script>

<style scoped lang="scss">
:deep(.kong-ui-app-layout-content-inner) {
  position: relative;
  min-height: 100%;
  padding: 32px 40px 80px !important;
}
</style>
