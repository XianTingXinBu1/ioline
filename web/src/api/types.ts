/**
 * API 类型定义 — 后续随后端接口扩展
 */

/** 健康检查响应 */
export interface HealthzResponse {
  status: string
}

/** 文件信息 */
export interface FileInfo {
  name: string
  path: string
  is_dir: boolean
  size: number
  modified: string
}
