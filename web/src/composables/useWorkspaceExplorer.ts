import { ref } from 'vue'
import { ApiError } from '@/api/client'
import { filesApi } from '@/api/files'
import { workspaceApi } from '@/api/workspace'
import type { FileContentResponse, FileListItemResponse } from '@/api/types'
import type { SidebarEntry, WorkspaceDirectory } from '@/components/Workspace'

function mapItemToEntry(
  item: FileListItemResponse,
  depth: number,
  expandedDirectories: Set<string>,
): SidebarEntry {
  return {
    name: item.name,
    path: item.path,
    kind: item.type,
    depth,
    readonly: item.readonly,
    hidden: item.hidden,
    expanded: expandedDirectories.has(item.path),
  }
}

function getEntryDepth(path: string): number {
  if (path === '.' || path === '') return 0
  return path.split('/').length - 1
}

export function useWorkspaceExplorer() {
  const workspaceName = ref('')
  const workspaceRootPath = ref('')
  const workspaceReady = ref(false)
  const sidebarEntries = ref<SidebarEntry[]>([])
  const workspaceDirectories = ref<WorkspaceDirectory[]>([])
  const workspaceDirectoryPath = ref('')
  const workspaceDirectoryParentPath = ref('')
  const loading = ref(false)
  const fileTreeLoading = ref(false)
  const directoryLoading = ref(false)
  const selectingWorkspacePath = ref<string | null>(null)
  const currentError = ref<string | null>(null)
  const loadedDirectories = ref(new Set<string>())
  const expandedDirectories = ref(new Set<string>())
  const workspacePickerOpen = ref(false)

  function resetWorkspaceState() {
    workspaceReady.value = false
    workspaceName.value = ''
    workspaceRootPath.value = ''
    sidebarEntries.value = []
    loadedDirectories.value = new Set<string>()
    expandedDirectories.value = new Set<string>()
    workspacePickerOpen.value = false
  }

  function findEntryIndex(path: string): number {
    return sidebarEntries.value.findIndex((entry) => entry.path === path)
  }

  function removeNestedEntries(parentPath: string) {
    const prefix = `${parentPath}/`
    sidebarEntries.value = sidebarEntries.value.filter((entry) => {
      if (entry.path === parentPath) return true
      return !entry.path.startsWith(prefix)
    })
  }

  function updateDirectoryExpandedState(path: string, expanded: boolean) {
    sidebarEntries.value = sidebarEntries.value.map((entry) => {
      if (entry.path !== path) return entry
      return {
        ...entry,
        expanded,
      }
    })
  }

  async function browseWorkspaceDirectories(path?: string) {
    directoryLoading.value = true
    currentError.value = null

    try {
      const res = await workspaceApi.getDirectories(path)
      workspaceDirectoryPath.value = res.currentPath
      workspaceDirectoryParentPath.value = res.parentPath
      workspaceDirectories.value = res.items.map((item) => ({
        name: item.name,
        path: item.path,
      }))
    } catch (error) {
      currentError.value = error instanceof ApiError ? error.message : '加载目录失败'
    } finally {
      directoryLoading.value = false
    }
  }

  async function showWorkspacePicker() {
    workspacePickerOpen.value = true
    await browseWorkspaceDirectories()
  }

  function hideWorkspacePicker() {
    workspacePickerOpen.value = false
  }

  async function openWorkspaceDirectory(path: string) {
    await browseWorkspaceDirectories(path)
  }

  async function openWorkspaceDirectoryParent() {
    if (!workspaceDirectoryParentPath.value) return
    await browseWorkspaceDirectories(workspaceDirectoryParentPath.value)
  }

  async function loadDirectory(path: string, options?: { depth?: number; append?: boolean }) {
    const depth = options?.depth ?? getEntryDepth(path)
    const append = options?.append ?? false

    fileTreeLoading.value = true
    currentError.value = null

    try {
      const res = await filesApi.list(path)
      const entries = res.items.map((item) => mapItemToEntry(item, depth, expandedDirectories.value))

      if (!append || path === '.') {
        sidebarEntries.value = entries
      } else {
        const parentIndex = findEntryIndex(path)
        if (parentIndex === -1) return

        removeNestedEntries(path)
        const refreshedParentIndex = findEntryIndex(path)
        const nextEntries = [...sidebarEntries.value]
        nextEntries.splice(refreshedParentIndex + 1, 0, ...entries)
        sidebarEntries.value = nextEntries
      }

      loadedDirectories.value.add(path)
      if (path !== '.') {
        expandedDirectories.value.add(path)
        updateDirectoryExpandedState(path, true)
      }
    } catch (error) {
      currentError.value = error instanceof ApiError ? error.message : '加载目录失败'
    } finally {
      fileTreeLoading.value = false
    }
  }

  function collapseDirectory(path: string) {
    expandedDirectories.value.delete(path)
    removeNestedEntries(path)
    updateDirectoryExpandedState(path, false)
  }

  async function openEntry(entry: SidebarEntry): Promise<FileContentResponse | null> {
    if (entry.kind === 'directory') {
      if (expandedDirectories.value.has(entry.path)) {
        collapseDirectory(entry.path)
        return null
      }

      await loadDirectory(entry.path, {
        depth: entry.depth + 1,
        append: true,
      })
      return null
    }

    loading.value = true
    currentError.value = null

    try {
      return await filesApi.getContent(entry.path)
    } catch (error) {
      currentError.value = error instanceof ApiError ? error.message : '打开文件失败'
      return null
    } finally {
      loading.value = false
    }
  }

  async function selectWorkspaceByPath(path: string): Promise<{ name?: string; rootPath?: string } | null> {
    selectingWorkspacePath.value = path
    currentError.value = null

    try {
      const res = await workspaceApi.setCurrent({ rootPath: path })
      workspaceReady.value = true
      workspaceName.value = res.name || path.split('/').filter(Boolean).pop() || '工作区'
      workspaceRootPath.value = res.rootPath || path
      workspacePickerOpen.value = false
      await loadDirectory('.')
      return res
    } catch (error) {
      currentError.value = error instanceof ApiError ? error.message : '设置工作区失败'
      return null
    } finally {
      selectingWorkspacePath.value = null
    }
  }

  async function initializeWorkspace() {
    loading.value = true
    currentError.value = null

    try {
      const current = await workspaceApi.getCurrent()

      if (!current.isSet) {
        resetWorkspaceState()
        return null
      }

      workspaceReady.value = true
      workspaceName.value = current.name || '工作区'
      workspaceRootPath.value = current.rootPath || ''
      await loadDirectory('.')
      return current
    } catch (error) {
      currentError.value = error instanceof ApiError ? error.message : '初始化工作区失败'
      resetWorkspaceState()
      return null
    } finally {
      loading.value = false
    }
  }

  return {
    currentError,
    directoryLoading,
    fileTreeLoading,
    hideWorkspacePicker,
    initializeWorkspace,
    loading,
    openEntry,
    openWorkspaceDirectory,
    openWorkspaceDirectoryParent,
    selectWorkspaceByPath,
    selectingWorkspacePath,
    showWorkspacePicker,
    sidebarEntries,
    workspaceDirectories,
    workspaceDirectoryParentPath,
    workspaceDirectoryPath,
    workspaceName,
    workspacePickerOpen,
    workspaceReady,
    workspaceRootPath,
    resetWorkspaceState,
  }
}
