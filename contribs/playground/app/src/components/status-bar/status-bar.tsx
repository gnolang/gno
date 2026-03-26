import React from 'react'
import { PiKeyboard } from 'react-icons/pi'

import { observer } from 'mobx-react-lite'

import { useStore } from '@/contexts'
import { css } from '@/styled-system/css'
import { hstack } from '@/styled-system/patterns'

import { GnoVMSelector } from './gnovm-selector'
import { StatusText } from './status-text'

const keybindingsAdapterMap: Record<string, string> = {
  classic: 'Classic',
  emacs: 'Emacs',
  vim: 'Vim',
}

export const StatusBar: React.FC = observer(() => {
  const store = useStore()
  const keybindingsStore = store.editor.keybindingsStore

  const hasKeyBuffer = keybindingsStore.keyBuffer.length > 0
  const displayMode = keybindingsStore.mode ? `--${keybindingsStore.mode.toUpperCase()}--` : null

  return (
    <div
      data-testid="statusbar"
      className={hstack({
        flexShrink: 0,
        justifyContent: 'space-between',
        alignItems: 'center',
        height: '8',
        borderTop: '1px solid',
        borderColor: 'header',
        fontSize: 'sm',
        px: 2,
      })}
    >
      <div className={css({ flexShrink: 0, flexGrow: 1, overflow: 'hidden', flexBasis: 1 })}>
        <StatusText />
      </div>

      <div className={hstack({ flexShrink: 0 })}>
        <GnoVMSelector allowChange />

        <div className={hstack({})}>
          <span>{hasKeyBuffer ? keybindingsStore.keyBuffer : displayMode}</span>

          <div className={hstack({ alignItems: 'start', gap: 1 })}>
            <PiKeyboard size={17} />
            <span>{keybindingsAdapterMap[store.settings.editorMode]}</span>
          </div>
        </div>

        {store.settings.enableFunctionGutter && (
          <span className={css({ ml: 3, fontSize: 'xs', color: 'fg.muted' })}>Experimental: Gutter run icons</span>
        )}

        <span data-testid="statusbar-position">
          Ln {store.editor.position.lineNumber}, Col {store.editor.position.column}
        </span>
      </div>
    </div>
  )
})
