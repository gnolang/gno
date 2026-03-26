import React from 'react'
import { PiX } from 'react-icons/pi'

import { Popover, Portal } from '@gnostudio/react'

import { observer } from 'mobx-react-lite'

import { useStore } from '@/contexts'
import { css, cx } from '@/styled-system/css'
import { link, popover } from '@/styled-system/recipes'

import { ConnectWalletStep } from './step-connect-wallet'
import { DeployStep } from './step-deploy'
import { InstallWalletStep } from './step-install-wallet'

export const DeployPopover: React.FC = observer(() => {
  const store = useStore()
  const step = store.deployer.step

  const popoverStyles = popover({ size: 'lg' })

  return (
    <Popover.Root autoFocus modal lazyMount positioning={{ offset: { mainAxis: 14 } }}>
      <Popover.Trigger className={cx(css({ paw: 'Click+Header+Deploy' }), link())}>Deploy</Popover.Trigger>
      <Portal>
        <Popover.Positioner>
          <Popover.Content data-testid="deploy-popover" className={popoverStyles.content}>
            <Popover.Arrow className={popoverStyles.arrow}>
              <Popover.ArrowTip className={popoverStyles.arrowTip} />
            </Popover.Arrow>
            <Popover.CloseTrigger className={popoverStyles.close}>
              <PiX size={16} />
            </Popover.CloseTrigger>
            {(step === 'install' || step === 'update') && <InstallWalletStep error={step} />}
            {step === 'connect' && <ConnectWalletStep />}
            {step === 'deploy' && <DeployStep />}
          </Popover.Content>
        </Popover.Positioner>
      </Portal>
    </Popover.Root>
  )
})
