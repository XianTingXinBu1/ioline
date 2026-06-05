import { api } from './client'
import type {
  WorkspaceCandidatesResponse,
  WorkspaceCurrentResponse,
  SetWorkspaceRequest,
} from './types'

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

  getCandidates(): Promise<WorkspaceCandidatesResponse> {
    return api.get('/workspaces/candidates')
  },
}
