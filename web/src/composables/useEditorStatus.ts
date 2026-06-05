import { computed, type Ref } from 'vue'
import type { FileContentResponse } from '@/api/types'

const specialFileTypes: Record<string, string> = {
  Makefile: 'makefile',
  Dockerfile: 'dockerfile',
  LICENSE: 'text',
  README: 'text',
}

function detectFileType(fileName: string): string {
  if (!fileName || fileName === '__welcome__.txt') return 'txt'
  if (specialFileTypes[fileName]) return specialFileTypes[fileName]

  const dotIndex = fileName.lastIndexOf('.')
  if (dotIndex > 0 && dotIndex < fileName.length - 1) {
    return fileName.slice(dotIndex + 1).toLowerCase()
  }

  return 'text'
}

function countLines(content: string): number {
  if (!content) return 1
  return content.split(/\r\n|\n|\r/).length
}

export function useEditorStatus(params: {
  code: Ref<string>
  currentFileName: Ref<string>
  currentFileInfo: Ref<FileContentResponse | null>
}) {
  const fileType = computed(() => detectFileType(params.currentFileName.value))
  const lineCount = computed(() => countLines(params.code.value))
  const charCount = computed(() => params.code.value.length)
  const encoding = computed(() => 'UTF-8')
  const lineEnding = computed(() => {
    const ending = params.currentFileInfo.value?.lineEnding
    return ending ? ending.toUpperCase() : 'LF'
  })

  return {
    charCount,
    encoding,
    fileType,
    lineCount,
    lineEnding,
  }
}
