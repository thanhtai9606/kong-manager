/**
 * Copy to `kconfig.js` (see .gitignore) or rely on `VITE_AUTH_REQUIRED=true` at build time.
 * Served as static `/kconfig.js` after `pnpm build` when the file exists in public/.
 */
window.K_CONFIG = {
  AUTH_REQUIRED: true,
  ADMIN_API_URL: '/kong-admin',
}
