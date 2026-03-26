import React, { useEffect } from 'react'
import { PiX } from 'react-icons/pi'

import { Popover, Portal, useWriteClipboard } from '@gnostudio/react'

import { observer } from 'mobx-react-lite'

import { useStore } from '@/contexts'
import { useSaveToCloudMutation } from '@/hooks'
import { css, cx } from '@/styled-system/css'
import { stack } from '@/styled-system/patterns'
import { button, link, popover } from '@/styled-system/recipes'

const SuccessContent: React.FC = observer(() => {
  const store = useStore()
  const shareUrl = store.workbench.shareUrl as string

  const popoverStyles = popover({ size: 'sm' })

  const [writeToClipboard, isCopied] = useWriteClipboard()

  const handleCopy = () => {
    writeToClipboard(shareUrl).catch(console.error)
  }

  return (
    <>
      <Popover.Title className={popoverStyles.title}>Share successful</Popover.Title>
      <div className={stack({ gap: '4', mt: '4' })}>
        <div>
          <p>Your snippet is available at</p>
          <a href={shareUrl} target="_blank" className={link()} rel="noreferrer">
            {shareUrl}
          </a>
        </div>
        <div className={css({ display: 'flex', justifyContent: 'flex-end' })}>
          <button className={cx(css({ paw: 'Click+Popover+Share+Copy' }), button())} onClick={handleCopy}>
            {isCopied ? 'Copied' : 'Copy'}
          </button>
        </div>
      </div>
    </>
  )
})

const ShareContent: React.FC = () => {
  const { mutate: saveMutate, isPending, isError, isSuccess } = useSaveToCloudMutation()

  useEffect(() => {
    saveMutate()
  }, [saveMutate])

  return (
    <>
      {isPending && <p>Generating...</p>}
      {isError && <p>Something went wrong</p>}
      {isSuccess && <SuccessContent />}
    </>
  )
}

export const SharePopover: React.FC = () => {
  const popoverStyles = popover({ size: 'sm' })

  return (
    <Popover.Root autoFocus modal lazyMount unmountOnExit positioning={{ offset: { mainAxis: 14 } }}>
      <Popover.Trigger className={cx(css({ paw: 'Click+Header+Share' }), link())}>Share</Popover.Trigger>
      <Portal>
        <Popover.Positioner data-testid="share-popover">
          <Popover.Content className={popoverStyles.content}>
            <Popover.Arrow className={popoverStyles.arrow}>
              <Popover.ArrowTip className={popoverStyles.arrowTip} />
            </Popover.Arrow>
            <Popover.CloseTrigger className={popoverStyles.close}>
              <PiX size={16} />
            </Popover.CloseTrigger>
            <ShareContent />
          </Popover.Content>
        </Popover.Positioner>
      </Portal>
    </Popover.Root>
  )
}
