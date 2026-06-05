export type SidebarEntry = {
  name: string
  path: string
  kind: 'file' | 'directory'
  depth: number
  readonly?: boolean
  hidden?: boolean
  expanded?: boolean
  loading?: boolean
}

export type WorkspaceCandidate = {
  name: string
  path: string
  source: string
}
