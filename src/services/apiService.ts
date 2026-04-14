import {
  type AxiosInstance,
  type AxiosRequestConfig,
} from 'axios'
import { config } from 'config'
import { kongAxios } from '@/services/kongAxios'

const KM_CLUSTER_SLUG_KEY = 'km_kong_cluster_slug'

function storedClusterSlug(): string {
  if (typeof sessionStorage === 'undefined') {
    return 'default'
  }
  return sessionStorage.getItem(KM_CLUSTER_SLUG_KEY) ?? 'default'
}

let activeClusterSlug = storedClusterSlug()

/**
 * Browser → BFF path for Kong Admin (same origin). Slug selects which DB row the BFF proxies to.
 * The real Kong Admin host:port lives in kong_clusters.admin_base_url, not in this string.
 */
export function kongAdminUrlForSlug(slug: string): string {
  if (!config.AUTH_REQUIRED) {
    return config.ADMIN_API_URL
  }
  const gui = config.ADMIN_GUI_PATH.replace(/\/$/, '') || ''
  const origin = typeof window !== 'undefined' ? window.location.origin : ''
  const s = slug || 'default'
  return `${origin}${gui}/kong-admin/c/${s}`
}

/** Called when user switches Kong cluster (avoid Pinia↔apiService circular imports). */
export function setActiveClusterSlug(slug: string) {
  activeClusterSlug = slug || 'default'
  if (typeof sessionStorage !== 'undefined') {
    sessionStorage.setItem(KM_CLUSTER_SLUG_KEY, activeClusterSlug)
  }
}

function adminApiUrl(): string {
  return kongAdminUrlForSlug(activeClusterSlug)
}

/** Same-origin BFF app API (e.g. `/api/admin/*`), not Kong Admin. */
function bffAppUrl(path: string): string {
  const gui = config.ADMIN_GUI_PATH.replace(/\/$/, '') || ''
  const p = path.startsWith('/') ? path : `/${path}`
  if (typeof window === 'undefined') {
    return p
  }
  return `${window.location.origin}${gui}${p}`
}

export { KM_TOKEN_KEY, setKmToken } from '@/services/kongAxios'

class ApiService {
  instance: AxiosInstance

  constructor() {
    this.instance = kongAxios
  }

  getInfo() {
    return this.instance.get(adminApiUrl())
  }

  // entity-specific methods
  findAll<T>(entity: string, params: Record<string, unknown>) {
    return this.instance.get<T>(`${adminApiUrl()}/${entity}`, { params })
  }

  findRecord<T>(entity: string, id: string) {
    return this.instance.get<T>(`${adminApiUrl()}/${entity}/${id}`)
  }

  createRecord(entity: string, data: Record<string, unknown>) {
    return this.instance.post(`${adminApiUrl()}/${entity}`, data)
  }

  updateRecord(entity: string, id: string, data: Record<string, unknown>) {
    return this.instance.patch(`${adminApiUrl()}/${entity}/${id}`, data)
  }

  deleteRecord(entity: string, id: string) {
    return this.instance.delete(`${adminApiUrl()}/${entity}/${id}`)
  }

  // generic methods
  get<T>(url = '', reqConfig: AxiosRequestConfig = {}) {
    return this.instance.get<T>(`${adminApiUrl()}/${url}`, reqConfig)
  }

  post(url = '', data?: Record<string, unknown>, reqConfig: AxiosRequestConfig = {}) {
    return this.instance.post(`${adminApiUrl()}/${url}`, data, reqConfig)
  }

  put(url = '', data?: Record<string, unknown>, reqConfig: AxiosRequestConfig = {}) {
    return this.instance.put(`${adminApiUrl()}/${url}`, data, reqConfig)
  }

  patch(url = '', data?: Record<string, unknown>, reqConfig: AxiosRequestConfig = {}) {
    return this.instance.patch(`${adminApiUrl()}/${url}`, data, reqConfig)
  }

  delete(url = '', reqConfig: AxiosRequestConfig = {}) {
    return this.instance.delete(`${adminApiUrl()}/${url}`, reqConfig)
  }

  /** GET against the Go BFF (JWT + Casbin), not Kong Admin. */
  bffGet<T>(path: string, reqConfig: AxiosRequestConfig = {}) {
    return this.instance.get<T>(bffAppUrl(path), reqConfig)
  }

  bffPost<T>(path: string, data?: Record<string, unknown>, reqConfig: AxiosRequestConfig = {}) {
    return this.instance.post<T>(bffAppUrl(path), data, reqConfig)
  }

  bffPut<T>(path: string, data?: Record<string, unknown>, reqConfig: AxiosRequestConfig = {}) {
    return this.instance.put<T>(bffAppUrl(path), data, reqConfig)
  }

  bffPatch<T>(path: string, data?: Record<string, unknown>, reqConfig: AxiosRequestConfig = {}) {
    return this.instance.patch<T>(bffAppUrl(path), data, reqConfig)
  }

  bffDelete(path: string, reqConfig: AxiosRequestConfig = {}) {
    return this.instance.delete(bffAppUrl(path), reqConfig)
  }
}

export const apiService = new ApiService()
