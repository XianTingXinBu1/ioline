<script setup lang="ts">
defineEmits<{
  openWorkspacePicker: []
}>()
</script>

<template>
  <div class="empty-state">
    <div class="empty-state__card">
      <div class="empty-state__title">未选择工作区</div>
      <div class="empty-state__desc">先打开一个临时文本页。点击下方入口后，再从弹窗中选择当前工作区。</div>
      <button class="empty-state__primary" @click="$emit('openWorkspacePicker')">点击以选择工作区</button>
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

.empty-state__primary {
  width: 100%;
  margin-top: 12px;
  border: 1px solid var(--border-color);
  color: var(--text-primary);
  border-radius: 10px;
  padding: 10px 12px;
  font-size: 13px;
  background: rgba(230, 57, 124, 0.12);
}
</style>
