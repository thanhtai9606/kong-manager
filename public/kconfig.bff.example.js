/**
 * Copy to `kconfig.js` (see .gitignore) or rely on `VITE_AUTH_REQUIRED=true` at build time.
 * Served as static `/kconfig.js` after `pnpm build` when the file exists in public/.
 *
 * With BFF + AUTH_REQUIRED, the SPA builds Kong API URLs as /kong-admin/c/{slug} (cluster switcher).
 * Kong Admin upstream is chosen on the server (KONG_ADMIN_URL) and per row in Admin → Kong clusters.
 */
window.K_CONFIG = {
  AUTH_REQUIRED: true,
  ADMIN_API_URL: '/kong-admin',
}
