import type { Completion } from '@codemirror/autocomplete'
import type { Diagnostic as CMDiagnostic } from '@codemirror/lint'
import type { Text } from '@codemirror/state'
import { DiagnosticSeverity, type Diagnostic } from 'vscode-languageserver-protocol'

import { classNames } from './styles'

export interface CursorPosition {
  line: number
  character: number
}

export const useLast = (values: readonly any[]) => values.reduce((_, v) => v, '')

export const isNil = <T>(v: T | null | undefined): v is undefined | null => typeof v === 'undefined' || v === null

export const isNotNil = <T>(v: T | null | undefined): v is T => !isNil(v)

export const posToOffset = (doc: Text, pos: { line: number; character: number }) => {
  if (pos.line >= doc.lines) return
  const offset = doc.line(pos.line + 1).from + pos.character
  if (offset > doc.length) return
  return offset
}

export const offsetToPos = (doc: Text, offset: number): CursorPosition => {
  const line = doc.lineAt(offset)
  return {
    line: line.number - 1,
    character: offset - line.from,
  }
}

export const toSet = (chars: Set<string>) => {
  let preamble = ''
  let flat = Array.from(chars).join('')
  const words = /\w/.test(flat)
  if (words) {
    preamble += '\\w'
    flat = flat.replace(/\w/g, '')
  }
  return `[${preamble}${flat.replace(/[^\w\s]/g, '\\$&')}]`
}

export const prefixMatch = (options: Completion[]) => {
  const first = new Set<string>()
  const rest = new Set<string>()

  for (const { apply } of options) {
    const [initial, ...restStr] = apply as string
    first.add(initial)
    for (const char of restStr) {
      rest.add(char)
    }
  }

  const source = `${toSet(first)}${toSet(rest)}*$`
  return [new RegExp(`^${source}`), new RegExp(source)]
}

export const formatSeverity = (severity: DiagnosticSeverity | undefined) => {
  switch (severity) {
    case DiagnosticSeverity.Error:
      return 'error'
    case DiagnosticSeverity.Warning:
      return 'warning'
    case DiagnosticSeverity.Information:
      return 'info'
    case DiagnosticSeverity.Hint:
      return 'info'
  }

  return undefined
}

type DiagnosticRenderFunc = CMDiagnostic['renderMessage']
const getDiagnosticRenderer = ({ code, source, message }: Diagnostic): DiagnosticRenderFunc | undefined => {
  if (!source && !code) {
    return undefined
  }

  return () => {
    const elem = document.createDocumentFragment()
    const msgNode = document.createElement('span')
    msgNode.className = classNames.diagnostic.message
    msgNode.innerText = message

    const codeNode = document.createElement('code')
    codeNode.className = classNames.diagnostic.code

    if (source) {
      codeNode.innerText = code ? `${source}(${code})` : source
    } else if (code) {
      codeNode.innerText = code.toString()
    } else {
      // Should not happen.
      return msgNode
    }

    elem.append(msgNode, codeNode)
    return elem
  }
}

/**
 * Maps LSP diagnostic messages to CodeMirror diagnostics.
 * @param doc Current document to map position.
 * @param diagnostics Diagnostics list
 */
export const mapDiagnostics = (doc: Text, diagnostics: Diagnostic[]): CMDiagnostic[] =>
  diagnostics
    .map((diag) => ({
      from: posToOffset(doc, diag.range.start),
      to: posToOffset(doc, diag.range.end),
      severity: formatSeverity(diag.severity),
      message: diag.message,
      renderMessage: getDiagnosticRenderer(diag),
    }))
    .filter((d): d is Required<CMDiagnostic> => isNotNil<number>(d.from) && isNotNil<number>(d.to))
    .sort((a, b) => {
      switch (true) {
        case a.from < b.from:
          return -1
        case a.from > b.from:
          return 1
      }
      return 0
    })
