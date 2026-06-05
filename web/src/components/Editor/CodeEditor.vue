<script setup lang="ts">
/**
 * CodeMirror 6 编辑器封装组件
 * 纯文本模式（无语法高亮）· 莫奈暗色主题
 */
import { ref, watch, onMounted, onUnmounted, type Ref } from 'vue'
import { EditorView, basicSetup } from 'codemirror'
import { Compartment, EditorSelection, EditorState, type Extension } from '@codemirror/state'
import { highlightActiveLineGutter, keymap, lineNumbers } from '@codemirror/view'
import { defaultKeymap, indentWithTab } from '@codemirror/commands'
import { useZoom } from '@/composables/useZoom'

const { fontSize, bindPinch, unbindPinch, onChange } = useZoom()
const fontSizeCompartment = new Compartment()
let unwatchZoom: (() => void) | null = null

const props = withDefaults(
  defineProps<{
    modelValue: string
    readonly?: boolean
  }>(),
  {
    modelValue: '',
    readonly: false,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
  focusChange: [focused: boolean]
  contentInput: []
}>()

const editorRef: Ref<HTMLDivElement | null> = ref(null)
let view: EditorView | null = null

function createFontSizeExtension(size: number): Extension {
  return EditorView.theme({
    '&': {
      fontSize: `${size}px`,
    },
    '.cm-content': {
      fontSize: `${size}px`,
      lineHeight: '1.65',
    },
    '.cm-line': {
      lineHeight: '1.65',
    },
    '.cm-gutters': {
      fontSize: `${size}px`,
    },
    '.cm-gutterElement': {
      fontSize: `${size}px`,
      lineHeight: '1.65',
    },
  })
}

function buildExtensions(): Extension[] {
  const exts: Extension[] = [
    basicSetup,
    lineNumbers(),
    highlightActiveLineGutter(),
    fontSizeCompartment.of(createFontSizeExtension(fontSize.value)),
    keymap.of([...defaultKeymap, indentWithTab]),
  ]

  if (props.readonly) {
    exts.push(EditorView.editable.of(false))
  }

  return exts
}

function createEditor() {
  if (!editorRef.value) return

  const state = EditorState.create({
    doc: props.modelValue,
    extensions: buildExtensions(),
  })

  view = new EditorView({
    state,
    parent: editorRef.value,
    dispatchTransactions: (trs) => {
      view?.update(trs)
      if (trs.some((tr) => tr.docChanged)) {
        const value = view?.state.doc.toString() ?? ''
        emit('update:modelValue', value)
        emit('contentInput')
      }
    },
  })

  view.dom.addEventListener('focusin', () => emit('focusChange', true))
  view.dom.addEventListener('focusout', () => emit('focusChange', false))

  unwatchZoom = onChange((size) => {
    view?.dispatch({
      effects: fontSizeCompartment.reconfigure(createFontSizeExtension(size)),
    })
    view?.requestMeasure()
  })
}

function destroyEditor() {
  unwatchZoom?.()
  unwatchZoom = null
  view?.destroy()
  view = null
}

watch(
  () => props.modelValue,
  (newVal) => {
    if (view && newVal !== view.state.doc.toString()) {
      view.dispatch({
        changes: {
          from: 0,
          to: view.state.doc.length,
          insert: newVal,
        },
      })
    }
  },
)

watch(
  () => props.readonly,
  () => {
    destroyEditor()
    createEditor()
  },
)

onMounted(() => {
  createEditor()
  if (editorRef.value) bindPinch(editorRef.value)
})
onUnmounted(() => {
  if (editorRef.value) unbindPinch(editorRef.value)
  destroyEditor()
})

function insertText(text: string) {
  if (!view || props.readonly) return
  const sel = view.state.selection.main
  view.dispatch({
    changes: { from: sel.from, to: sel.to, insert: text },
    selection: EditorSelection.cursor(sel.from + text.length),
  })
  view.focus()
}

function triggerKey(name: 'Tab' | 'Escape' | 'Enter', modifiers?: { ctrl?: boolean; alt?: boolean }) {
  if (!view) return
  view.focus()
  const event = new KeyboardEvent('keydown', {
    key: name === 'Escape' ? 'Escape' : name,
    code: name === 'Tab' ? 'Tab' : name === 'Enter' ? 'Enter' : 'Escape',
    ctrlKey: Boolean(modifiers?.ctrl),
    altKey: Boolean(modifiers?.alt),
    bubbles: true,
    cancelable: true,
  })
  view.contentDOM.dispatchEvent(event)

  if (event.defaultPrevented) {
    return
  }

  if (name === 'Tab') {
    insertText('\t')
    return
  }

  if (name === 'Enter' && !props.readonly) {
    insertText('\n')
  }
}

defineExpose({
  getSelection(): string {
    if (!view) return ''
    const sel = view.state.selection.main
    return view.state.doc.sliceString(sel.from, sel.to)
  },
  focus() {
    view?.focus()
  },
  insertText,
  triggerKey,
})
</script>

<template>
  <div ref="editorRef" class="code-editor"></div>
</template>

<style scoped>
/* ============================================
   CodeMirror 高级暗色调 · 炭黑 + 粉色撞色
   ============================================ */

/***** 容器 *****/
.code-editor {
  height: 100%;
  overflow: hidden;
  background: #1a1a1d;
}

.code-editor :deep(.cm-editor) {
  background: #1a1a1d;
  height: 100%;
}

.code-editor :deep(.cm-scroller) {
  overflow: auto;
  line-height: 1.65;
}

/***** 光标 / 选区 *****/
.code-editor :deep(.cm-content) {
  caret-color: #ffffff !important;
  padding: 8px 0;
}

.code-editor :deep(.cm-line) {
  padding: 0 6px;
}

.code-editor :deep(.cm-cursor) {
  border-left: 2px solid #ffffff !important;
}

.code-editor :deep(.cm-selectionBackground) {
  background: rgba(230, 57, 124, 0.18) !important;
}

.code-editor :deep(.cm-focused .cm-selectionBackground) {
  background: rgba(230, 57, 124, 0.24) !important;
}

/***** 行高亮 — 极其克制 *****/
.code-editor :deep(.cm-activeLine) {
  background: rgba(230, 57, 124, 0.08);
}

/***** 行号栏 *****/
.code-editor :deep(.cm-gutters) {
  background: #17171a !important;
  color: #7e7784;
  border: none !important;
  border-right: 1px solid rgba(230, 57, 124, 0.10) !important;
}

.code-editor :deep(.cm-lineNumbers) {
  min-width: 0 !important;
  width: fit-content !important;
  text-align: right;
}

.code-editor :deep(.cm-lineNumbers .cm-gutterElement) {
  display: block;
  width: 100%;
  box-sizing: border-box;
  line-height: 1.65;
  padding-left: 2px;
  padding-right: 0;
  color: #7e7784;
  text-align: right;
}

/***** 行号栏高亮行 *****/
.code-editor :deep(.cm-activeLineGutter) {
  background: rgba(230, 57, 124, 0.08) !important;
  color: #e6397c !important;
}

/***** 折叠 gutter 仍隐藏 *****/
.code-editor :deep(.cm-foldGutter) {
  display: none !important;
  width: 0 !important;
  min-width: 0 !important;
  padding: 0 !important;
  margin: 0 !important;
  overflow: hidden !important;
}

/***** 括号匹配 *****/
.code-editor :deep(.cm-matchingBracket) {
  background: rgba(230, 57, 124, 0.14);
  outline: 1px solid rgba(230, 57, 124, 0.26);
}

/***** 搜索 *****/
.code-editor :deep(.cm-panel) {
  background: #202026;
  color: #e8e6ea;
}

.code-editor :deep(.cm-search) {
  background: #202026;
  border-bottom: 1px solid #34343b;
  padding: 8px;
}

.code-editor :deep(.cm-search input) {
  background: #2a2a31;
  color: #e8e6ea;
  border: 1px solid #34343b;
  border-radius: 4px;
  padding: 4px 8px;
}

.code-editor :deep(.cm-search button) {
  background: transparent;
  color: #e6397c;
  border: 1px solid #34343b;
  border-radius: 4px;
  padding: 4px 8px;
}

/***** 提示框 *****/
.code-editor :deep(.cm-tooltip) {
  background: #202026 !important;
  color: #e8e6ea;
  border: 1px solid #34343b;
  border-radius: 6px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.5);
}

.code-editor :deep(.cm-tooltip-autocomplete li[aria-selected]) {
  background: rgba(230, 57, 124, 0.14);
}

/***** 去焦点环 *****/
.code-editor :deep(.cm-editor.cm-focused) {
  outline: none;
}
</style>
