import { computed, ref } from 'vue'
import { ApiError } from '@/api/client'
import { filesApi } from '@/api/files'
import type { FileContentResponse } from '@/api/types'
import type { OpenFileTab } from '@/components/Workspace'

const WELCOME_PATH = '__welcome__.txt'
const UNTITLED_PATH = '__untitled__.txt'

const welcomeContent = `欢迎使用 ioline

当前还没有选择工作区。

你可以：
1. 点击左上角打开侧栏
2. 在文件树面板中选择工作区
3. 选择成功后开始浏览和打开文件
`

const untitledContent = ''

function createWelcomeTab(): OpenFileTab {
  return {
    name: 'welcome.txt',
    path: WELCOME_PATH,
    dirty: false,
    temporary: true,
    closable: true,
    draft: welcomeContent,
  }
}

function createUntitledTab(): OpenFileTab {
  return {
    name: 'untitled.txt',
    path: UNTITLED_PATH,
    dirty: false,
    temporary: true,
    closable: true,
    draft: untitledContent,
  }
}

function upsertTab(tabs: OpenFileTab[], tab: OpenFileTab): OpenFileTab[] {
  const exists = tabs.some((item) => item.path === tab.path)
  if (exists) {
    return tabs.map((item) => (item.path === tab.path ? { ...item, ...tab } : item))
  }
  return [...tabs, tab]
}

export function useWorkspaceSession() {
  const code = ref(welcomeContent)
  const currentFile = ref(WELCOME_PATH)
  const currentFileInfo = ref<FileContentResponse | null>(null)
  const isEditorReadonly = ref(true)
  const openFileTabs = ref<OpenFileTab[]>([createWelcomeTab()])
  const saving = ref(false)
  const currentError = ref<string | null>(null)
  const toastMessage = ref<string | null>(null)
  const saveAllConfirmOpen = ref(false)
  const closingTabPath = ref<string | null>(null)
  const closeConfirmOpen = ref(false)

  const currentFileName = computed(() => {
    const activeTab = openFileTabs.value.find((tab) => tab.path === currentFile.value)
    return activeTab?.name ?? 'untitled.txt'
  })

  const currentFileRelativePath = computed(() => {
    if (currentFile.value === WELCOME_PATH) return 'welcome.txt'
    if (currentFile.value === UNTITLED_PATH) return 'untitled.txt'
    return currentFile.value
  })

  function pushToast(message: string) {
    toastMessage.value = message
  }

  function resetToWelcome() {
    currentFile.value = WELCOME_PATH
    currentFileInfo.value = null
    code.value = welcomeContent
    isEditorReadonly.value = true
    openFileTabs.value = [createWelcomeTab()]
  }

  function resetToUntitled() {
    currentFile.value = UNTITLED_PATH
    currentFileInfo.value = null
    code.value = untitledContent
    isEditorReadonly.value = false
    openFileTabs.value = [createUntitledTab()]
  }

  function onWorkspaceCleared() {
    resetToWelcome()
  }

  function onWorkspaceReady() {
    resetToUntitled()
  }

  function currentTab() {
    return openFileTabs.value.find((tab) => tab.path === currentFile.value) ?? null
  }

  function setTabState(path: string, patch: Partial<OpenFileTab>) {
    openFileTabs.value = openFileTabs.value.map((item) => {
      if (item.path !== path) return item
      return {
        ...item,
        ...patch,
      }
    })
  }

  function markCurrentTabDirtyByCode() {
    const tab = currentTab()
    if (!tab) return

    let dirty = false
    if (tab.path === WELCOME_PATH) {
      dirty = code.value !== welcomeContent
    } else if (tab.path === UNTITLED_PATH) {
      dirty = code.value !== untitledContent
    } else if (currentFileInfo.value) {
      dirty = code.value !== currentFileInfo.value.content
    }

    setTabState(tab.path, {
      dirty,
      draft: code.value,
    })
  }

  function closeUntitledIfPristine() {
    openFileTabs.value = openFileTabs.value.filter((tab) => {
      if (tab.path !== UNTITLED_PATH) return true
      return Boolean(tab.dirty)
    })
  }

  function activateFile(params: {
    path: string
    name: string
    content: string
    readonly: boolean
    info: FileContentResponse
  }) {
    closeUntitledIfPristine()
    currentFile.value = params.path
    currentFileInfo.value = params.info
    code.value = params.content
    isEditorReadonly.value = params.readonly
    openFileTabs.value = upsertTab(openFileTabs.value, {
      name: params.name,
      path: params.path,
      dirty: false,
      temporary: false,
      closable: true,
      draft: params.content,
    })
    pushToast(`已打开：${params.name}`)
  }

  function activateWelcome() {
    currentFile.value = WELCOME_PATH
    currentFileInfo.value = null
    code.value = welcomeContent
    isEditorReadonly.value = true
    openFileTabs.value = upsertTab(openFileTabs.value, createWelcomeTab())
  }

  function activateUntitled() {
    currentFile.value = UNTITLED_PATH
    currentFileInfo.value = null
    code.value = untitledContent
    isEditorReadonly.value = false
    openFileTabs.value = upsertTab(openFileTabs.value, createUntitledTab())
  }

  function switchTab(path: string) {
    const tab = openFileTabs.value.find((item) => item.path === path)
    if (!tab) return false

    if (path === WELCOME_PATH) {
      activateWelcome()
      return true
    }

    if (path === UNTITLED_PATH) {
      activateUntitled()
      return true
    }

    return false
  }

  function requestCloseTab(path: string): boolean {
    const tab = openFileTabs.value.find((item) => item.path === path)
    if (!tab) return false

    if (tab.dirty) {
      closingTabPath.value = path
      closeConfirmOpen.value = true
      return false
    }

    closeTabImmediately(path)
    return true
  }

  function closeTabImmediately(path: string) {
    const nextTabs = openFileTabs.value.filter((tab) => tab.path !== path)
    const wasActive = currentFile.value === path
    openFileTabs.value = nextTabs

    if (!wasActive) {
      ensureFallbackTab()
      return
    }

    const nextActiveTab = nextTabs[nextTabs.length - 1] ?? null
    if (!nextActiveTab) {
      ensureFallbackTab()
      return
    }

    if (nextActiveTab.path === WELCOME_PATH) {
      activateWelcome()
      return
    }

    if (nextActiveTab.path === UNTITLED_PATH) {
      activateUntitled()
      return
    }

    currentFile.value = nextActiveTab.path
  }

  function confirmCloseDirtyTab() {
    if (!closingTabPath.value) return
    const path = closingTabPath.value
    closeConfirmOpen.value = false
    closingTabPath.value = null
    closeTabImmediately(path)
  }

  function cancelCloseDirtyTab() {
    closeConfirmOpen.value = false
    closingTabPath.value = null
  }

  function ensureFallbackTab() {
    if (openFileTabs.value.length > 0) return
    if (currentFile.value === WELCOME_PATH) {
      activateWelcome()
      return
    }
    activateUntitled()
  }

  async function saveSingleTab(tab: OpenFileTab): Promise<boolean> {
    if (tab.path === WELCOME_PATH || tab.path === UNTITLED_PATH) {
      return true
    }

    const content = tab.path === currentFile.value ? code.value : tab.draft ?? ''

    try {
      const res = await filesApi.saveContent({
        path: tab.path,
        content,
      })

      setTabState(tab.path, {
        dirty: false,
        draft: content,
      })

      if (currentFile.value === tab.path && currentFileInfo.value) {
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
    const tab = currentTab()
    if (!tab || tab.path === WELCOME_PATH || tab.path === UNTITLED_PATH || isEditorReadonly.value) {
      return false
    }

    saving.value = true
    currentError.value = null
    try {
      return await saveSingleTab(tab)
    } finally {
      saving.value = false
    }
  }

  function requestSaveAll() {
    saveAllConfirmOpen.value = true
  }

  function cancelSaveAll() {
    saveAllConfirmOpen.value = false
  }

  async function confirmSaveAll(): Promise<boolean> {
    saveAllConfirmOpen.value = false
    const dirtyTabs = openFileTabs.value.filter(
      (tab) => tab.dirty && tab.path !== WELCOME_PATH && tab.path !== UNTITLED_PATH,
    )
    if (dirtyTabs.length === 0) {
      return true
    }

    saving.value = true
    currentError.value = null
    try {
      for (const tab of dirtyTabs) {
        const saved = await saveSingleTab(tab)
        if (!saved) {
          return false
        }
      }
      return true
    } finally {
      saving.value = false
    }
  }

  return {
    activateFile,
    cancelCloseDirtyTab,
    cancelSaveAll,
    closeConfirmOpen,
    code,
    confirmCloseDirtyTab,
    confirmSaveAll,
    currentError,
    currentFile,
    currentFileInfo,
    currentFileName,
    currentFileRelativePath,
    handleCodeInput: markCurrentTabDirtyByCode,
    isEditorReadonly,
    onWorkspaceCleared,
    onWorkspaceReady,
    openFileTabs,
    requestCloseTab,
    requestSaveAll,
    saveAllConfirmOpen,
    saveCurrentFile,
    saving,
    switchTab,
    toastMessage,
  }
}
