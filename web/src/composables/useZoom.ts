import { ref, readonly } from 'vue'

const MIN = 8
const MAX = 32
const DEFAULT = 14

const fontSize = ref(DEFAULT)
const listeners = new Set<(size: number) => void>()

export function useZoom() {
  function set(v: number) {
    const clamped = Math.min(MAX, Math.max(MIN, v))
    const next = Math.round(clamped * 10) / 10
    if (next === fontSize.value) return

    fontSize.value = next
    listeners.forEach((listener) => listener(next))
  }

  function onChange(listener: (size: number) => void) {
    listeners.add(listener)
    listener(fontSize.value)
    return () => listeners.delete(listener)
  }

  /* ================================================
     移动端二指捏合缩放
     ================================================ */
  let pinchStartDist = 0
  let pinchStartSize = DEFAULT
  let pinching = false

  function getTouchDist(touches: TouchList): number {
    if (touches.length < 2) return 0
    const dx = touches[0].clientX - touches[1].clientX
    const dy = touches[0].clientY - touches[1].clientY
    return Math.sqrt(dx * dx + dy * dy)
  }

  function onTouchStart(e: TouchEvent) {
    if (e.touches.length !== 2) return
    pinching = true
    pinchStartDist = getTouchDist(e.touches)
    pinchStartSize = fontSize.value
    e.preventDefault()
  }

  function onTouchMove(e: TouchEvent) {
    if (!pinching || e.touches.length !== 2) return
    const dist = getTouchDist(e.touches)
    if (dist === 0 || pinchStartDist === 0) return
    set(pinchStartSize * (dist / pinchStartDist))
    e.preventDefault()
  }

  function onTouchEnd() {
    pinching = false
  }

  function bindPinch(el: HTMLElement) {
    el.addEventListener('touchstart', onTouchStart, { passive: false })
    el.addEventListener('touchmove', onTouchMove, { passive: false })
    el.addEventListener('touchend', onTouchEnd)
    el.addEventListener('touchcancel', onTouchEnd)
  }

  function unbindPinch(el: HTMLElement) {
    el.removeEventListener('touchstart', onTouchStart)
    el.removeEventListener('touchmove', onTouchMove)
    el.removeEventListener('touchend', onTouchEnd)
    el.removeEventListener('touchcancel', onTouchEnd)
  }

  return {
    fontSize: readonly(fontSize),
    set,
    onChange,
    bindPinch,
    unbindPinch,
  }
}
