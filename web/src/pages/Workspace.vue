<script setup lang="ts">
/**
 * Workspace 主工作区页面
 * 当前仅面向移动端：编辑器主视图 + 侧栏抽屉
 */
import { onMounted, ref } from 'vue'
import { CodeEditor } from '@/components/Editor'
import { WorkspaceSidebar } from '@/components/Workspace'
import { useWorkspaceExplorer } from '@/composables/useWorkspaceExplorer'

const {
  candidatesLoading,
  code,
  currentError,
  currentFile,
  currentFileName,
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
} = useWorkspaceExplorer()

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
  const opened = await openEntry(entry)
  if (opened) {
    isSidebarOpen.value = false
  }
}

async function handleWorkspaceSelect(candidate: (typeof workspaceCandidates.value)[number]) {
  const selected = await selectWorkspace(candidate)
  if (selected) {
    isSidebarOpen.value = false
  }
}

onMounted(() => {
  void initializeWorkspace()
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
      <span class="toolbar__title">ioline</span>
      <span class="toolbar__file">{{ currentFileName }}</span>
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
        :candidates="workspaceCandidates"
        :candidate-loading="candidatesLoading"
        :selecting-workspace-path="selectingWorkspacePath"
        @close="closeSidebar"
        @switch-panel="switchSidebarPanel"
        @select-file="handleEntrySelect"
        @select-workspace="handleWorkspaceSelect"
        @refresh-candidates="loadWorkspaceCandidates"
      />

      <section class="editor-pane">
        <CodeEditor v-model="code" :readonly="isEditorReadonly" />
      </section>
    </main>

    <footer class="status-bar">
      <span class="status-bar__item status-bar__item--purple">
        {{ workspaceReady ? '工作区模式' : '欢迎页' }}
      </span>
      <span class="status-bar__item">{{ currentFileName }}</span>
      <span class="status-bar__spacer"></span>
      <span class="status-bar__item status-bar__item--cyan">UTF-8</span>
      <span class="status-bar__item" :class="currentError ? 'status-bar__item--error' : 'status-bar__item--pink'">
        {{ loading || fileTreeLoading ? '加载中' : statusText }}
      </span>
    </footer>
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

.toolbar__title {
  font-size: 15px;
  font-weight: 600;
  color: var(--accent);
  letter-spacing: 0.3px;
}

.toolbar__file {
  font-size: 12px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
  padding: 2px 10px;
  background: var(--bg-tertiary);
  border-radius: 4px;
  border: 1px solid var(--border-color);
}

.toolbar__spacer {
  flex: 1;
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
  gap: 16px;
}

.status-bar__item {
  color: var(--text-tertiary);
  white-space: nowrap;
}

.status-bar__item--purple {
  color: var(--accent);
}

.status-bar__item--pink {
  color: var(--accent-pink);
}

.status-bar__item--cyan {
  color: var(--accent-cyan);
}

.status-bar__item--error {
  color: var(--error);
}

.status-bar__spacer {
  flex: 1;
}

.toolbar__file {
  display: none;
}

.status-bar__item:nth-child(2) {
  display: none;
}
</style>
