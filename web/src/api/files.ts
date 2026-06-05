import { api } from './client'
import type {
  FileContentResponse,
  FileListResponse,
  SaveFileContentRequest,
  SaveFileContentResponse,
} from './types'

export const filesApi = {
  list(path = '.'): Promise<FileListResponse> {
    const params = new URLSearchParams({ path })
    return api.get(`/files/list?${params.toString()}`)
  },

  getContent(path: string): Promise<FileContentResponse> {
    const params = new URLSearchParams({ path })
    return api.get(`/file/content?${params.toString()}`)
  },

  saveContent(payload: SaveFileContentRequest): Promise<SaveFileContentResponse> {
    return api.put('/file/content', payload)
  },
}
