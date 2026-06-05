<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { CodeEditor } from '@/components/Editor'
import {
  EditorQuickKeys,
  WorkspaceConfirmDialog,
  WorkspacePickerDialog,
  WorkspaceSidebar,
  WorkspaceTabs,
  WorkspaceToast,
} from '@/components/Workspace'
import { useEditorQuickKeys } from '@/composables/useEditorQuickKeys'
import { useEditorStatus } from '@/composables/useEditorStatus'
import { useWorkspaceExplorer } from '@/composables/useWorkspaceExplorer'
import { useWorkspaceSession } from '@/composables/useWorkspaceSession'
import { useWorkspaceToast } from '@/composables/useWorkspaceToast'

const explorer = useWorkspaceExplorer()
const session = useWorkspaceSession()

const {
  currentError,
  directoryLoading,
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
} = explorer

const {
  activateFile,
  cancelCloseDirtyTab,
  cancelSaveAll,
  closeConfirmOpen,
  code,
  confirmCloseDirtyTab,
  confirmSaveAll,
  currentFile,
  currentFileInfo,
  currentFileName,
  currentFileRelativePath,
  handleCodeInput,
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
  toastMessage: sessionToastMessage,
} = session

const { charCount, encoding, fileType, lineCount, lineEnding } = useEditorStatus({
  code,
  currentFileName,
  currentFileInfo,
})
const combinedToastMessage = computed(() => sessionToastMessage.value || currentError.value)
const { toastMessage, toastVisible } = useWorkspaceToast(combinedToastMessage)

type EditorExposed = {
  focus: () => void
  insertText: (text: string) => void
  triggerKey: (name: 'Tab' | 'Escape' | 'Enter', modifiers?: { ctrl?: boolean; alt?: boolean }) => void
}

const { altActive, clearModifiers, ctrlActive, toggleAlt, toggleCtrl } = useEditorQuickKeys()
const editorRef = ref<EditorExposed | null>(null)
const editorFocused = ref(false)
const isSidebarOpen = ref(false)
const activeSidebarPanel = ref<'files' | 'settings'>('files')

function toggleSidebar() {
  isSidebarOpen.value = !isSidebarOpen.value
  if (isSidebarOpen.value) {
    activeSidebarPanel.value = 'files'
  }
}

function closeSidebar() {
  isSidebarOpen.value = false
}

function switchSidebarPanel(panel: 'files' | 'settings') {
  activeSidebarPanel.value = panel
}

async function handleEntrySelect(entry: (typeof sidebarEntries.value)[number]) {
  const result = await openEntry(entry)
  if (!result) return
  activateFile({
    path: entry.path,
    name: entry.name,
    content: result.content,
    readonly: result.readonly,
    info: result,
  })
}

async function handleCurrentDirectorySelect() {
  const selected = await selectWorkspaceByPath(workspaceDirectoryPath.value)
  if (selected) {
    onWorkspaceReady()
  }
}

async function handleTabSelect(path: string) {
  const switched = switchTab(path)
  if (switched) return

  const tab = openFileTabs.value.find((item: { path: string }) => item.path === path)
  if (!tab) return
  const result = await openEntry({
    name: tab.name,
    path: tab.path,
    kind: 'file',
    depth: 0,
  })
  if (!result) return
  activateFile({
    path: tab.path,
    name: tab.name,
    content: result.content,
    readonly: result.readonly,
    info: result,
  })
}

function handleTabClose(path: string) {
  requestCloseTab(path)
}

function withModifiers() {
  return {
    ctrl: ctrlActive.value,
    alt: altActive.value,
  }
}

function handleQuickTab() {
  const modifiers = withModifiers()
  clearModifiers()
  editorRef.value?.triggerKey('Tab', modifiers)
  editorRef.value?.focus()
}

function handleQuickEsc() {
  const modifiers = withModifiers()
  clearModifiers()
  editorRef.value?.triggerKey('Escape', modifiers)
  editorRef.value?.focus()
}

function handleQuickEnter() {
  const modifiers = withModifiers()
  clearModifiers()
  editorRef.value?.triggerKey('Enter', modifiers)
  editorRef.value?.focus()
}

function handleEditorFocusChange(focused: boolean) {
  editorFocused.value = focused
}

function handleEditorContentInput() {
  handleCodeInput()
  if (ctrlActive.value || altActive.value) {
    clearModifiers()
  }
}

async function handleSave() {
  await saveCurrentFile()
}

function handleSaveAll() {
  requestSaveAll()
}

async function handleSaveAllConfirm() {
  await confirmSaveAll()
}

onMounted(async () => {
  const current = await initializeWorkspace()
  if (current?.isSet) {
    onWorkspaceReady()
    return
  }
  onWorkspaceCleared()
})

onUnmounted(() => {
  // 保留生命周期出口，便于后续若需补回监听时继续扩展
})
</script>

<template>
  <div class="workspace">
    <header class="toolbar">
      <button
        class="toolbar__btn"
        :aria-expanded="isSidebarOpen"
        aria-controls="workspace-sidebar"
        title="菜单"
        @click="toggleSidebar"
      >
        <span class="icon">☰</span>
      </button>
      <span class="toolbar__path">{{ currentFileRelativePath }}</span>
      <div class="toolbar__spacer"></div>
      <button class="toolbar__btn toolbar__btn--accent" title="运行">
        <span class="icon">▶</span>
      </button>
    </header>

    <main class="main-content">
      <WorkspaceSidebar
        :open="isSidebarOpen"
        :active-panel="activeSidebarPanel"
        :entries="sidebarEntries"
        :current-file="currentFile"
        :workspace-ready="workspaceReady"
        :workspace-name="workspaceName"
        :workspace-picker-open="workspacePickerOpen"
        @close="closeSidebar"
        @switch-panel="switchSidebarPanel"
        @select-file="handleEntrySelect"
        @open-workspace-picker="showWorkspacePicker"
      />

      <WorkspacePickerDialog
        :open="workspacePickerOpen"
        :loading="directoryLoading"
        :selecting-path="selectingWorkspacePath"
        :current-path="workspaceDirectoryPath"
        :parent-path="workspaceDirectoryParentPath"
        :items="workspaceDirectories"
        @close="hideWorkspacePicker"
        @refresh="showWorkspacePicker"
        @open-directory="openWorkspaceDirectory"
        @open-parent="openWorkspaceDirectoryParent"
        @select-current="handleCurrentDirectorySelect"
      />

      <section class="editor-pane">
        <WorkspaceTabs
          :tabs="openFileTabs"
          :active-path="currentFile"
          :saving="saving"
          @select="handleTabSelect"
          @close="handleTabClose"
          @save="handleSave"
          @save-all="handleSaveAll"
        />
        <CodeEditor
          ref="editorRef"
          v-model="code"
          :readonly="isEditorReadonly"
          @focus-change="handleEditorFocusChange"
          @content-input="handleEditorContentInput"
        />
      </section>
    </main>

    <WorkspaceConfirmDialog
      :open="saveAllConfirmOpen"
      title="保存全部"
      description="确认保存所有已打开且已修改的文件？"
      confirm-text="保存全部"
      @confirm="handleSaveAllConfirm"
      @cancel="cancelSaveAll"
    />

    <WorkspaceConfirmDialog
      :open="closeConfirmOpen"
      title="关闭标签"
      description="当前文件有未保存修改，确认直接关闭？"
      confirm-text="直接关闭"
      @confirm="confirmCloseDirtyTab"
      @cancel="cancelCloseDirtyTab"
    />

    <WorkspaceToast :visible="toastVisible" :message="toastMessage" />

    <footer class="status-bar">
      <span class="status-bar__item status-bar__item--strong">{{ fileType }}</span>
      <span class="status-bar__item">{{ lineCount }} 行</span>
      <span class="status-bar__item">{{ charCount }} 字</span>
      <span class="status-bar__spacer"></span>
      <span class="status-bar__item">{{ encoding }}</span>
      <span class="status-bar__item">{{ lineEnding }}</span>
    </footer>

    <EditorQuickKeys
      v-if="editorFocused"
      :ctrl-active="ctrlActive"
      :alt-active="altActive"
      :disabled="isEditorReadonly"
      @press-tab="handleQuickTab"
      @press-esc="handleQuickEsc"
      @press-enter="handleQuickEnter"
      @toggle-ctrl="toggleCtrl"
      @toggle-alt="toggleAlt"
    />
  </div>
</template>

<style scoped>
.workspace {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: var(--bg-primary);
  color: var(--text-primary);
}

.toolbar {
  display: flex;
  align-items: center;
  height: var(--toolbar-height);
  padding: 0 8px;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
  gap: 8px;
  flex-shrink: 0;
}

.toolbar__btn {
  background: none;
  border: none;
  color: var(--text-secondary);
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  font-size: 16px;
  touch-action: manipulation;
  transition: background 0.15s;
}

.toolbar__btn:active {
  background: var(--border-color);
}

.toolbar__btn--accent {
  color: var(--accent-pink);
}

.toolbar__btn--accent:active {
  background: rgba(230, 57, 124, 0.14);
}

.toolbar__path {
  min-width: 0;
  flex: 1;
  font-size: 12px;
  color: var(--text-secondary);
  font-family: var(--font-mono);
  padding: 2px 10px;
  background: var(--bg-tertiary);
  border-radius: 6px;
  border: 1px solid var(--border-color);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.toolbar__spacer {
  display: none;
}

.main-content {
  position: relative;
  display: flex;
  flex: 1;
  overflow: hidden;
}

.editor-pane {
  flex: 1;
  overflow: hidden;
  min-width: 0;
  width: 100%;
}

.status-bar {
  display: flex;
  align-items: center;
  height: var(--statusbar-height);
  padding: 0 14px;
  font-size: 12px;
  background: var(--bg-secondary);
  border-top: 1px solid var(--border-color);
  flex-shrink: 0;
  gap: 14px;
}

.status-bar__item {
  color: var(--text-tertiary);
  white-space: nowrap;
  font-variant-numeric: tabular-nums;
}

.status-bar__item--strong {
  color: var(--text-primary);
  text-transform: uppercase;
}

.status-bar__spacer {
  flex: 1;
}
</style>
