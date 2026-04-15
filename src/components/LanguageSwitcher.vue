<template>
  <div
    class="language-switcher"
    data-testid="language-switcher"
  >
    <span class="language-switcher__label">{{ t('global.language') }}</span>
    <select
      class="language-switcher__select"
      :value="currentLocaleCode"
      :aria-label="t('global.language')"
      data-testid="language-switcher-select"
      @change="onChange"
    >
      <option
        v-for="code in LOCALE_CODES"
        :key="code"
        :value="code"
      >
        {{ LOCALE_LABELS[code] }}
      </option>
    </select>
  </div>
</template>

<script setup lang="ts">
import {
  LOCALE_CODES,
  LOCALE_LABELS,
  currentLocaleCode,
  type LocaleCode,
} from '@/locales/catalog'
import { useI18n } from '@/composables/useI18n'
import { useLocaleStore } from '@/stores/locale'

defineOptions({ name: 'LanguageSwitcher' })

const { t } = useI18n()
const localeStore = useLocaleStore()

function onChange(ev: Event) {
  const code = (ev.target as HTMLSelectElement).value as LocaleCode
  localeStore.setLocale(code)
}
</script>

<style scoped lang="scss">
.language-switcher {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-right: 0.75rem;
}

.language-switcher__label {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--kui-color-text-neutral, #525252);
  white-space: nowrap;
}

.language-switcher__select {
  min-width: 9rem;
  max-width: 14rem;
  font-size: 0.875rem;
  padding: 0.35rem 0.5rem;
  border-radius: var(--kui-border-radius-20, 4px);
  border: 1px solid var(--kui-color-border-neutral, #ccc);
  background: var(--kui-color-background, #fff);
}
</style>
