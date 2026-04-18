<template>
  <div class="brand-logo">
    <router-link
      :to="{ name: 'overview' }"
      :title="t('navbar.overviewLinkTitle')"
    >
      <img
        src="@/assets/logo.svg?external"
        :alt="t('navbar.logoAlt')"
      >
    </router-link>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'

const { t } = useI18n()
</script>
