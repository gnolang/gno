import React, { forwardRef, useEffect, useRef } from 'react'

import { observer } from 'mobx-react-lite'

import { useStore } from '@/contexts'
import { cx } from '@/styled-system/css'
import { input, tabs } from '@/styled-system/recipes'

interface Props {
  onSubmit: () => void
  onCancel: () => void
  onBlur: () => void
  isDisabled?: boolean
  fullWidth?: boolean
}

export const FileInputName = observer(
  forwardRef<HTMLInputElement, Props>(function InputName(props, externalRef) {
    const inputRef = useRef<HTMLInputElement | null>(null)
    const store = useStore()
    const tabsStyle = tabs()
    const inputStyles = input({ body: 'strict', size: 'xs' })

    const handleInputKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
      if (event.key === 'Escape') {
        event.preventDefault()
        props.onCancel()
      }

      if (event.key === 'Enter') {
        event.preventDefault()
        props.onSubmit()
      }

      if (event.key === 'ArrowLeft' || event.key === 'ArrowRight') {
        event.stopPropagation()
      }
    }

    useEffect(() => {
      if (typeof store.workbench.pendingFileName !== 'string') return

      // Select text without the extension
      const fileName = store.workbench.pendingFileName.split('.')[0]
      inputRef.current?.setSelectionRange(0, fileName.length)
      // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])

    return (
      <input
        ref={(el) => {
          inputRef.current = el
          if (externalRef) {
            ;(externalRef as React.MutableRefObject<HTMLInputElement | null>).current = el
          }
        }}
        data-testid="file-name-input"
        type="text"
        placeholder="Untitled"
        value={store.workbench.pendingFileName === true ? '' : (store.workbench.pendingFileName as string)}
        className={cx(tabsStyle.textInput, inputStyles.root)}
        onKeyDown={handleInputKeyDown}
        onBlur={props.onBlur}
        disabled={props.isDisabled}
        onChange={(event) => store.workbench.setPendingFileName(event.target.value)}
        style={props.fullWidth ? { width: '100%' } : undefined}
        autoFocus
      />
    )
  }),
)
