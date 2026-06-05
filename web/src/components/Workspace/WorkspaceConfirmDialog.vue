<script setup lang="ts">
withDefaults(
  defineProps<{
    open: boolean
    title: string
    description: string
    confirmText?: string
    cancelText?: string
  }>(),
  {
    open: false,
    confirmText: '确认',
    cancelText: '取消',
  },
)

const emit = defineEmits<{
  confirm: []
  cancel: []
}>()
</script>

<template>
  <div v-if="open" class="confirm-dialog">
    <div class="confirm-dialog__backdrop" aria-hidden="true" @click="emit('cancel')"></div>
    <div class="confirm-dialog__panel" role="dialog" aria-modal="true" :aria-label="title">
      <div class="confirm-dialog__title">{{ title }}</div>
      <div class="confirm-dialog__desc">{{ description }}</div>
      <div class="confirm-dialog__actions">
        <button class="confirm-dialog__btn confirm-dialog__btn--secondary" @click="emit('cancel')">{{ cancelText }}</button>
        <button class="confirm-dialog__btn confirm-dialog__btn--primary" @click="emit('confirm')">{{ confirmText }}</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.confirm-dialog {
  position: fixed;
  inset: 0;
  z-index: 35;
}

.confirm-dialog__backdrop {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.42);
}

.confirm-dialog__panel {
  position: absolute;
  left: 16px;
  right: 16px;
  top: 50%;
  transform: translateY(-50%);
  border: 1px solid var(--border-color);
  background: var(--bg-secondary);
  border-radius: 14px;
  padding: 16px;
}

.confirm-dialog__title {
  color: var(--text-primary);
  font-size: 15px;
  font-weight: 600;
}

.confirm-dialog__desc {
  margin-top: 8px;
  color: var(--text-secondary);
  font-size: 12px;
  line-height: 1.6;
}

.confirm-dialog__actions {
  margin-top: 14px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}

.confirm-dialog__btn {
  border: 1px solid var(--border-color);
  border-radius: 10px;
  min-height: 38px;
  color: var(--text-primary);
  font-size: 13px;
}

.confirm-dialog__btn--secondary {
  background: transparent;
}

.confirm-dialog__btn--primary {
  background: rgba(230, 57, 124, 0.12);
}
</style>
