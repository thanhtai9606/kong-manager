declare global {
  interface Window {
    K_CONFIG: Record<string, string>
  }
}

declare module 'vue-router' {
  interface RouteMeta {
    shell?: 'admin'
    public?: boolean
    adminPageKey?: string
  }
}

export {}
