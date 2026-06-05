import { ref } from 'vue'

export function useEditorQuickKeys() {
  const ctrlActive = ref(false)
  const altActive = ref(false)

  function toggleCtrl() {
    ctrlActive.value = !ctrlActive.value
  }

  function toggleAlt() {
    altActive.value = !altActive.value
  }

  function clearModifiers() {
    ctrlActive.value = false
    altActive.value = false
  }

  return {
    altActive,
    clearModifiers,
    ctrlActive,
    toggleAlt,
    toggleCtrl,
  }
}
