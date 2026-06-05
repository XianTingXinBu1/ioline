<script setup lang="ts">
import { FolderTree, Settings2 } from '@lucide/vue'
import WorkspaceEmptyState from './WorkspaceEmptyState.vue'
import WorkspaceFileList from './WorkspaceFileList.vue'
import type { SidebarEntry } from './types'

const props = withDefaults(
  defineProps<{
    open: boolean
    activePanel: 'files' | 'settings'
    entries: SidebarEntry[]
    currentFile: string
    workspaceReady?: boolean
    workspaceName?: string
    workspacePickerOpen?: boolean
  }>(),
  {
    open: false,
    activePanel: 'files',
    workspaceReady: false,
    workspaceName: '',
    workspacePickerOpen: false,
  },
)

const emit = defineEmits<{
  close: []
  switchPanel: [panel: 'files' | 'settings']
  selectFile: [entry: SidebarEntry]
  openWorkspacePicker: []
}>()

const longPressDelay = 550
let longPressTimer: ReturnType<typeof setTimeout> | null = null
let longPressTriggered = false

function startWorkspaceIconPress() {
  longPressTriggered = false
  longPressTimer = setTimeout(() => {
    longPressTriggered = true
    emit('openWorkspacePicker')
  }, longPressDelay)
}

function cancelWorkspaceIconPress() {
  if (longPressTimer) {
    clearTimeout(longPressTimer)
    longPressTimer = null
  }
}

function handleWorkspaceIconClick() {
  if (longPressTriggered) {
    longPressTriggered = false
    return
  }

  emit('switchPanel', 'files')
}
</script>

<template>
  <div class="workspace-sidebar">
    <div
      v-if="open"
      class="sidebar-backdrop"
      aria-hidden="true"
      @click="emit('close')"
    ></div>

    <aside
      id="workspace-sidebar"
      class="file-tree"
      :class="{ 'file-tree--open': open }"
    >
      <div class="sidebar-rail">
        <button
          class="sidebar-rail__btn"
          :class="{ 'sidebar-rail__btn--active': activePanel === 'files' }"
          :title="workspaceReady ? '文件树（长按可重新选择工作区）' : '文件树'"
          @pointerdown="startWorkspaceIconPress"
          @pointerup="cancelWorkspaceIconPress"
          @pointerleave="cancelWorkspaceIconPress"
          @pointercancel="cancelWorkspaceIconPress"
          @click="handleWorkspaceIconClick"
        >
          <FolderTree aria-hidden="true" />
        </button>
        <button
          class="sidebar-rail__btn"
          :class="{ 'sidebar-rail__btn--active': activePanel === 'settings' }"
          title="设置"
          @click="emit('switchPanel', 'settings')"
        >
          <Settings2 aria-hidden="true" />
        </button>
      </div>

      <div class="sidebar-panel">
        <div class="file-tree__header-row">
          <div class="file-tree__header">
            {{ activePanel === 'files' ? (workspaceReady ? workspaceName || '文件' : '选择工作区') : '设置' }}
          </div>
          <button class="file-tree__close" title="关闭侧栏" @click="emit('close')">✕</button>
        </div>

        <template v-if="activePanel === 'files'">
          <WorkspaceFileList
            v-if="workspaceReady"
            :entries="entries"
            :current-file="currentFile"
            @select="emit('selectFile', $event)"
          />
          <WorkspaceEmptyState
            v-else
            @open-workspace-picker="emit('openWorkspacePicker')"
          />
        </template>

        <div v-else class="settings-panel">
          <div class="settings-panel__card">
            <div class="settings-panel__title">设置面板占位</div>
            <div class="settings-panel__desc">后续可在这里接入编辑器偏好、主题与账户相关配置。</div>
          </div>
        </div>
      </div>
    </aside>
  </div>
</template>

<style scoped>
.workspace-sidebar {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.workspace-sidebar > * {
  pointer-events: auto;
}

.sidebar-backdrop {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.42);
  z-index: 9;
}

.file-tree {
  position: absolute;
  top: 0;
  left: 0;
  bottom: 0;
  width: min(86vw, 360px);
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-color);
  overflow: hidden;
  display: flex;
  transform: translateX(-100%);
  transition: transform 0.2s ease;
  z-index: 10;
}

.file-tree--open {
  transform: translateX(0);
}

.sidebar-rail {
  width: 56px;
  background: rgba(0, 0, 0, 0.12);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 12px 8px;
  flex-shrink: 0;
}

.sidebar-rail__btn {
  width: 40px;
  height: 40px;
  border: none;
  border-radius: 12px;
  background: transparent;
  color: var(--text-tertiary);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s, color 0.15s;
}

.sidebar-rail__btn :deep(svg) {
  width: 22px;
  height: 22px;
  stroke-width: 1.85;
}

.sidebar-rail__btn:active {
  background: var(--border-color);
}

.sidebar-rail__btn--active {
  background: rgba(230, 57, 124, 0.12);
  color: var(--text-primary);
}

.sidebar-panel {
  min-width: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.file-tree__header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 10px 4px 16px;
}

.file-tree__header {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.8px;
  color: var(--text-tertiary);
}

.file-tree__close {
  background: none;
  border: none;
  color: var(--text-secondary);
  width: 36px;
  height: 36px;
  border-radius: 8px;
  font-size: 16px;
}

.file-tree__close:active {
  background: var(--border-color);
}

.settings-panel {
  flex: 1;
  padding: 12px;
  overflow-y: auto;
}

.settings-panel__card {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 14px;
}

.settings-panel__title {
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 600;
}

.settings-panel__desc {
  margin-top: 8px;
  color: var(--text-secondary);
  font-size: 12px;
  line-height: 1.6;
}
</style>
