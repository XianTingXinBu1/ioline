import { ref } from 'vue'
import { api, ApiError } from '@/api/client'
import type { HealthzResponse } from '@/api/types'

/**
 * 通用 API 请求组合式函数
 */
export function useApi() {
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function healthCheck() {
    loading.value = true
    error.value = null
    try {
      const res = await api.get<HealthzResponse>('/healthz')
      return res
    } catch (e) {
      const msg = e instanceof ApiError ? e.message : '网络错误'
      error.value = msg
      throw e
    } finally {
      loading.value = false
    }
  }

  return { loading, error, healthCheck }
}
