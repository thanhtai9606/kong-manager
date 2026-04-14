<template>
  <div
    v-if="config.AUTH_REQUIRED"
    class="kong-cluster-switch"
  >
    <span class="kong-cluster-switch__label">{{ t('admin.clusters.switchLabel') }}</span>
    <select
      class="kong-cluster-switch__select"
      :value="store.selectedSlug"
      :disabled="store.loading || enabledClusters.length === 0"
      :aria-label="t('admin.clusters.switchLabel')"
      @change="onChange"
    >
      <option
        v-for="c in enabledClusters"
        :key="c.id"
        :value="c.slug"
      >
        {{ c.name }} ({{ c.slug }})
      </option>
    </select>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { config } from 'config'
import { useI18n } from '@/composables/useI18n'
import { useInfoStore } from '@/stores/info'
import { useKongClusterStore } from '@/stores/kongCluster'

defineOptions({ name: 'KongClusterSwitch' })

const { t } = useI18n()
const store = useKongClusterStore()
const infoStore = useInfoStore()

const enabledClusters = computed(() => store.clusters.filter((c) => c.enabled))

function onChange(ev: Event) {
  const slug = (ev.target as HTMLSelectElement).value
  store.selectSlug(slug)
  void infoStore.getInfo({ silent: true })
}
</script>

<style scoped lang="scss">
.kong-cluster-switch {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-right: 0.75rem;
}

.kong-cluster-switch__label {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--kui-color-text-neutral, #525252);
  white-space: nowrap;
}

.kong-cluster-switch__select {
  min-width: 10rem;
  max-width: 18rem;
  font-size: 0.875rem;
  padding: 0.35rem 0.5rem;
  border-radius: var(--kui-border-radius-20, 4px);
  border: 1px solid var(--kui-color-border-neutral, #ccc);
  background: var(--kui-color-background, #fff);
}
</style>
