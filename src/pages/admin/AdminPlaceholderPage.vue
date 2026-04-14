<template>
  <PageHeader :title="pageTitle" />
  <p class="admin-placeholder">
    {{ t('admin.placeholder.description') }}
  </p>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from '@/composables/useI18n'

defineOptions({ name: 'AdminPlaceholderPage' })

const route = useRoute()
const { t } = useI18n()

const pageKey = computed(() => (route.meta.adminPageKey as string | undefined) ?? 'home')

const pageTitle = computed(() => {
  switch (pageKey.value) {
    case 'users':
      return t('admin.pages.users.title')
    case 'rbac':
      return t('admin.pages.rbac.title')
    case 'clusters':
      return t('admin.pages.clusters.title')
    case 'notifications':
      return t('admin.pages.notifications.title')
    case 'home':
    default:
      return t('admin.pages.home.title')
  }
})
</script>

<style scoped lang="scss">
.admin-placeholder {
  margin-top: 1rem;
  color: var(--kui-color-text-neutral, #525252);
  max-width: 42rem;
  line-height: 1.5;
}
</style>
