<script setup lang="ts">
import { Save } from '@lucide/vue'
import type { OpenFileTab } from './types'

const props = withDefaults(
  defineProps<{
    tabs: OpenFileTab[]
    activePath?: string
    saving?: boolean
  }>(),
  {
    tabs: () => [],
    activePath: '',
    saving: false,
  },
)

const emit = defineEmits<{
  select: [path: string]
  close: [path: string]
  save: []
  saveAll: []
}>()

const longPressDelay = 550
let longPressTimer: ReturnType<typeof setTimeout> | null = null
let longPressTriggered = false

function isSaveDisabled() {
  return props.saving || props.tabs.length === 0
}

function startSavePress() {
  if (isSaveDisabled()) return
  longPressTriggered = false
  longPressTimer = setTimeout(() => {
    longPressTriggered = true
    emit('saveAll')
  }, longPressDelay)
}

function cancelSavePress() {
  if (longPressTimer) {
    clearTimeout(longPressTimer)
    longPressTimer = null
  }
}

function handleSaveClick() {
  if (isSaveDisabled()) return
  if (longPressTriggered) {
    longPressTriggered = false
    return
  }
  emit('save')
}
</script>

<template>
  <div class="workspace-tabs-wrap">
    <div class="workspace-tabs__scroll">
      <div v-if="tabs.length > 0" class="workspace-tabs">
        <button
          v-for="tab in tabs"
          :key="tab.path"
          class="workspace-tabs__item"
          :class="{ 'workspace-tabs__item--active': tab.path === activePath }"
          @click="emit('select', tab.path)"
        >
          <span class="workspace-tabs__indicator" :class="{ 'workspace-tabs__indicator--dirty': tab.dirty }"></span>
          <span class="workspace-tabs__name">{{ tab.name }}</span>
          <span
            class="workspace-tabs__close"
            role="button"
            aria-label="关闭标签"
            @click.stop="emit('close', tab.path)"
          >
            ✕
          </span>
        </button>
      </div>
    </div>

    <div class="workspace-tabs__save-wrap">
      <button
        class="workspace-tabs__save"
        :class="{ 'workspace-tabs__save--busy': saving }"
        :disabled="saving || tabs.length === 0"
        title="保存当前文件（长按保存所有已打开文件）"
        @pointerdown="startSavePress"
        @pointerup="cancelSavePress"
        @pointerleave="cancelSavePress"
        @pointercancel="cancelSavePress"
        @click="handleSaveClick"
      >
        <Save aria-hidden="true" />
      </button>
    </div>
  </div>
</template>

<style scoped>
.workspace-tabs-wrap {
  display: flex;
  align-items: stretch;
  gap: 0;
  height: 38px;
  padding: 0;
  background: var(--bg-primary);
  border-bottom: 1px solid var(--border-color);
  overflow: hidden;
}

.workspace-tabs__scroll {
  min-width: 0;
  flex: 1;
  overflow-x: auto;
  overflow-y: hidden;
  -webkit-overflow-scrolling: touch;
}

.workspace-tabs {
  display: flex;
  gap: 0;
  min-width: max-content;
  height: 100%;
}

.workspace-tabs__item {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  max-width: 168px;
  flex: 0 0 auto;
  height: 100%;
  border: none;
  border-right: 1px solid var(--border-color);
  border-radius: 0;
  background: rgba(255, 255, 255, 0.03);
  color: var(--text-secondary);
  padding: 0 10px;
}

.workspace-tabs__item--active {
  background: rgba(230, 57, 124, 0.18);
  color: var(--text-primary);
}

.workspace-tabs__indicator {
  width: 7px;
  height: 7px;
  border-radius: 999px;
  background: var(--text-tertiary);
  flex: 0 0 auto;
}

.workspace-tabs__indicator--dirty {
  background: var(--accent-pink);
}

.workspace-tabs__name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 12px;
}

.workspace-tabs__close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  color: var(--text-tertiary);
  font-size: 11px;
  flex: 0 0 auto;
}

.workspace-tabs__item--active .workspace-tabs__close {
  color: var(--text-primary);
}

.workspace-tabs__save-wrap {
  flex: 0 0 auto;
  position: sticky;
  right: 0;
  z-index: 1;
  background: linear-gradient(90deg, rgba(26, 26, 29, 0) 0%, rgba(26, 26, 29, 1) 18%);
}

.workspace-tabs__save {
  width: 40px;
  height: 100%;
  border: none;
  border-left: 1px solid var(--border-color);
  border-radius: 0;
  background: var(--bg-secondary);
  color: var(--text-tertiary);
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.workspace-tabs__save :deep(svg) {
  width: 16px;
  height: 16px;
  stroke-width: 2;
}

.workspace-tabs__save--busy,
.workspace-tabs__save:active {
  color: var(--text-primary);
}

.workspace-tabs__save:disabled {
  opacity: 0.6;
}
</style>
