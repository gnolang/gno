import type { LanguageServerClient } from '@gnostudio/codemirror-lsp'

import type { EditorState, StateEffect } from '@codemirror/state'
import type { EditorView, ViewUpdate } from '@codemirror/view'
import { getCM } from '@replit/codemirror-vim'

import type { Callback, CursorPosition, EditorController, SelectionRange } from '../common/controller'
import type { InputListener, InputMode } from '../common/layout'
import { addHighlight, clearHighlightsEffect, isHighlighted } from './extensions/highlight'
import { resetLspExtensionEffect } from './extensions/lsp'
import { readOnlyEffect } from './extensions/readonly'
import { getStateData, startLspPluginEffect, updateStateDataEffect, type StateData } from './state'
import type { FormatResult } from './types'
import { docFromString, mapMarkersToDiagnostics } from './utils/doc'

/**
 * Provides offscreen control over CodeMirror editor instance in store or outside of a component.
 * Abstracts editor logic to simplify monaco migration.
 *
 * @see EditorController
 */
export class CMEditorController implements EditorController {
  private isFormatting = false
  private currentInputMode: InputMode = 'classic'
  private lspClient: LanguageServerClient | null = null

  onCursorPositionChange?: Callback<CursorPosition>
  onGutterClick?: Callback<SelectionRange>
  onInputModeChange?: Callback<InputListener | null>
  onLoadingStateChange?: Callback<boolean>

  constructor(
    private readonly editor: EditorView,
    private formatter?: (doc: string) => Promise<FormatResult>,
  ) {
    // Empty line to avoid corrupted formatting by prettier that linter doesn't line.
  }

  private async doFormat() {
    try {
      if (!this.formatter) {
        return
      }

      const { doc } = this.editor.state
      const { text, markers } = await this.formatter(doc.toString())
      const effects: Array<StateEffect<any>> = [readOnlyEffect(false)]

      if (markers) {
        effects.push(
          updateStateDataEffect.of({
            diagnostics: mapMarkersToDiagnostics(doc, markers),
          }),
        )
      }

      const { anchor, head } = this.editor.state.selection.main

      // Always update document contents to trigger linter render.
      const transaction = {
        changes: {
          from: 0,
          to: doc.length,
          insert: text ? docFromString(text) : doc.toString(),
        },
        effects,
      }

      this.editor.dispatch(transaction)

      const hasSelection = anchor !== head

      // Restore selection if it was there before formatting.
      if (hasSelection) {
        this.editor.dispatch({
          selection: {
            anchor,
            head: Math.min(head, text?.length ?? doc.length, doc.length),
          },
        })
      }
    } catch (err) {
      // TODO: show error as marker data
      console.error('error during formatting: ', err)
    } finally {
      this.isFormatting = false
      this.onLoadingStateChange?.(false)
    }
  }

  get doc() {
    return this.editor.state.doc.toString()
  }

  focus() {
    this.editor.focus()
  }

  setLspClient(client: LanguageServerClient | null) {
    this.lspClient = client
  }

  restartLanguageServer() {
    if (!this.lspClient || !this.editor) {
      return
    }

    const stateData = getStateData(this.editor.state)
    if (!stateData.isInitialised || !stateData.fileName) {
      console.warn('Cannot restart LSP plugin as editor state is not configured yet.')
      return
    }

    if (stateData.isLoading) {
      // Document isn't ready, thus LSP plugin is disabled.
      this.lspClient.reconnect()
      return
    }

    this.editor.dispatch({
      effects: resetLspExtensionEffect(),
    })

    requestAnimationFrame(() => {
      if (!this.lspClient || !this.editor) {
        return
      }

      this.lspClient.reconnect()
      const effect = startLspPluginEffect(this.editor.state, this.lspClient)
      if (effect) {
        this.editor.dispatch({
          effects: effect,
        })
      }
    })
  }

  setFormatter(formatter?: (doc: string) => Promise<FormatResult>) {
    this.formatter = formatter
  }

  formatDocument() {
    if (this.isFormatting || !this.formatter) {
      return
    }

    // read-only effect should be done in sync to make `doFormat` operate
    // on a latest doc state and guarantee order of transactions.
    this.isFormatting = true
    this.editor.dispatch({
      effects: readOnlyEffect(true),
    })

    this.onLoadingStateChange?.(true)
    void this.doFormat()
  }

  highlightLine(line: number) {
    const isToggle = isHighlighted(this.editor.state, line)
    this.removeHighlight()
    if (isToggle) {
      return
    }

    const linePos = this.editor.state.doc.line(line).from
    this.editor.dispatch({
      effects: addHighlight.of(linePos),
    })
  }

  removeHighlight() {
    this.editor.dispatch({
      effects: clearHighlightsEffect(),
    })
  }

  dispose() {
    // unused
  }

  /**
   * Notifies `onCursorPositionChange` subscriber if cursor position has been changed.
   */
  broadcastCursorState(state: EditorState | undefined) {
    if (!state) {
      this.onCursorPositionChange?.({ lineNumber: 0, column: 0 })
      return
    }

    // See: https://discuss.codemirror.net/t/get-cursor-position-line-column/6519/3
    const { doc, selection } = state
    const cursor = doc.lineAt(selection.main.head)

    const lineNumber = cursor.number
    const column = selection.main.head - cursor.from
    this.onCursorPositionChange?.({ lineNumber, column })
  }

  /**
   * Broadcasts events to listeners based on editor state update.
   */
  handleViewUpdate({ selectionSet, state }: ViewUpdate) {
    if (selectionSet) {
      this.broadcastCursorState(state)
    }
  }

  /**
   * Calls `onInputModeChange` subscribers if input mode is changed in StateData delta.
   */
  checkInputModeChanges(changes: Partial<StateData>) {
    const { inputMode } = changes
    if (!inputMode || this.currentInputMode === inputMode) {
      // skip if change triggered when compartment is updated in for a different file.
      return
    }

    this.currentInputMode = inputMode
    let listener: InputListener | null = null
    switch (changes.inputMode) {
      case 'vim': {
        const vimCm = getCM(this.editor)
        listener = {
          mode: 'vim',
          onModeChange: (cb) => vimCm?.on('vim-mode-change', (e: { mode: string }) => cb(e.mode)),
          onKeyPress: (cb) => vimCm?.on('vim-keypress', cb),
          onCommandDone: (cb) => vimCm?.on('vim-command-done', cb),
          onDispose: (cb) => vimCm?.on('dispose', cb),
        }
        break
      }
      case 'emacs':
        // TODO: investigate how to hook-up to emacs events
        listener = {
          mode: 'emacs',
          onDidMarkChange: (_cb) => {
            // TODO
          },
          onDidChangeKey: (_cb) => {
            // TODO
          },
        }
        break
    }

    this.onInputModeChange?.(listener)
  }
}
