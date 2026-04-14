import type { App } from 'vue'
import type { AxiosRequestConfig } from 'axios'
import axios from 'axios'

/** Must match @kong-ui-public/entities-shared (see getAxiosInstance → inject). */
const GET_AXIOS_INSTANCE = 'get-axios-instance'

export const KM_TOKEN_KEY = 'km_token'

export function setKmToken(token: string | null) {
  if (token) {
    sessionStorage.setItem(KM_TOKEN_KEY, token)
  } else {
    sessionStorage.removeItem(KM_TOKEN_KEY)
  }
}

/**
 * Single axios instance for Kong entity UIs. entities-shared calls useAxios() per component;
 * without app.provide(get-axios-instance), each call used axios.create() with no JWT interceptor.
 */
export const kongAxios = axios.create({
  withCredentials: true,
  timeout: 30_000,
})

kongAxios.interceptors.request.use((reqConfig) => {
  const token = typeof sessionStorage !== 'undefined' ? sessionStorage.getItem(KM_TOKEN_KEY) : null
  if (token) {
    reqConfig.headers.Authorization = `Bearer ${token}`
  }
  return reqConfig
})

/** Wire Kong UI packages to the same axios instance (JWT + same-origin Kong paths). */
export function provideKongAxios(app: App) {
  app.provide(GET_AXIOS_INSTANCE, (_merge?: AxiosRequestConfig) => kongAxios)
}
