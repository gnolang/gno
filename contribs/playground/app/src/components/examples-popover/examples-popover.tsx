import React, { useEffect, useState } from 'react'
import { PiX } from 'react-icons/pi'

import { Popover, Portal } from '@gnostudio/react'

import { css, cx } from '@/styled-system/css'
import { link, popover } from '@/styled-system/recipes'

import { ExamplesContent } from './examples-content'

export const ExamplesPopover: React.FC = () => {
  const [isOpen, setIsOpen] = useState(false)
  const popoverStyles = popover({ size: 'lg' })

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if ((event.ctrlKey || event.metaKey) && event.key === 'e') {
        event.preventDefault()
        setIsOpen(!isOpen)
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => {
      window.removeEventListener('keydown', handleKeyDown)
    }
  }, [isOpen])

  useEffect(() => {
    if (window.location.hash === '#examples') {
      setIsOpen(true)
      window.history.replaceState(null, '', window.location.pathname + window.location.search)
    }
  }, [])

  const handleOpenChange = (details: { open: boolean }) => {
    setIsOpen(details.open)
  }

  return (
    <Popover.Root
      open={isOpen}
      onOpenChange={handleOpenChange}
      modal
      lazyMount
      positioning={{
        placement: 'bottom',
        offset: { mainAxis: 14 },
      }}
    >
      <Popover.Trigger asChild>
        <button
          onClick={() => setIsOpen(true)}
          className={cx(
            css({ paw: 'Click+Header+Examples' }),
            link({
              color: 'white',
              fontWeight: 'medium',
              _hover: {
                color: 'white',
                textDecoration: 'underline',
              },
            }),
          )}
        >
          Examples
        </button>
      </Popover.Trigger>

      <Portal>
        <Popover.Positioner>
          <Popover.Content
            className={cx(
              popoverStyles.content,
              css({
                width: '90vw',
                maxWidth: '1000px',
                height: '80vh',
                maxHeight: '700px',
              }),
            )}
          >
            <Popover.Arrow className={popoverStyles.arrow}>
              <Popover.ArrowTip className={popoverStyles.arrowTip} />
            </Popover.Arrow>
            <Popover.CloseTrigger className={popoverStyles.close}>
              <PiX size={16} />
            </Popover.CloseTrigger>
            <Popover.Title className={popoverStyles.title}>Examples</Popover.Title>
            <ExamplesContent onExampleSelected={() => setIsOpen(false)} />
          </Popover.Content>
        </Popover.Positioner>
      </Portal>
    </Popover.Root>
  )
}
