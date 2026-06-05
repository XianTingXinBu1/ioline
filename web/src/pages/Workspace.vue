<script setup lang="ts">
/**
 * Workspace 主工作区页面
 * 当前仅面向移动端：编辑器主视图 + 侧栏抽屉
 */
import { computed, ref } from 'vue'
import { CodeEditor } from '@/components/Editor'

type SidebarEntry = {
  name: string
  path: string
  kind: 'file' | 'directory'
}

const code = ref(`import { createApp } from 'vue'
import App from './App.vue'

// ioline — 移动端代码编辑器
const app = createApp(App)
app.mount('#app')
`)

const sidebarEntries: SidebarEntry[] = [
  { name: 'src', path: 'src', kind: 'directory' },
  { name: 'main.ts', path: 'src/main.ts', kind: 'file' },
  { name: 'App.vue', path: 'src/App.vue', kind: 'file' },
  { name: 'styles', path: 'src/styles', kind: 'directory' },
  { name: 'global.css', path: 'src/styles/global.css', kind: 'file' },
  { name: 'README.md', path: 'README.md', kind: 'file' },
]

const currentFile = ref('src/main.ts')
const isSidebarOpen = ref(false)
const activeSidebarPanel = ref<'files' | 'settings'>('files')
const currentFileName = computed(() => {
  const activeEntry = sidebarEntries.find((entry) => entry.path === currentFile.value)
  return activeEntry?.name ?? currentFile.value
})

function openFile(entry: SidebarEntry) {
  if (entry.kind !== 'file') return

  currentFile.value = entry.path
  isSidebarOpen.value = false
}

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
</script>

<template>
  <div class="workspace">
    <!-- 工具栏 -->
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

    <!-- 主内容区 -->
    <main class="main-content">
      <div
        v-if="isSidebarOpen"
        class="sidebar-backdrop"
        aria-hidden="true"
        @click="closeSidebar"
      ></div>

      <!-- 文件树抽屉 -->
      <aside
        id="workspace-sidebar"
        class="file-tree"
        :class="{ 'file-tree--open': isSidebarOpen }"
      >
        <div class="sidebar-rail">
          <button
            class="sidebar-rail__btn"
            :class="{ 'sidebar-rail__btn--active': activeSidebarPanel === 'files' }"
            title="文件树"
            @click="switchSidebarPanel('files')"
          >
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path
                d="M4 6.75A1.75 1.75 0 0 1 5.75 5h4.19c.5 0 .97.21 1.31.57l1.18 1.3c.1.11.25.18.4.18h5.42A1.75 1.75 0 0 1 20 8.8v9.45A1.75 1.75 0 0 1 18.25 20H5.75A1.75 1.75 0 0 1 4 18.25V6.75Zm2.25-.25a.25.25 0 0 0-.25.25v1.3h12.25a.25.25 0 0 0 .25-.25v-.55a.25.25 0 0 0-.25-.25h-5.42a1.75 1.75 0 0 1-1.3-.57l-1.18-1.3a.25.25 0 0 0-.19-.08H6.25Z"
              />
            </svg>
          </button>
          <button
            class="sidebar-rail__btn"
            :class="{ 'sidebar-rail__btn--active': activeSidebarPanel === 'settings' }"
            title="设置"
            @click="switchSidebarPanel('settings')"
          >
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path
                d="M10.77 3.68a1 1 0 0 1 .96-.68h.54a1 1 0 0 1 .96.68l.33 1.02a7.97 7.97 0 0 1 1.58.66l.98-.45a1 1 0 0 1 1.15.2l.38.38a1 1 0 0 1 .2 1.15l-.45.98c.25.5.47 1.02.65 1.57l1.03.34a1 1 0 0 1 .68.95v.54a1 1 0 0 1-.68.96l-1.03.33a7.9 7.9 0 0 1-.65 1.58l.45.98a1 1 0 0 1-.2 1.15l-.38.38a1 1 0 0 1-1.15.2l-.98-.45a7.97 7.97 0 0 1-1.58.66l-.33 1.02a1 1 0 0 1-.96.68h-.54a1 1 0 0 1-.96-.68l-.33-1.02a7.97 7.97 0 0 1-1.58-.66l-.98.45a1 1 0 0 1-1.15-.2l-.38-.38a1 1 0 0 1-.2-1.15l.45-.98a7.9 7.9 0 0 1-.65-1.58l-1.03-.33a1 1 0 0 1-.68-.96v-.54a1 1 0 0 1 .68-.95l1.03-.34a7.9 7.9 0 0 1 .65-1.57l-.45-.98a1 1 0 0 1 .2-1.15l.38-.38a1 1 0 0 1 1.15-.2l.98.45a7.97 7.97 0 0 1 1.58-.66l.33-1.02ZM12 15.25A3.25 3.25 0 1 0 12 8.75a3.25 3.25 0 0 0 0 6.5Z"
              />
            </svg>
          </button>
        </div>

        <div class="sidebar-panel">
          <div class="file-tree__header-row">
            <div class="file-tree__header">{{ activeSidebarPanel === 'files' ? '文件' : '设置' }}</div>
            <button class="file-tree__close" title="关闭侧栏" @click="closeSidebar">✕</button>
          </div>

          <div v-if="activeSidebarPanel === 'files'" class="file-tree__list">
            <button
              v-for="entry in sidebarEntries"
              :key="entry.path"
              class="file-tree__item"
              :class="{
                'file-tree__item--active': entry.kind === 'file' && currentFile === entry.path,
                'file-tree__item--directory': entry.kind === 'directory',
                'file-tree__item--file': entry.kind === 'file',
              }"
              @click="openFile(entry)"
            >
              <span class="file-tree__name">{{ entry.name }}</span>
            </button>
          </div>

          <div v-else class="settings-panel">
            <div class="settings-panel__card">
              <div class="settings-panel__title">设置面板占位</div>
              <div class="settings-panel__desc">后续可在这里接入编辑器偏好、主题与账户相关配置。</div>
            </div>
          </div>
        </div>
      </aside>

      <!-- 编辑器 -->
      <section class="editor-pane">
        <CodeEditor v-model="code" />
      </section>
    </main>

    <!-- 状态栏 -->
    <footer class="status-bar">
      <span class="status-bar__item status-bar__item--purple">纯文本</span>
      <span class="status-bar__item">{{ currentFileName }}</span>
      <span class="status-bar__spacer"></span>
      <span class="status-bar__item status-bar__item--cyan">UTF-8</span>
      <span class="status-bar__item status-bar__item--pink">Ln 4, Col 22</span>
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

/* ===== Toolbar ===== */
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

/* ===== Main Content ===== */
.main-content {
  position: relative;
  display: flex;
  flex: 1;
  overflow: hidden;
}

.sidebar-backdrop {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.42);
  z-index: 9;
}

/* ===== File Tree ===== */
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

.sidebar-rail__btn svg {
  width: 22px;
  height: 22px;
  fill: currentColor;
}

.sidebar-rail__btn:active {
  background: var(--border-color);
}

.sidebar-rail__btn--active {
  background: rgba(230, 57, 124, 0.14);
  color: var(--text-primary);
  box-shadow: inset 2px 0 0 var(--accent);
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

.file-tree__list {
  display: flex;
  flex-direction: column;
  gap: 1px;
  padding: 0 6px 8px;
  overflow-y: auto;
}

.file-tree__item {
  background: none;
  border: none;
  color: var(--text-secondary);
  padding: 10px 12px;
  border-radius: 10px;
  cursor: pointer;
  font-size: 13px;
  text-align: left;
  display: flex;
  align-items: center;
  width: 100%;
  touch-action: manipulation;
  transition: background 0.15s, color 0.15s;
}

.file-tree__item:active {
  background: var(--border-color);
}

.file-tree__item--file {
  color: var(--text-secondary);
}

.file-tree__item--directory {
  color: var(--text-primary);
  font-weight: 600;
}

.file-tree__item--active {
  background: rgba(230, 57, 124, 0.14);
  color: var(--text-primary);
  box-shadow: inset 2px 0 0 var(--accent);
}

.file-tree__name {
  display: block;
  width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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

/* ===== Editor Pane ===== */
.editor-pane {
  flex: 1;
  overflow: hidden;
  min-width: 0;
  width: 100%;
}

/* ===== Status Bar ===== */
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
