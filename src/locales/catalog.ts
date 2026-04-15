import { cloneDeep, merge } from 'lodash-es'
import { createI18n } from '@kong-ui-public/i18n'
import { ref } from 'vue'
import en from '@/locales/en.json'
import viPartial from '@/locales/overrides/vi.json'
import zhCNPartial from '@/locales/overrides/zh-CN.json'
import jaJPPartial from '@/locales/overrides/ja-JP.json'
import koKRPartial from '@/locales/overrides/ko-KR.json'
import frFRPartial from '@/locales/overrides/fr-FR.json'
import deDEPartial from '@/locales/overrides/de-DE.json'
import esESPartial from '@/locales/overrides/es-ES.json'

export type MessageSource = typeof en

export const LOCALE_CODES = [
  'en-us',
  'vi-vn',
  'zh-cn',
  'ja-jp',
  'ko-kr',
  'fr-fr',
  'de-de',
  'es-es',
] as const

export type LocaleCode = (typeof LOCALE_CODES)[number]

export const DEFAULT_LOCALE: LocaleCode = 'en-us'

/** Native labels for the language switcher (conventional names, not translated). */
export const LOCALE_LABELS: Record<LocaleCode, string> = {
  'en-us': 'English',
  'vi-vn': 'Tiếng Việt',
  'zh-cn': '简体中文',
  'ja-jp': '日本語',
  'ko-kr': '한국어',
  'fr-fr': 'Français',
  'de-de': 'Deutsch',
  'es-es': 'Español',
}

/** Synced with Pinia locale store — used outside Vue setup (e.g. axios). */
export const currentLocaleCode = ref<LocaleCode>(DEFAULT_LOCALE)

const STORAGE_KEY = 'km_ui_locale'

function buildLocale(partial: Record<string, unknown>): MessageSource {
  return merge(cloneDeep(en), partial) as MessageSource
}

export const catalogs: Record<LocaleCode, MessageSource> = {
  'en-us': en,
  'vi-vn': buildLocale(viPartial as Record<string, unknown>),
  'zh-cn': buildLocale(zhCNPartial as Record<string, unknown>),
  'ja-jp': buildLocale(jaJPPartial as Record<string, unknown>),
  'ko-kr': buildLocale(koKRPartial as Record<string, unknown>),
  'fr-fr': buildLocale(frFRPartial as Record<string, unknown>),
  'de-de': buildLocale(deDEPartial as Record<string, unknown>),
  'es-es': buildLocale(esESPartial as Record<string, unknown>),
}

function normalizeBrowserLang(nav: string): LocaleCode | null {
  const n = nav.toLowerCase().replace('_', '-')
  if (n.startsWith('vi')) {
    return 'vi-vn'
  }
  if (n.startsWith('zh')) {
    return 'zh-cn'
  }
  if (n.startsWith('ja')) {
    return 'ja-jp'
  }
  if (n.startsWith('ko')) {
    return 'ko-kr'
  }
  if (n.startsWith('fr')) {
    return 'fr-fr'
  }
  if (n.startsWith('de')) {
    return 'de-de'
  }
  if (n.startsWith('es')) {
    return 'es-es'
  }
  if (n.startsWith('en')) {
    return 'en-us'
  }
  return null
}

export function resolveInitialLocale(): LocaleCode {
  if (typeof localStorage !== 'undefined') {
    const saved = localStorage.getItem(STORAGE_KEY)
    if (saved && LOCALE_CODES.includes(saved as LocaleCode)) {
      return saved as LocaleCode
    }
  }
  if (typeof navigator !== 'undefined') {
    const fromNav = normalizeBrowserLang(navigator.language)
    if (fromNav) {
      return fromNav
    }
    for (const lang of navigator.languages ?? []) {
      const m = normalizeBrowserLang(lang)
      if (m) {
        return m
      }
    }
  }
  return DEFAULT_LOCALE
}

export function persistLocale(code: LocaleCode) {
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem(STORAGE_KEY, code)
  }
}

/** (Re)apply global Kong i18n instance — call after changing locale. */
export function applyI18nLocale(code: LocaleCode) {
  const i18n = createI18n(code as never, catalogs[code], { isGlobal: true })
  currentLocaleCode.value = code
  persistLocale(code)
  if (typeof document !== 'undefined') {
    document.documentElement.lang = code.split('-')[0] || 'en'
  }
  return i18n
}

export function getPermissionDeniedMessage(): string {
  return catalogs[currentLocaleCode.value].global.permissionDenied
}
