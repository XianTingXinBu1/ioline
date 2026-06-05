/**
 * HTTP API 客户端封装
 * 统一处理后端 success/data/error 响应结构
 */
import type { ApiEnvelope, ApiErrorPayload } from './types'

const BASE_URL = '/api'

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
    public code?: string,
    public payload?: ApiErrorPayload,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

async function parseErrorResponse(res: Response): Promise<ApiError> {
  const text = await res.text().catch(() => '')

  if (!text) {
    return new ApiError(res.status, res.statusText || '请求失败')
  }

  try {
    const parsed = JSON.parse(text) as ApiEnvelope<unknown>
    const message = parsed.error?.message || res.statusText || '请求失败'
    return new ApiError(res.status, message, parsed.error?.code, parsed.error)
  } catch {
    return new ApiError(res.status, text || res.statusText || '请求失败')
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const url = `${BASE_URL}${path}`

  const res = await fetch(url, {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    ...options,
  })

  if (!res.ok) {
    throw await parseErrorResponse(res)
  }

  if (res.status === 204) return undefined as T

  const envelope = await res.json() as ApiEnvelope<T>

  if (!envelope.success) {
    throw new ApiError(
      res.status,
      envelope.error?.message || '请求失败',
      envelope.error?.code,
      envelope.error,
    )
  }

  return envelope.data
}

export const api = {
  get<T>(path: string): Promise<T> {
    return request<T>(path, { method: 'GET' })
  },

  post<T>(path: string, body?: unknown): Promise<T> {
    return request<T>(path, {
      method: 'POST',
      body: body ? JSON.stringify(body) : undefined,
    })
  },

  put<T>(path: string, body?: unknown): Promise<T> {
    return request<T>(path, {
      method: 'PUT',
      body: body ? JSON.stringify(body) : undefined,
    })
  },

  delete<T>(path: string): Promise<T> {
    return request<T>(path, { method: 'DELETE' })
  },
}
