import type { Completion, CompletionContext, CompletionResult } from '@codemirror/autocomplete'
import { setDiagnostics } from '@codemirror/lint'
import { type ChangeSpec } from '@codemirror/state'
import type { EditorView, PluginValue, Tooltip, ViewUpdate } from '@codemirror/view'
import { type Highlighter } from '@lezer/highlight'
import type { CompletionItem, CompletionTriggerKind, PublishDiagnosticsParams } from 'vscode-languageserver-protocol'

import type { LanguageServerClient, Notification } from './client'
import { client, documentUri, languageId } from './facets'
import {
  classNames,
  getCompletionItemType,
  isNil,
  isNotNil,
  mapDiagnostics,
  MarkupRenderer,
  posToOffset,
  prefixMatch,
  renderCompletionDoc,
  type CursorPosition,
} from './utils'

const changesDelay = 500

interface TriggerContext {
  triggerKind: CompletionTriggerKind
  triggerCharacter: string | undefined
}

export class LanguageServerPlugin implements PluginValue {
  public client: LanguageServerClient

  private readonly documentUri: string
  private readonly languageId: string
  private documentVersion: number

  private readonly docRenderer: MarkupRenderer
  private changesTimeout: number

  constructor(
    private readonly view: EditorView,
    highlighter?: Highlighter,
  ) {
    this.docRenderer = new MarkupRenderer(highlighter)
    this.client = this.view.state.facet(client)
    this.documentUri = this.view.state.facet(documentUri)
    this.languageId = this.view.state.facet(languageId)
    this.documentVersion = 0
    this.changesTimeout = 0

    this.client.attachPlugin(this)

    void this.initialize()
  }

  update({ docChanged }: ViewUpdate) {
    if (!docChanged) return
    if (this.changesTimeout) clearTimeout(this.changesTimeout)
    this.changesTimeout = self.setTimeout(() => {
      void this.sendChange({
        documentText: this.view.state.doc.toString(),
      })
    }, changesDelay)
  }

  destroy() {
    this.client.detachPlugin(this)
  }

  private supportsResolveProvider() {
    return this.client.capabilities?.completionProvider?.resolveProvider
  }

  async initialize() {
    const documentText = this.view.state.doc.toString()

    if (this.client.initializePromise) {
      await this.client.initializePromise
    }

    await this.client.textDocumentDidOpen({
      textDocument: {
        uri: this.documentUri,
        languageId: this.languageId,
        text: documentText,
        version: this.documentVersion,
      },
    })
  }

  async sendChange({ documentText }: { documentText: string }) {
    if (!this.client.ready) return
    try {
      await this.client.textDocumentDidChange({
        textDocument: {
          uri: this.documentUri,
          version: this.documentVersion++,
        },
        contentChanges: [{ text: documentText }],
      })
    } catch (e) {
      console.error(e)
    }
  }

  async format() {
    if (!this.client.ready) {
      return
    }

    const result = await this.client.textDocumentFormatting({
      textDocument: { uri: this.documentUri },
      options: {
        // TODO: Make this configurable
        tabSize: this.view.state.tabSize,
        insertSpaces: false,
      },
    })

    if (!result.length) {
      return
    }

    const changes: ChangeSpec[] = []
    result.forEach(({ newText, range }) => {
      const from = posToOffset(this.view.state.doc, range.start)
      const to = posToOffset(this.view.state.doc, range.end)

      if (isNotNil(from) && isNotNil(to)) {
        changes.push({
          from,
          to,
          insert: newText,
        })
      }
    })

    const newTextLength = result[0].newText.length
    const { head } = this.view.state.selection.main
    const newSelectionAnchor = head > newTextLength ? newTextLength : head

    this.view.dispatch({
      changes,
      selection: { anchor: newSelectionAnchor },
    })
  }

  async requestDiagnostics(view: EditorView) {
    if (!this.client.ready) {
      return
    }

    const result = await this.client.textDocumentLinting({
      textDocument: {
        uri: this.documentUri,
        version: this.documentVersion++,
      },
      contentChanges: [{ text: view.state.doc.toString() }],
    })

    if (!result || result.uri !== this.documentUri) return

    const diagnostics = mapDiagnostics(this.view.state.doc, result.diagnostics)
    return diagnostics
  }

  async requestHoverTooltip(
    view: EditorView,
    { line, character }: { line: number; character: number },
  ): Promise<Tooltip | null> {
    if (!this.client.ready || !this.client.capabilities?.hoverProvider) return null
    const result = await this.client.textDocumentHover({
      textDocument: { uri: this.documentUri },
      position: { line, character },
    })

    if (!result?.contents) {
      return null
    }

    const { contents, range } = result
    let pos: number | undefined = posToOffset(view.state.doc, { line, character })
    let end: number | undefined
    if (range) {
      pos = posToOffset(view.state.doc, range.start)
      end = posToOffset(view.state.doc, range.end)
    }

    if (isNil(pos)) {
      return null
    }

    const dom = document.createElement('div')
    dom.classList.add(classNames.tooltip)
    this.docRenderer.renderContents(dom, contents)

    const tooltip: Tooltip = { pos, end, create: (_view) => ({ dom }), above: true }
    return tooltip
  }

  async requestCompletion(
    context: CompletionContext,
    { line, character }: CursorPosition,
    { triggerKind, triggerCharacter }: TriggerContext,
  ): Promise<CompletionResult | null> {
    if (!this.client.ready || !this.client.capabilities?.completionProvider) return null
    await this.sendChange({
      documentText: context.state.doc.toString(),
    })

    const result = await this.client.textDocumentCompletion({
      textDocument: { uri: this.documentUri },
      position: { line, character },
      context: {
        triggerKind,
        triggerCharacter,
      },
    })

    if (!result) return null

    const isIncomplete = 'isIncomplete' in result ? result.isIncomplete : false
    const items = 'items' in result ? result.items : result

    let options = items.map((completionItem) => {
      const { detail, label, kind, insertText, textEdit, sortText, filterText } = completionItem
      const completion: Completion & {
        filterText: string
        sortText?: string
        apply: string
      } = {
        label,
        detail,
        apply: textEdit?.newText ?? insertText ?? label,
        type: kind && getCompletionItemType(kind),
        sortText: sortText ?? label,
        filterText: filterText ?? label,
        info: this.newCompletionItemResolver(completionItem, isIncomplete),
      }
      return completion
    })

    const [_span, match] = prefixMatch(options)
    const token = context.matchBefore(match)
    let { pos } = context

    if (token) {
      pos = token.from
      const word = token.text.toLowerCase()
      if (/^\w+$/.test(word)) {
        options = options
          .filter(({ filterText }) => filterText.toLowerCase().startsWith(word))
          .sort(({ apply: a }, { apply: b }) => {
            switch (true) {
              case a.startsWith(token.text) && !b.startsWith(token.text):
                return -1
              case !a.startsWith(token.text) && b.startsWith(token.text):
                return 1
            }
            return 0
          })
      }
    }
    return {
      from: pos,
      options,
    }
  }

  private newCompletionItemResolver(item: CompletionItem, isIncomplete: boolean): Completion['info'] {
    if (isIncomplete) {
      if (!this.supportsResolveProvider()) {
        return undefined
      }

      return async () => {
        const resolved = await this.client.completionItemResolve(item)
        return renderCompletionDoc(this.docRenderer, resolved.documentation ?? item.documentation)
      }
    }

    if (item.documentation) {
      return () => renderCompletionDoc(this.docRenderer, item.documentation)
    }

    return undefined
  }

  handleNotification(notification: Notification) {
    try {
      switch (notification.method) {
        case 'textDocument/publishDiagnostics':
          this.handleDiagnostics(notification.params)
          break
        default:
          console.warn('processNotification: unknown notification type', notification)
      }
    } catch (err) {
      console.error(err)
    }
  }

  handleDiagnostics(params: PublishDiagnosticsParams) {
    if (params.uri !== this.documentUri) return

    const diagnostics = mapDiagnostics(this.view.state.doc, params.diagnostics)
    this.view.dispatch(setDiagnostics(this.view.state, diagnostics))
  }
}
