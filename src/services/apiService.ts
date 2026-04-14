import {
  type AxiosInstance,
  type AxiosRequestConfig,
} from 'axios'
import { config } from 'config'
import { useAxios } from '@kong-ui-public/entities-shared'

function adminApiUrl(): string {
  return config.ADMIN_API_URL
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

/** Session storage key for the Kong Manager BFF JWT. */
export const KM_TOKEN_KEY = 'km_token'

export function setKmToken(token: string | null) {
  if (token) {
    sessionStorage.setItem(KM_TOKEN_KEY, token)
  } else {
    sessionStorage.removeItem(KM_TOKEN_KEY)
  }
}

class ApiService {
  instance: AxiosInstance

  constructor() {
    this.instance = useAxios().axiosInstance
    this.instance.interceptors.request.use((reqConfig) => {
      const token = typeof sessionStorage !== 'undefined' ? sessionStorage.getItem(KM_TOKEN_KEY) : null
      if (token) {
        reqConfig.headers.Authorization = `Bearer ${token}`
      }
      return reqConfig
    })
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
}

export const apiService = new ApiService()
