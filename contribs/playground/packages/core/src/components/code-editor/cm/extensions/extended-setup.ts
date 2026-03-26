import { indentWithTab } from '@codemirror/commands'
import { EditorView, keymap, lineNumbers, type ViewUpdate } from '@codemirror/view'
import { vscodeKeymap } from '@replit/codemirror-vscode-keymap'

import { newStateDataFieldExtension, type StateProps } from '../state'
import { basicSetup } from './core'
import { highlightField } from './highlight'
import { newHotKeyHandler, type HotkeyActions } from './hotkeys'
import { newInputModeCompartment } from './input'
import { newFormatErrorsRenderer } from './linter'
import { newLspCompartment } from './lsp'
import { newReadOnlyCompartment } from './readonly'
import { runGutterExtension, type onRunCallback } from './run-gutter'
import { newSyntaxCompartment } from './syntax'
import { defaultThemeStyles, newThemeCompartment } from './themes'

const keyBindings = [indentWithTab, ...vscodeKeymap]
const corePlugins = [defaultThemeStyles, highlightField, keymap.of(keyBindings)]

interface ExtendedSetupOpts {
  getState: () => StateProps
  onDocumentChange?: (doc: string) => void
  onGutterClick?: (line: number) => void
  onSave?: HotkeyActions['onSave']
  onContextMenuAction?: HotkeyActions['onContextMenuAction']
  onFormatDocument?: HotkeyActions['onFormat']
  onRunFunction?: onRunCallback
  onViewUpdate?: (update: ViewUpdate) => void
  showLineNumbers?: boolean
}

const defaultOpts: Partial<ExtendedSetupOpts> = {
  showLineNumbers: true,
}

export const extendedSetup = (opts: ExtendedSetupOpts) => {
  opts = { ...defaultOpts, ...opts }

  const changeHandlerExtension = EditorView.updateListener.of((update) => {
    if (update.docChanged) {
      opts?.onDocumentChange?.(update.state.doc.toString())
    }

    opts.onViewUpdate?.(update)
  })

  const hotkeyHandler = newHotKeyHandler({
    onSave: (text) => opts?.onSave?.(text),
    onContextMenuAction: (e) => opts?.onContextMenuAction?.(e),
    onFormat: () => opts?.onFormatDocument?.(),
  })

  const extensions = [
    hotkeyHandler, // Should be on top to avoid overlap with standard keymap!
    newReadOnlyCompartment(),
    newInputModeCompartment(opts?.getState().inputMode),
    newStateDataFieldExtension(opts.getState),
    newFormatErrorsRenderer(),
    ...basicSetup(),
    ...corePlugins,
    newSyntaxCompartment(opts?.getState().value?.path),
    newThemeCompartment(opts?.getState().theme),
    newLspCompartment(),
    changeHandlerExtension,
  ]

  if (opts.showLineNumbers) {
    extensions.unshift(
      lineNumbers({
        domEventHandlers: {
          click: (view, line) => {
            // Broadcast line number click
            const lineNumber = view.state.doc.lineAt(line.from).number
            opts?.onGutterClick?.(lineNumber)
            return true
          },
        },
      }),
    )
  }

  /**
   * Add run gutter if onRunFunction is provided
   */
  if (opts?.onRunFunction) {
    extensions.unshift(
      runGutterExtension({
        onRun: opts?.onRunFunction,
      }),
    )
  }

  return extensions
}
