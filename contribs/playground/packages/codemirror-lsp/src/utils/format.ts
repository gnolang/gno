import { go } from '@codemirror/lang-go'
import { highlightTree, type Highlighter } from '@lezer/highlight'
import markdownit from 'markdown-it'
import { CompletionItemKind, MarkupKind, type MarkedString, type MarkupContent } from 'vscode-languageserver-protocol'

import { classNames } from './styles'

type Content = MarkupContent | MarkedString | MarkedString[]

const CompletionItemKindMap = Object.fromEntries(
  Object.entries(CompletionItemKind).map(([key, value]) => [value, key]),
) as Record<CompletionItemKind, string>

/**
 * Maps CM completion type from LSP kind.
 */
export const getCompletionItemType = (k: CompletionItemKind) => {
  // Edge case - CM use 'namespace' instead of 'module'
  if (k === CompletionItemKind.Module) {
    return 'namespace'
  }

  return CompletionItemKindMap[k].toLowerCase()
}

const getMarkupKind = (entry: MarkupContent | { language: string }, fallback: MarkupKind): MarkupKind => {
  if ('kind' in entry) {
    return entry.kind
  }

  if ('language' in entry) {
    return entry.language as MarkupKind
  }

  return fallback
}

const normalizeMarkupContent = (content: Content): MarkupContent => {
  if (Array.isArray(content)) {
    return content.reduce(
      (acc: MarkupContent, item: MarkupContent | MarkedString): MarkupContent => {
        const { kind, value } = acc
        switch (typeof item) {
          case 'string':
            return {
              kind,
              value: value + '\n\n' + item,
            }
          case 'object':
            return {
              kind: getMarkupKind(item, kind),
              value: value + '\n\n' + item.value,
            }
          default:
            return acc
        }
      },
      { kind: MarkupKind.PlainText, value: '' },
    )
  }

  if (typeof content === 'string') {
    return { kind: MarkupKind.PlainText, value: content }
  }

  return {
    kind: getMarkupKind(content, MarkupKind.PlainText),
    value: content.value,
  }
}

export class MarkupRenderer {
  private readonly langGo = go().language
  private readonly printer: markdownit
  constructor(private readonly highlighter?: Highlighter) {
    this.printer = markdownit({
      html: false,
      breaks: true,
      highlight: (str: string, lang: string, _attrs: string) => {
        return this.highlightSource(str, lang)
      },
    })
  }

  private highlightSource(src: string, lang: string) {
    if (lang !== 'go' || !this.highlighter) {
      return `<pre class="code">${src}</pre>`
    }

    const tree = this.langGo.parser.parse(src)

    let buff = ''
    let prevEnd = 0
    highlightTree(tree, this.highlighter, (from, to, classes) => {
      // Preserve untokenized spaces
      if (prevEnd > 0) {
        const spaces = src.slice(prevEnd, from)
        if (spaces.length > 0) {
          buff += `<span>${spaces}</span>`
        }
      }

      prevEnd = to
      buff += `<span class="${classes}">${src.slice(from, to)}</span>`
    })

    return `<div class="code-highlighted">${buff}</div>`
  }

  renderContents(dst: HTMLElement, contents: Content) {
    const { kind, value } = normalizeMarkupContent(contents)
    if (kind === MarkupKind.Markdown) {
      dst.innerHTML = this.printer.render(value)
      return
    }

    const pre = document.createElement('pre')
    pre.innerText = value
    dst.appendChild(pre)
  }
}

export const renderCompletionDoc = (renderer: MarkupRenderer, doc?: Content) => {
  if (!doc) {
    return null
  }

  const node = document.createElement('div')
  node.classList.add(classNames.completionDoc)
  if (typeof doc === 'string') {
    node.innerText = doc
  } else {
    renderer.renderContents(node, doc)
  }

  return node
}
