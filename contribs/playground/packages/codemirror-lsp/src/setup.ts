import { autocompletion } from '@codemirror/autocomplete'
import { linter } from '@codemirror/lint'
import { EditorView, hoverTooltip, keymap, ViewPlugin } from '@codemirror/view'
import { CompletionTriggerKind, type WorkspaceFolder } from 'vscode-languageserver-protocol'

import type { LanguageServerClient } from './client/client'
import { client, documentUri, languageId } from './facets'
import { LanguageServerPlugin } from './plugin'
import { offsetToPos } from './utils'
import { coreStyles, type PluginTheme } from './utils/styles'

export interface LanguageServerPluginOptions {
  /**
   * LSP document language ID.
   */
  languageId: string

  /**
   * URI of active document.
   */
  documentUri: string

  /**
   * LSP client instance.
   */
  client: LanguageServerClient

  /**
   * Additional workspace folders for LSP context.
   *
   * Currently unused.
   */
  workspaceFolders?: WorkspaceFolder[]

  /**
   * Plugin styles and syntax highlighter.
   */
  theme?: PluginTheme
}

/**
 * Initializes and returns a new language server plugin extension for CM6.
 */
export function languageServerPlugin(options: LanguageServerPluginOptions) {
  let plugin: LanguageServerPlugin | null = null

  return [
    EditorView.theme(options.theme?.styleSpec ?? coreStyles),
    client.of(options.client),
    documentUri.of(options.documentUri),
    languageId.of(options.languageId),
    linter(async (view) => {
      const diagnostics = await plugin?.requestDiagnostics(view)
      return diagnostics ?? []
    }),
    ViewPlugin.define((view) => (plugin = new LanguageServerPlugin(view, options.theme?.highlighter))),
    hoverTooltip((view, pos) => plugin?.requestHoverTooltip(view, offsetToPos(view.state.doc, pos)) ?? null),
    keymap.of([
      {
        key: 'Mod-Shift-f',
        run: (_view) => {
          void plugin?.format()
          return true
        },
      },
      {
        key: 'Mod-l',
        run: (view) => {
          void plugin?.requestDiagnostics(view)
          return true
        },
      },
    ]),
    autocompletion({
      override: [
        async (context) => {
          if (plugin == null) return null

          const { state, pos, explicit } = context
          const line = state.doc.lineAt(pos)
          let trigKind: CompletionTriggerKind = CompletionTriggerKind.Invoked
          let trigChar: string | undefined
          if (
            !explicit &&
            plugin.client.capabilities?.completionProvider?.triggerCharacters?.includes(line.text[pos - line.from - 1])
          ) {
            trigKind = CompletionTriggerKind.TriggerCharacter
            trigChar = line.text[pos - line.from - 1]
          }
          if (trigKind === CompletionTriggerKind.Invoked && !context.matchBefore(/\w+$/)) {
            return null
          }
          return await plugin.requestCompletion(context, offsetToPos(state.doc, pos), {
            triggerKind: trigKind,
            triggerCharacter: trigChar,
          })
        },
      ],
    }),
  ]
}
