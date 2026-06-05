<script setup lang="ts">
import type { SidebarEntry } from './types'

defineProps<{
  entries: SidebarEntry[]
  currentFile: string
}>()

const indentSize = 14

const emit = defineEmits<{
  select: [entry: SidebarEntry]
}>()

function handleSelect(entry: SidebarEntry) {
  emit('select', entry)
}
</script>

<template>
  <div class="file-tree__list">
    <button
      v-for="entry in entries"
      :key="entry.path"
      class="file-tree__item"
      :class="{
        'file-tree__item--active': entry.kind === 'file' && currentFile === entry.path,
        'file-tree__item--directory': entry.kind === 'directory',
        'file-tree__item--file': entry.kind === 'file',
      }"
      @click="handleSelect(entry)"
    >
      <span class="file-tree__name" :style="{ paddingLeft: `${entry.depth * indentSize}px` }">
        {{ entry.name }}
      </span>
    </button>
  </div>
</template>

<style scoped>
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
  background: rgba(230, 57, 124, 0.10);
  color: var(--text-primary);
}

.file-tree__name {
  display: block;
  width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
