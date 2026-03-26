import React from 'react'
import { PiGear, PiX } from 'react-icons/pi'

import { Popover, Portal } from '@gnostudio/react'

import { observer } from 'mobx-react-lite'

import { useStore } from '@/contexts'
import { css, cx } from '@/styled-system/css'
import { stack } from '@/styled-system/patterns'
import { button, input, popover } from '@/styled-system/recipes'

export const SettingsPopover: React.FC = observer(() => {
  const store = useStore()
  const popoverStyles = popover()
  const selectStyles = input({ size: 'lg' })

  const handleKeybindingsModeChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    store.settings.setEditorMode(e.target.value as any)
  }

  return (
    <Popover.Root autoFocus modal lazyMount unmountOnExit positioning={{ offset: { mainAxis: 8 } }}>
      <Popover.Trigger className={cx(css({ paw: 'Click+Header+Settings' }), button({ variant: 'ghost' }))}>
        <PiGear size={24} />
      </Popover.Trigger>
      <Portal>
        <Popover.Positioner>
          <Popover.Content className={popoverStyles.content}>
            <Popover.Arrow className={popoverStyles.arrow}>
              <Popover.ArrowTip className={popoverStyles.arrowTip} />
            </Popover.Arrow>
            <Popover.CloseTrigger className={popoverStyles.close}>
              <PiX size={16} />
            </Popover.CloseTrigger>
            <Popover.Title className={popoverStyles.title}>Settings</Popover.Title>

            <div className={stack({ mt: 4 })}>
              <strong>Input Mode</strong>
              <select
                value={store.settings.editorMode}
                disabled={store.editor.isLoading}
                onChange={handleKeybindingsModeChange}
                className={cx(css({ paw: 'Select+Input+Settings' }), selectStyles.root, css({ w: 'full' }))}
              >
                <option value="emacs">Emacs</option>
                <option value="vim">Vim</option>
                <option value="classic">Classic</option>
              </select>
            </div>

            <div className={stack({ mt: 4 })}>
              <strong>Experimental</strong>
              <label className={css({ display: 'flex', alignItems: 'center', gap: 2 })}>
                <input
                  type="checkbox"
                  checked={store.settings.enableFunctionGutter}
                  onChange={(e) => store.settings.setFunctionGutterEnabled(e.target.checked)}
                />
                <span>Gutter run icons</span>
              </label>
            </div>
          </Popover.Content>
        </Popover.Positioner>
      </Portal>
    </Popover.Root>
  )
})
