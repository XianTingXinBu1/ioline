/**
 * API 类型定义
 */

export interface ApiErrorPayload {
  code: string
  message: string
}

export interface ApiEnvelope<T> {
  success: boolean
  data: T
  error?: ApiErrorPayload
}

export interface HealthzResponse {
  status: string
}

export interface SystemInfoResponse {
  name: string
  goVersion: string
  os: string
  arch: string
  termux: boolean
  workspaceSet: boolean
  terminalMaxSessions: number
}

export interface WorkspaceCurrentResponse {
  rootPath?: string
  name?: string
  isSet: boolean
  setAt?: string
}

export interface WorkspaceCandidateItem {
  name: string
  path: string
  exists: boolean
  source: string
}

export interface WorkspaceCandidatesResponse {
  items: WorkspaceCandidateItem[]
}

export interface SetWorkspaceRequest {
  rootPath: string
}

export interface FileListItemResponse {
  name: string
  path: string
  type: 'file' | 'directory'
  size: number
  modifiedAt: string
  readonly: boolean
  hidden: boolean
}

export interface FileListResponse {
  items: FileListItemResponse[]
}

export interface FileContentResponse {
  path: string
  content: string
  size: number
  modifiedAt: string
  readonly: boolean
  binary: boolean
  lineEnding: 'lf' | 'crlf'
}

export interface SaveFileContentRequest {
  path: string
  content: string
}

export interface SaveFileContentResponse {
  path: string
  size: number
  modifiedAt: string
}
