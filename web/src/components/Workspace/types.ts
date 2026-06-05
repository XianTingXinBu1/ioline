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

export type WorkspaceDirectory = {
  name: string
  path: string
}

export type OpenFileTab = {
  name: string
  path: string
  dirty?: boolean
}
