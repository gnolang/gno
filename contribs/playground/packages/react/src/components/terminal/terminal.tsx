import React, { useEffect, useRef, type ComponentPropsWithoutRef } from 'react'

import { type TerminalStore } from '@gnostudio/core'

import { ark, Presence } from '@ark-ui/react'
import { observer } from 'mobx-react-lite'

import { TerminalProvider, useTerminalContext } from './terminal-context'

import '../../../../core/src/stores/terminal/terminal.css'

export type TerminalProps = { store: TerminalStore; onClose?: () => void } & ComponentPropsWithoutRef<typeof ark.div>

const Root: React.FC<TerminalProps> = observer((props) => {
  const { store, onClose, children, ...rest } = props
  const domRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!domRef.current) return
    store.mount(domRef.current)
  }, [store, domRef])

  return (
    <TerminalProvider value={{ store, onClose }}>
      <Presence present={store.isOpen} style={{ height: '100%' }}>
        <ark.div data-scope="terminal" data-part="root" {...rest}>
          <div ref={domRef} data-scope="terminal" data-part="console" />

          {children}
        </ark.div>
      </Presence>
    </TerminalProvider>
  )
})

export type TerminalCloseButtonProps = ComponentPropsWithoutRef<typeof ark.button>

const CloseButton: React.FC<TerminalCloseButtonProps> = (props) => {
  const ctx = useTerminalContext()

  const handleClose = () => {
    if (ctx.onClose) ctx.onClose()
    ctx.store.close()
  }

  return <ark.button data-scope="terminal" data-part="close" {...props} onClick={handleClose} />
}

export const Terminal = Object.assign(Root, {
  Root,
  CloseButton,
})
