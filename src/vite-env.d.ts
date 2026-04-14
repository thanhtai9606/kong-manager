/// <reference types="vite/client" />

interface ImportMetaEnv {
  /**
   * Set at build time (e.g. `VITE_AUTH_REQUIRED=true pnpm build`) to require login
   * when `kconfig.js` is missing or does not set AUTH_REQUIRED.
   */
  readonly VITE_AUTH_REQUIRED?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
