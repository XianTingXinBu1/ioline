import { computed, ref } from 'vue'
import { ApiError } from '@/api/client'
import { filesApi } from '@/api/files'
import { workspaceApi } from '@/api/workspace'
import type { FileContentResponse, FileListItemResponse } from '@/api/types'
import type { OpenFileTab, SidebarEntry, WorkspaceDirectory } from '@/components/Workspace'

const welcomeContent = `欢迎使用 ioline

当前还没有选择工作区。

你可以：
1. 点击左上角打开侧栏
2. 在文件树面板中选择工作区
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

function upsertOpenFileTab(tabs: OpenFileTab[], tab: OpenFileTab): OpenFileTab[] {
  const exists = tabs.some((item) => item.path === tab.path)
  if (exists) {
    return tabs.map((item) => (item.path === tab.path ? { ...item, ...tab } : item))
  }
  return [...tabs, tab]
}

export function useWorkspaceExplorer() {
  const code = ref(welcomeContent)
  const currentFile = ref('__welcome__.txt')
  const isEditorReadonly = ref(true)
  const workspaceName = ref('')
  const workspaceRootPath = ref('')
  const workspaceReady = ref(false)
  const sidebarEntries = ref<SidebarEntry[]>([])
  const openFileTabs = ref<OpenFileTab[]>([])
  const workspaceDirectories = ref<WorkspaceDirectory[]>([])
  const workspaceDirectoryPath = ref('')
  const workspaceDirectoryParentPath = ref('')
  const currentFileInfo = ref<FileContentResponse | null>(null)
  const loading = ref(false)
  const saving = ref(false)
  const fileTreeLoading = ref(false)
  const directoryLoading = ref(false)
  const selectingWorkspacePath = ref<string | null>(null)
  const currentError = ref<string | null>(null)
  const loadedDirectories = ref(new Set<string>())
  const expandedDirectories = ref(new Set<string>())
  const workspacePickerOpen = ref(false)

  const currentFileName = computed(() => {
    if (!workspaceReady.value) return 'welcome.txt'

    const activeTab = openFileTabs.value.find((tab) => tab.path === currentFile.value)
    if (activeTab) return activeTab.name

    const activeEntry = sidebarEntries.value.find((entry) => entry.path === currentFile.value)
    return activeEntry?.name ?? currentFile.value
  })

  const currentFileRelativePath = computed(() => {
    if (!workspaceReady.value || currentFile.value === '__welcome__.txt') {
      return 'welcome.txt'
    }
    return currentFile.value
  })

  function markCurrentTabDirty(dirty: boolean) {
    if (currentFile.value === '__welcome__.txt') return
    openFileTabs.value = openFileTabs.value.map((tab) => {
      if (tab.path !== currentFile.value) return tab
      return {
        ...tab,
        dirty,
      }
    })
  }

  function syncDirtyStateFromCode() {
    if (currentFile.value === '__welcome__.txt' || !currentFileInfo.value) return
    markCurrentTabDirty(code.value !== currentFileInfo.value.content)
  }

  function resetEditorToWelcome() {
    currentFile.value = '__welcome__.txt'
    currentFileInfo.value = null
    code.value = welcomeContent
    isEditorReadonly.value = true
  }

  function resetWorkspaceState() {
    workspaceReady.value = false
    workspaceName.value = ''
    workspaceRootPath.value = ''
    sidebarEntries.value = []
    openFileTabs.value = []
    resetEditorToWelcome()
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

  async function openFile(path: string, name: string): Promise<boolean> {
    loading.value = true
    currentError.value = null

    try {
      const res = await filesApi.getContent(path)
      currentFile.value = path
      currentFileInfo.value = res
      code.value = res.content
      isEditorReadonly.value = res.readonly
      openFileTabs.value = upsertOpenFileTab(openFileTabs.value, {
        name,
        path,
        dirty: false,
      })
      return true
    } catch (error) {
      currentError.value = error instanceof ApiError ? error.message : '打开文件失败'
      return false
    } finally {
      loading.value = false
    }
  }

  async function saveFile(path: string, content: string): Promise<boolean> {
    try {
      const res = await filesApi.saveContent({ path, content })

      openFileTabs.value = openFileTabs.value.map((tab) => {
        if (tab.path !== path) return tab
        return {
          ...tab,
          dirty: false,
        }
      })

      if (currentFile.value === path && currentFileInfo.value) {
        currentFileInfo.value = {
          ...currentFileInfo.value,
          content,
          size: res.size,
          modifiedAt: res.modifiedAt,
        }
      }

      return true
    } catch (error) {
      currentError.value = error instanceof ApiError ? error.message : '保存文件失败'
      return false
    }
  }

  async function saveCurrentFile(): Promise<boolean> {
    if (!workspaceReady.value || isEditorReadonly.value || currentFile.value === '__welcome__.txt') {
      return false
    }

    saving.value = true
    currentError.value = null

    try {
      return await saveFile(currentFile.value, code.value)
    } finally {
      saving.value = false
    }
  }

  async function saveAllOpenFiles(): Promise<boolean> {
    if (!workspaceReady.value || saving.value) {
      return false
    }

    const dirtyTabs = openFileTabs.value.filter((tab) => tab.dirty)
    if (dirtyTabs.length === 0) {
      return true
    }

    saving.value = true
    currentError.value = null

    try {
      for (const tab of dirtyTabs) {
        const content = tab.path === currentFile.value ? code.value : currentFileInfo.value?.path === tab.path ? currentFileInfo.value.content : null
        if (content === null) {
          continue
        }
        const saved = await saveFile(tab.path, content)
        if (!saved) {
          return false
        }
      }
      return true
    } finally {
      saving.value = false
    }
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

    return openFile(entry.path, entry.name)
  }

  async function openTab(path: string): Promise<boolean> {
    const tab = openFileTabs.value.find((item) => item.path === path)
    if (!tab) return false
    return openFile(tab.path, tab.name)
  }

  function closeTab(path: string): boolean {
    const index = openFileTabs.value.findIndex((tab) => tab.path === path)
    if (index === -1) return false

    const nextTabs = [...openFileTabs.value]
    nextTabs.splice(index, 1)

    const wasActive = currentFile.value === path
    const nextActiveTab = nextTabs[index - 1] ?? nextTabs[index] ?? null

    openFileTabs.value = nextTabs

    if (!wasActive) {
      return false
    }

    if (!nextActiveTab) {
      resetEditorToWelcome()
      return true
    }

    void openFile(nextActiveTab.path, nextActiveTab.name)
    return true
  }

  async function selectWorkspaceByPath(path: string): Promise<boolean> {
    selectingWorkspacePath.value = path
    currentError.value = null

    try {
      const res = await workspaceApi.setCurrent({ rootPath: path })
      workspaceReady.value = true
      workspaceName.value = res.name || path.split('/').filter(Boolean).pop() || '工作区'
      workspaceRootPath.value = res.rootPath || path
      openFileTabs.value = []
      resetEditorToWelcome()
      workspacePickerOpen.value = false
      await loadDirectory('.')
      return true
    } catch (error) {
      currentError.value = error instanceof ApiError ? error.message : '设置工作区失败'
      return false
    } finally {
      selectingWorkspacePath.value = null
    }
  }

  async function selectCurrentWorkspaceDirectory(): Promise<boolean> {
    if (!workspaceDirectoryPath.value) return false
    return selectWorkspaceByPath(workspaceDirectoryPath.value)
  }

  async function initializeWorkspace() {
    loading.value = true
    currentError.value = null

    try {
      const current = await workspaceApi.getCurrent()

      if (!current.isSet) {
        resetWorkspaceState()
        return
      }

      workspaceReady.value = true
      workspaceName.value = current.name || '工作区'
      workspaceRootPath.value = current.rootPath || ''
      resetEditorToWelcome()
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
    closeTab,
    currentError,
    currentFile,
    currentFileInfo,
    currentFileName,
    currentFileRelativePath,
    directoryLoading,
    fileTreeLoading,
    hideWorkspacePicker,
    initializeWorkspace,
    isEditorReadonly,
    loading,
    markCurrentTabDirty,
    openEntry,
    openFileTabs,
    openTab,
    openWorkspaceDirectory,
    openWorkspaceDirectoryParent,
    saveAllOpenFiles,
    saveCurrentFile,
    saving,
    selectCurrentWorkspaceDirectory,
    selectingWorkspacePath,
    showWorkspacePicker,
    sidebarEntries,
    syncDirtyStateFromCode,
    workspaceDirectories,
    workspaceDirectoryParentPath,
    workspaceDirectoryPath,
    workspaceName,
    workspacePickerOpen,
    workspaceReady,
    workspaceRootPath,
  }
}
