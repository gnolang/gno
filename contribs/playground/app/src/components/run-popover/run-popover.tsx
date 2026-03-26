import React, { useEffect, useRef, useState } from 'react'
import { PiCaretDownFill, PiX } from 'react-icons/pi'

import { Popover, Portal } from '@gnostudio/react'

import { observer } from 'mobx-react-lite'

import { css, cx } from '@/styled-system/css'
import { hstack, stack, visuallyHidden } from '@/styled-system/patterns'
import { button, input, link, popover } from '@/styled-system/recipes'

interface RunPopoverContentProps {
  isOpen?: boolean
  initialValue?: string
  onRunClick?: (expr: string) => void
}

const RunPopoverContent: React.FC<RunPopoverContentProps> = observer(({ onRunClick, isOpen, initialValue }) => {
  const [expr, setExpr] = useState(initialValue ?? '')
  const [isValid, setIsValid] = useState(false)
  const popoverStyles = popover({ size: 'sm' })
  const inputStyles = input({ size: 'lg' })
  const inputRef = useRef<HTMLInputElement | null>(null)

  const onSubmit = (event: React.SyntheticEvent) => {
    event.preventDefault()
    if (isValid) {
      onRunClick?.(expr)
    }
  }

  useEffect(() => {
    setIsValid(!!expr.trim().length)
  }, [expr, setIsValid])

  useEffect(() => {
    if (!isOpen) {
      return
    }

    // Dirty hack to make input focus work.
    // Unlike RAF, setTimeout works when editor was previously focused.
    setTimeout(() => inputRef.current?.focus(), 0)
  }, [isOpen, inputRef])

  return (
    <>
      <Popover.Title className={popoverStyles.title}>Run expression</Popover.Title>
      <form onSubmit={onSubmit}>
        <div className={stack({ gap: '4', mt: '4' })}>
          <div>Please enter value of expression to evaluate:</div>
          <input
            type="text"
            name="expr"
            ref={inputRef}
            value={expr}
            className={cx('a', inputStyles.root, css({ fontFamily: 'mono' }))}
            placeholder='For example: main() or println("Hello world")'
            onChange={({ target: { value } }) => setExpr(value.trim())}
          />
          <div className={hstack({ justifyContent: 'flex-end' })}>
            <button type="submit" className={button()} disabled={!isValid}>
              Run
            </button>
          </div>
        </div>
      </form>
    </>
  )
})
interface RunPopoverProps extends RunPopoverContentProps {
  onVisibilityChange?: (val: boolean) => void
}

export const RunPopover: React.FC<RunPopoverProps> = observer(({ isOpen, onVisibilityChange, ...props }) => {
  const popoverStyles = popover({ size: 'lg' })

  return (
    <Popover.Root
      autoFocus
      open={isOpen}
      modal
      lazyMount
      onOpenChange={({ open }) => onVisibilityChange?.(open)}
      positioning={{ offset: { mainAxis: 14 } }}
    >
      <Popover.Trigger className={cx(css({ paw: 'Click+Header+Run+Expression' }), link())}>
        <span className={visuallyHidden()}>Run Expression</span>
        <PiCaretDownFill size={14} />
      </Popover.Trigger>
      <Portal>
        <Popover.Positioner data-testid="run-popover">
          <Popover.Content className={popoverStyles.content}>
            <Popover.Arrow className={popoverStyles.arrow}>
              <Popover.ArrowTip className={popoverStyles.arrowTip} />
            </Popover.Arrow>
            <Popover.CloseTrigger className={popoverStyles.close}>
              <PiX size={16} />
            </Popover.CloseTrigger>
            <RunPopoverContent isOpen={isOpen} {...props} />
          </Popover.Content>
        </Popover.Positioner>
      </Portal>
    </Popover.Root>
  )
})
