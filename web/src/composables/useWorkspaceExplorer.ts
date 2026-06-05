import { computed, ref } from 'vue'
import { ApiError } from '@/api/client'
import { filesApi } from '@/api/files'
import { workspaceApi } from '@/api/workspace'
import type { FileContentResponse, FileListItemResponse } from '@/api/types'
import type { SidebarEntry, WorkspaceCandidate } from '@/components/Workspace'

const welcomeContent = `欢迎使用 ioline

当前还没有选择工作区。

你可以：
1. 点击左上角打开侧栏
2. 在文件树面板中选择候选工作区
3. 选择成功后开始浏览和打开文件
`

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
  const code = ref(welcomeContent)
  const currentFile = ref('__welcome__.txt')
  const isEditorReadonly = ref(true)
  const workspaceName = ref('')
  const workspaceReady = ref(false)
  const sidebarEntries = ref<SidebarEntry[]>([])
  const workspaceCandidates = ref<WorkspaceCandidate[]>([])
  const currentFileInfo = ref<FileContentResponse | null>(null)
  const loading = ref(false)
  const fileTreeLoading = ref(false)
  const candidatesLoading = ref(false)
  const selectingWorkspacePath = ref<string | null>(null)
  const currentError = ref<string | null>(null)
  const loadedDirectories = ref(new Set<string>())
  const expandedDirectories = ref(new Set<string>())

  const currentFileName = computed(() => {
    if (!workspaceReady.value) return 'welcome.txt'

    const activeEntry = sidebarEntries.value.find((entry) => entry.path === currentFile.value)
    return activeEntry?.name ?? currentFile.value
  })

  const statusText = computed(() => {
    if (loading.value || fileTreeLoading.value) return '加载中'
    if (currentError.value) return currentError.value
    if (!workspaceReady.value) return '未选择工作区'
    if (currentFileInfo.value?.readonly) return '只读'
    return '已连接'
  })

  function resetWorkspaceState() {
    workspaceReady.value = false
    workspaceName.value = ''
    sidebarEntries.value = []
    currentFile.value = '__welcome__.txt'
    currentFileInfo.value = null
    code.value = welcomeContent
    isEditorReadonly.value = true
    loadedDirectories.value = new Set<string>()
    expandedDirectories.value = new Set<string>()
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

  async function loadWorkspaceCandidates() {
    candidatesLoading.value = true
    currentError.value = null

    try {
      const res = await workspaceApi.getCandidates()
      workspaceCandidates.value = res.items.map((item) => ({
        name: item.name,
        path: item.path,
        source: item.source,
      }))
    } catch (error) {
      currentError.value = error instanceof ApiError ? error.message : '加载候选工作区失败'
    } finally {
      candidatesLoading.value = false
    }
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

  async function openEntry(entry: SidebarEntry): Promise<boolean> {
    if (entry.kind === 'directory') {
      if (expandedDirectories.value.has(entry.path)) {
        collapseDirectory(entry.path)
        return false
      }

      await loadDirectory(entry.path, {
        depth: entry.depth + 1,
        append: true,
      })
      return false
    }

    loading.value = true
    currentError.value = null

    try {
      const res = await filesApi.getContent(entry.path)
      currentFile.value = entry.path
      currentFileInfo.value = res
      code.value = res.content
      isEditorReadonly.value = res.readonly
      return true
    } catch (error) {
      currentError.value = error instanceof ApiError ? error.message : '打开文件失败'
      return false
    } finally {
      loading.value = false
    }
  }

  async function selectWorkspace(candidate: WorkspaceCandidate): Promise<boolean> {
    selectingWorkspacePath.value = candidate.path
    currentError.value = null

    try {
      const res = await workspaceApi.setCurrent({ rootPath: candidate.path })
      workspaceReady.value = true
      workspaceName.value = res.name || candidate.name
      currentFile.value = '__workspace__.txt'
      currentFileInfo.value = null
      code.value = `工作区已切换\n\n${candidate.path}\n\n正在加载文件树...\n`
      isEditorReadonly.value = true
      await loadDirectory('.')
      return true
    } catch (error) {
      currentError.value = error instanceof ApiError ? error.message : '设置工作区失败'
      return false
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
        await loadWorkspaceCandidates()
        return
      }

      workspaceReady.value = true
      workspaceName.value = current.name || '工作区'
      isEditorReadonly.value = true
      code.value = `工作区：${current.rootPath || current.name || ''}\n\n正在加载文件树...\n`
      await loadDirectory('.')
    } catch (error) {
      currentError.value = error instanceof ApiError ? error.message : '初始化工作区失败'
      resetWorkspaceState()
    } finally {
      loading.value = false
    }
  }

  return {
    code,
    currentFile,
    currentFileName,
    currentFileInfo,
    currentError,
    fileTreeLoading,
    initializeWorkspace,
    isEditorReadonly,
    loadWorkspaceCandidates,
    loading,
    openEntry,
    selectWorkspace,
    selectingWorkspacePath,
    sidebarEntries,
    statusText,
    workspaceCandidates,
    workspaceName,
    workspaceReady,
    candidatesLoading,
  }
}
