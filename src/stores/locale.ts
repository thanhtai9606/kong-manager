import { defineStore } from 'pinia'
import {
  applyI18nLocale,
  LOCALE_CODES,
  type LocaleCode,
} from '@/locales/catalog'

export const useLocaleStore = defineStore('locale', {
  state: () => ({
    /** Bump to force subtree remount after Kong i18n locale swap. */
    remountKey: 0,
  }),
  actions: {
    setLocale(code: LocaleCode) {
      if (!LOCALE_CODES.includes(code)) {
        return
      }
      applyI18nLocale(code)
      this.remountKey++
    },
  },
})
