<script setup lang="ts">
import type { WorkspaceCandidate } from './types'

const props = withDefaults(
  defineProps<{
    candidates: WorkspaceCandidate[]
    loading?: boolean
    selectingPath?: string | null
  }>(),
  {
    loading: false,
    selectingPath: null,
  },
)

const emit = defineEmits<{
  selectWorkspace: [candidate: WorkspaceCandidate]
  refreshCandidates: []
}>()

function handleSelect(candidate: WorkspaceCandidate) {
  emit('selectWorkspace', candidate)
}
</script>

<template>
  <div class="empty-state">
    <div class="empty-state__card">
      <div class="empty-state__title">未选择工作区</div>
      <div class="empty-state__desc">点击候选目录即可作为当前工作区，随后会加载根目录文件树。</div>

      <button class="empty-state__refresh" :disabled="loading" @click="emit('refreshCandidates')">
        {{ loading ? '刷新中...' : '刷新候选目录' }}
      </button>

      <div v-if="candidates.length > 0" class="empty-state__list">
        <button
          v-for="candidate in props.candidates"
          :key="candidate.path"
          class="empty-state__item"
          :disabled="selectingPath === candidate.path"
          @click="handleSelect(candidate)"
        >
          <span class="empty-state__name">{{ candidate.name }}</span>
          <span class="empty-state__path">{{ candidate.path }}</span>
          <span class="empty-state__meta">
            {{ selectingPath === candidate.path ? '设置中...' : candidate.source }}
          </span>
        </button>
      </div>

      <div v-else class="empty-state__placeholder">暂无可用候选目录，可稍后重试。</div>
    </div>
  </div>
</template>

<style scoped>
.empty-state {
  flex: 1;
  padding: 12px;
  overflow-y: auto;
}

.empty-state__card {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 14px;
}

.empty-state__title {
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 600;
}

.empty-state__desc {
  margin-top: 8px;
  color: var(--text-secondary);
  font-size: 12px;
  line-height: 1.6;
}

.empty-state__refresh {
  margin-top: 12px;
  width: 100%;
  border: 1px solid var(--border-color);
  background: transparent;
  color: var(--text-primary);
  border-radius: 10px;
  padding: 10px 12px;
  font-size: 13px;
}

.empty-state__refresh:disabled,
.empty-state__item:disabled {
  opacity: 0.6;
}

.empty-state__list {
  margin-top: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.empty-state__item {
  width: 100%;
  border: 1px solid var(--border-color);
  background: rgba(255, 255, 255, 0.02);
  color: var(--text-primary);
  border-radius: 10px;
  padding: 10px 12px;
  text-align: left;
}

.empty-state__name {
  display: block;
  font-size: 13px;
  font-weight: 600;
}

.empty-state__path {
  display: block;
  margin-top: 4px;
  color: var(--text-secondary);
  font-size: 11px;
  line-height: 1.5;
  word-break: break-all;
}

.empty-state__meta {
  display: block;
  margin-top: 6px;
  color: var(--text-tertiary);
  font-size: 11px;
  text-transform: uppercase;
}

.empty-state__placeholder {
  margin-top: 12px;
  color: var(--text-tertiary);
  font-size: 12px;
  line-height: 1.6;
}
</style>
