<script setup lang="ts">
withDefaults(
  defineProps<{
    ctrlActive?: boolean
    altActive?: boolean
    disabled?: boolean
  }>(),
  {
    ctrlActive: false,
    altActive: false,
    disabled: false,
  },
)

const emit = defineEmits<{
  pressTab: []
  pressEsc: []
  pressEnter: []
  toggleCtrl: []
  toggleAlt: []
}>()

function keepEditorFocus(event: PointerEvent | MouseEvent) {
  event.preventDefault()
}
</script>

<template>
  <div class="quick-keys">
    <button
      class="quick-keys__btn quick-keys__btn--wide"
      :disabled="disabled"
      @pointerdown="keepEditorFocus"
      @mousedown="keepEditorFocus"
      @click="emit('pressTab')"
    >
      Tab
    </button>
    <button
      class="quick-keys__btn"
      :disabled="disabled"
      @pointerdown="keepEditorFocus"
      @mousedown="keepEditorFocus"
      @click="emit('pressEsc')"
    >
      Esc
    </button>
    <button
      class="quick-keys__btn"
      :class="{ 'quick-keys__btn--active': ctrlActive }"
      :disabled="disabled"
      @pointerdown="keepEditorFocus"
      @mousedown="keepEditorFocus"
      @click="emit('toggleCtrl')"
    >
      Ctrl
    </button>
    <button
      class="quick-keys__btn"
      :class="{ 'quick-keys__btn--active': altActive }"
      :disabled="disabled"
      @pointerdown="keepEditorFocus"
      @mousedown="keepEditorFocus"
      @click="emit('toggleAlt')"
    >
      Alt
    </button>
    <button
      class="quick-keys__btn quick-keys__btn--wide"
      :disabled="disabled"
      @pointerdown="keepEditorFocus"
      @mousedown="keepEditorFocus"
      @click="emit('pressEnter')"
    >
      Enter
    </button>
  </div>
</template>

<style scoped>
.quick-keys {
  display: grid;
  grid-template-columns: 1.15fr 1fr 1fr 1fr 1.15fr;
  gap: 6px;
  padding: 6px 10px;
  background: var(--bg-secondary);
  border-top: 1px solid var(--border-color);
}

.quick-keys__btn {
  min-height: 28px;
  border: 1px solid var(--border-color);
  border-radius: 9px;
  background: rgba(255, 255, 255, 0.03);
  color: var(--text-secondary);
  font-size: 12px;
}

.quick-keys__btn--active {
  background: rgba(230, 57, 124, 0.14);
  color: var(--text-primary);
}

.quick-keys__btn:disabled {
  opacity: 0.55;
}
</style>
