import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'
import autoprefixer from 'autoprefixer'
import { createHtmlPlugin } from 'vite-plugin-html'
import { visualizer } from 'rollup-plugin-visualizer'

const basePath = process.env.NODE_ENV !== 'production' || process.env.DISABLE_BASE_PATH === 'true' ? '/' : '/__km_base__/'

/** Go BFF (login + Kong Admin proxy). Vite runs on :8080 by default; run the binary on another port. */
const kongManagerBff = process.env.KONG_MANAGER_BFF_URL || 'http://127.0.0.1:8081'

// https://vitejs.dev/config/
export default defineConfig({
  base: basePath,
  resolve: {
    alias: [
      {
        find: '@',
        replacement: path.resolve(__dirname, 'src'),
      },
      {
        find: 'config',
        replacement: path.resolve(__dirname, 'src/config'),
      },
    ],
  },
  plugins: [
    vue(),
    createHtmlPlugin({
      minify: false,
      inject: {
        data: {
          basePath,
        },
      },
    }),
    visualizer({
      filename: path.resolve(__dirname, 'bundle-analyzer/stats-treemap.html'),
      template: 'treemap', // sunburst|treemap|network
      gzipSize: true,
      brotliSize: true,
    }),
  ],
  server: {
    proxy: {
      // changeOrigin: false keeps the browser Host (e.g. localhost:8080) when forwarding to the BFF.
      // If true, Host becomes 127.0.0.1:8081 and OIDC redirect_uri no longer matches what you register in Keycloak.
      '/api': { target: kongManagerBff, changeOrigin: false },
      '/kong-admin': { target: kongManagerBff, changeOrigin: false },
      '/kconfig.js': process.env.KONG_GUI_URL || 'http://127.0.0.1:8002',
    },
    port: 8080,
  },
  preview: {
    port: 8080,
  },
  css: {
    preprocessorOptions: {
      scss: {
        additionalData: '@use "@kong/design-tokens/tokens/scss/variables" as *;',
      },
    },
    postcss: {
      plugins: [
        // @ts-ignore
        autoprefixer,
      ],
    },
  },
})
