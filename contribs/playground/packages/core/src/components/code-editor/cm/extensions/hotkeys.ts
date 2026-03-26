import { Prec } from '@codemirror/state'
import { keymap } from '@codemirror/view'

import { ContextMenuAction } from '../../common/actions'
import { newDocReplaceChange } from '../utils/doc'
import { readOnlyEffect } from './readonly'

export interface HotkeyActions {
  onSave?: ((content: string) => Promise<string>) | ((content: string) => void)
  onFormat?: () => void | Promise<void>
  onContextMenuAction?: (action: ContextMenuAction) => void
}

/**
 * Returns a new hotkey plugin that handles editor settings.
 */
export const newHotKeyHandler = ({ onSave, onFormat, onContextMenuAction }: HotkeyActions) =>
  Prec.highest(
    keymap.of([
      {
        key: 'Mod-s',
        preventDefault: true,
        run: (view) => {
          // Explicit save handler may supply changed text (e.g format on save).
          // Editor should remain read-only during format.
          const changesPromise = onSave?.(view.state.doc.toString())
          if (!changesPromise) {
            return true
          }

          view.dispatch({
            effects: readOnlyEffect(true),
          })
          changesPromise
            .then((changes) => {
              view.dispatch({
                changes: newDocReplaceChange(view.state, changes),
              })
            })
            .finally(() => {
              view.dispatch({
                effects: readOnlyEffect(false),
              })
            })

          return true
        },
      },
      {
        key: 'Mod-Enter',
        preventDefault: true,
        run: () => {
          onContextMenuAction?.(ContextMenuAction.RunLastAction)
          return true
        },
      },
      {
        key: 'Mod-Shift-Enter',
        preventDefault: true,
        run: () => {
          onContextMenuAction?.(ContextMenuAction.OpenRunPrompt)
          return true
        },
      },
      {
        // 'Shift-Alt-f' doesn't work for some unknown reason.
        key: 'Mod-Shift-f',
        run: () => {
          void onFormat?.()
          return true
        },
      },
    ]),
  )
