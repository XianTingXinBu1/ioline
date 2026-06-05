<script setup lang="ts">
withDefaults(
  defineProps<{
    visible: boolean
    message?: string | null
  }>(),
  {
    visible: false,
    message: '',
  },
)
</script>

<template>
  <transition name="toast-fade">
    <div v-if="visible && message" class="workspace-toast" role="status" aria-live="polite">
      {{ message }}
    </div>
  </transition>
</template>

<style scoped>
.workspace-toast {
  position: fixed;
  left: 50%;
  bottom: calc(var(--statusbar-height) + 14px);
  transform: translateX(-50%);
  max-width: min(88vw, 420px);
  padding: 10px 14px;
  border-radius: 999px;
  background: rgba(42, 42, 49, 0.96);
  border: 1px solid rgba(255, 255, 255, 0.08);
  color: var(--text-primary);
  font-size: 12px;
  line-height: 1.5;
  text-align: center;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.28);
  z-index: 40;
}

.toast-fade-enter-active,
.toast-fade-leave-active {
  transition: opacity 0.18s ease, transform 0.18s ease;
}

.toast-fade-enter-from,
.toast-fade-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(6px);
}
</style>
