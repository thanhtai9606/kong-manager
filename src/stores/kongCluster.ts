import { defineStore } from 'pinia'
import { ref } from 'vue'
import { apiService, setActiveClusterSlug } from '@/services/apiService'

export interface KongClusterRow {
  id: number
  name: string
  slug: string
  admin_base_url: string
  has_token: boolean
  enabled: boolean
  sort_order: number
}

function readStoredSlug(): string {
  if (typeof sessionStorage === 'undefined') {
    return 'default'
  }
  return sessionStorage.getItem('km_kong_cluster_slug') ?? 'default'
}

export const useKongClusterStore = defineStore('kongCluster', () => {
  const clusters = ref<KongClusterRow[]>([])
  const loaded = ref(false)
  const loading = ref(false)
  const selectedSlug = ref(readStoredSlug())
  setActiveClusterSlug(selectedSlug.value)

  async function loadClusters() {
    loading.value = true
    try {
      const { data } = await apiService.bffGet<KongClusterRow[]>('/api/admin/kong-clusters')
      clusters.value = Array.isArray(data) ? data : []
      loaded.value = true
      const enabled = clusters.value.filter((c) => c.enabled)
      const want = readStoredSlug()
      if (want && enabled.some((c) => c.slug === want)) {
        selectSlug(want)
        return
      }
      if (enabled.length > 0) {
        selectSlug(enabled[0]!.slug)
      } else {
        selectSlug('default')
      }
    } finally {
      loading.value = false
    }
  }

  function selectSlug(slug: string) {
    selectedSlug.value = slug
    setActiveClusterSlug(slug)
  }

  return { clusters, loaded, loading, selectedSlug, loadClusters, selectSlug }
})
