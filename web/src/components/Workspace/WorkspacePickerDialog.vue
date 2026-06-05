<script setup lang="ts">
import type { WorkspaceDirectory } from './types'

withDefaults(
  defineProps<{
    open: boolean
    loading?: boolean
    selectingPath?: string | null
    currentPath?: string
    parentPath?: string
    items: WorkspaceDirectory[]
  }>(),
  {
    open: false,
    loading: false,
    selectingPath: null,
    currentPath: '',
    parentPath: '',
    items: () => [],
  },
)

const emit = defineEmits<{
  close: []
  refresh: []
  selectCurrent: []
  openDirectory: [path: string]
  openParent: []
}>()
</script>

<template>
  <div v-if="open" class="picker-dialog">
    <div class="picker-dialog__backdrop" aria-hidden="true" @click="emit('close')"></div>

    <div class="picker-dialog__panel" role="dialog" aria-modal="true" aria-label="选择工作区">
      <div class="picker-dialog__header">
        <div class="picker-dialog__title">选择工作区</div>
        <button class="picker-dialog__close" title="关闭" @click="emit('close')">✕</button>
      </div>

      <div class="picker-dialog__content">
        <div class="picker-dialog__desc">逐级进入目录，确认后将当前目录设置为工作区。</div>

        <div class="picker-dialog__path-card">
          <div class="picker-dialog__path-label">当前目录</div>
          <div class="picker-dialog__path-value">{{ currentPath || '加载中...' }}</div>
        </div>

        <div class="picker-dialog__actions">
          <button class="picker-dialog__secondary" :disabled="loading" @click="emit('refresh')">
            {{ loading ? '刷新中...' : '刷新目录' }}
          </button>
          <button class="picker-dialog__secondary" :disabled="loading || !parentPath" @click="emit('openParent')">
            返回上一级
          </button>
        </div>

        <button class="picker-dialog__primary" :disabled="loading || !currentPath || selectingPath === currentPath" @click="emit('selectCurrent')">
          {{ selectingPath === currentPath ? '设置中...' : '选择当前目录' }}
        </button>

        <div v-if="items.length > 0" class="picker-dialog__list">
          <button
            v-for="item in items"
            :key="item.path"
            class="picker-dialog__item"
            :disabled="loading"
            @click="emit('openDirectory', item.path)"
          >
            <span class="picker-dialog__name">{{ item.name }}</span>
            <span class="picker-dialog__path">{{ item.path }}</span>
          </button>
        </div>

        <div v-else class="picker-dialog__placeholder">当前目录下暂无可继续进入的子目录，你可以直接选择当前目录作为工作区。</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.picker-dialog {
  position: fixed;
  inset: 0;
  z-index: 30;
}

.picker-dialog__backdrop {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.48);
}

.picker-dialog__panel {
  position: absolute;
  left: 12px;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 16px 40px rgba(0, 0, 0, 0.35);
}

.picker-dialog__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 14px 8px;
}

.picker-dialog__title {
  color: var(--text-primary);
  font-size: 15px;
  font-weight: 600;
}

.picker-dialog__close {
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 10px;
  background: transparent;
  color: var(--text-secondary);
  font-size: 16px;
}

.picker-dialog__close:active {
  background: var(--border-color);
}

.picker-dialog__content {
  padding: 0 14px 14px;
}

.picker-dialog__desc,
.picker-dialog__placeholder,
.picker-dialog__path-label,
.picker-dialog__path-value,
.picker-dialog__path {
  line-height: 1.6;
}

.picker-dialog__desc {
  color: var(--text-secondary);
  font-size: 12px;
}

.picker-dialog__path-card {
  margin-top: 12px;
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.02);
}

.picker-dialog__path-label {
  color: var(--text-tertiary);
  font-size: 11px;
  text-transform: uppercase;
}

.picker-dialog__path-value {
  margin-top: 4px;
  color: var(--text-primary);
  font-size: 12px;
  word-break: break-all;
}

.picker-dialog__actions {
  margin-top: 12px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}

.picker-dialog__primary,
.picker-dialog__secondary {
  border: 1px solid var(--border-color);
  border-radius: 10px;
  padding: 10px 12px;
  font-size: 13px;
  color: var(--text-primary);
}

.picker-dialog__primary {
  width: 100%;
  margin-top: 8px;
  background: rgba(230, 57, 124, 0.12);
}

.picker-dialog__secondary {
  background: transparent;
}

.picker-dialog__primary:disabled,
.picker-dialog__secondary:disabled,
.picker-dialog__item:disabled {
  opacity: 0.6;
}

.picker-dialog__list {
  margin-top: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: min(44vh, 360px);
  overflow-y: auto;
}

.picker-dialog__item {
  width: 100%;
  border: 1px solid var(--border-color);
  background: rgba(255, 255, 255, 0.02);
  color: var(--text-primary);
  border-radius: 10px;
  padding: 10px 12px;
  text-align: left;
}

.picker-dialog__name {
  display: block;
  font-size: 13px;
  font-weight: 600;
}

.picker-dialog__path {
  display: block;
  margin-top: 4px;
  color: var(--text-secondary);
  font-size: 11px;
  word-break: break-all;
}

.picker-dialog__placeholder {
  margin-top: 12px;
  color: var(--text-tertiary);
  font-size: 12px;
}
</style>
