<script setup lang="ts">
/**
 * Workspace 主工作区页面
 * 莫奈暗色 · 紫粉青印象派色调
 */
import { ref } from 'vue'
import { CodeEditor } from '@/components/Editor'

const code = ref(`import { createApp } from 'vue'
import App from './App.vue'

// ioline — 移动端代码编辑器
const app = createApp(App)
app.mount('#app')
`)

const currentFile = ref('main.ts')

function openFile(filename: string) {
  currentFile.value = filename
}
</script>

<template>
  <div class="workspace">
    <!-- 工具栏 -->
    <header class="toolbar">
      <button class="toolbar__btn" title="菜单">
        <span class="icon">☰</span>
      </button>
      <span class="toolbar__title">ioline</span>
      <span class="toolbar__file">{{ currentFile }}</span>
      <div class="toolbar__spacer"></div>
      <button class="toolbar__btn toolbar__btn--accent" title="运行">
        <span class="icon">▶</span>
      </button>
    </header>

    <!-- 主内容区 -->
    <main class="main-content">
      <!-- 文件树 -->
      <aside class="file-tree">
        <div class="file-tree__header">文件</div>
        <div class="file-tree__list">
          <button
            class="file-tree__item"
            :class="{ 'file-tree__item--active': currentFile === 'main.ts' }"
            @click="openFile('main.ts')"
          >
            <span class="file-icon">📄</span>
            main.ts
          </button>
          <button
            class="file-tree__item"
            :class="{ 'file-tree__item--active': currentFile === 'App.vue' }"
            @click="openFile('App.vue')"
          >
            <span class="file-icon">📄</span>
            App.vue
          </button>
          <button
            class="file-tree__item"
            :class="{ 'file-tree__item--active': currentFile === 'style.css' }"
            @click="openFile('style.css')"
          >
            <span class="file-icon">📄</span>
            style.css
          </button>
          <button
            class="file-tree__item"
            :class="{ 'file-tree__item--active': currentFile === 'README.md' }"
            @click="openFile('README.md')"
          >
            <span class="file-icon">📄</span>
            README.md
          </button>
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
      <span class="status-bar__item">{{ currentFile }}</span>
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
  -webkit-app-region: drag;
  flex-shrink: 0;
}

.toolbar__btn {
  -webkit-app-region: no-drag;
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
  display: flex;
  flex: 1;
  overflow: hidden;
}

/* ===== File Tree ===== */
.file-tree {
  width: var(--filetree-width);
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-color);
  overflow-y: auto;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
}

.file-tree__header {
  padding: 16px 16px 8px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.8px;
  color: var(--text-tertiary);
}

.file-tree__list {
  display: flex;
  flex-direction: column;
  gap: 1px;
  padding: 0 6px;
}

.file-tree__item {
  background: none;
  border: none;
  color: var(--text-secondary);
  padding: 8px 12px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  text-align: left;
  display: flex;
  align-items: center;
  gap: 8px;
  touch-action: manipulation;
  transition: background 0.15s;
}

.file-tree__item:active {
  background: var(--border-color);
}

.file-tree__item--active {
  background: rgba(230, 57, 124, 0.10);
  color: var(--text-primary);
  border-left: 2px solid var(--accent);
  padding-left: 10px;
}

.file-icon {
  font-size: 14px;
  flex-shrink: 0;
}

/* ===== Editor Pane ===== */
.editor-pane {
  flex: 1;
  overflow: hidden;
  min-width: 0;
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

/* ===== Mobile ===== */
@media (max-width: 768px) {
  .file-tree {
    display: none;
  }

  .toolbar__file {
    display: none;
  }

  .status-bar__item:nth-child(2) {
    display: none;
  }
}
</style>
