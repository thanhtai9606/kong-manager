import { reactive } from 'vue'
import { kongAdminUrlForSlug } from '@/services/apiService'
import { useInfoStore } from '@/stores/info'
import { useKongClusterStore } from '@/stores/kongCluster'
import type { KongManagerConfig } from '@kong-ui-public/entities-shared'

export const useBaseGeneralConfig = () => {
  const infoStore = useInfoStore()
  const kongClusterStore = useKongClusterStore()
  return reactive({
    app: 'kongManager' as const,
    workspace: '',
    gatewayInfo: {
      edition: infoStore.kongEdition,
      version: infoStore.kongVersion,
    },
    get apiBaseUrl() {
      void kongClusterStore.selectedSlug
      return kongAdminUrlForSlug(kongClusterStore.selectedSlug || 'default')
    },
  }) as KongManagerConfig
}
