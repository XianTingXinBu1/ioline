import { onBeforeUnmount, ref, watch } from 'vue'

export function useWorkspaceToast(source: { value: string | null }) {
  const visible = ref(false)
  const message = ref('')
  let timer: ReturnType<typeof setTimeout> | null = null

  function clearTimer() {
    if (timer) {
      clearTimeout(timer)
      timer = null
    }
  }

  function show(nextMessage: string) {
    clearTimer()
    message.value = nextMessage
    visible.value = true
    timer = setTimeout(() => {
      visible.value = false
      timer = null
    }, 2200)
  }

  watch(
    () => source.value,
    (next, prev) => {
      if (!next || next === prev) return
      show(next)
    },
  )

  onBeforeUnmount(() => {
    clearTimer()
  })

  return {
    toastMessage: message,
    toastVisible: visible,
  }
}
