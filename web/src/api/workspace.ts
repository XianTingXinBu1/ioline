import { api } from './client'
import type { WorkspaceCurrentResponse, WorkspaceDirectoriesResponse, SetWorkspaceRequest } from './types'

export const workspaceApi = {
  getCurrent(): Promise<WorkspaceCurrentResponse> {
    return api.get('/workspace/current')
  },

  setCurrent(payload: SetWorkspaceRequest): Promise<WorkspaceCurrentResponse> {
    return api.put('/workspace/current', payload)
  },

  clearCurrent(): Promise<WorkspaceCurrentResponse> {
    return api.delete('/workspace/current')
  },

  getDirectories(path?: string): Promise<WorkspaceDirectoriesResponse> {
    const params = new URLSearchParams()
    if (path) {
      params.set('path', path)
    }
    const suffix = params.toString()
    return api.get(`/workspace/directories${suffix ? `?${suffix}` : ''}`)
  },
}
