import React, { useMemo, useState, type ComponentProps } from 'react'

import type { Chain } from '@gnostudio/core'
import { Popover, Portal } from '@gnostudio/react'

import { cx } from '@/styled-system/css'
import { popover } from '@/styled-system/recipes'

import { NetworkLabel } from './network-label'
import { NetworkList } from './network-list'
import { NewNetworkForm } from './new-network-form'
import { WalletConnectPrompt } from './wallet-connect-placeholder'

type InheritedProps = ComponentProps<typeof NetworkList> &
  ComponentProps<typeof NewNetworkForm> &
  ComponentProps<typeof WalletConnectPrompt>
interface Props extends InheritedProps {
  connected?: boolean
}

export const NetworkSelector: React.FC<Props> = ({
  disabled,
  connected,
  predefinedItems,
  editableItems,
  onAdd,
  onConnectRequest,
  selectedItem,
  chainIdProvider,
  walletAvailability,
  ...listProps
}) => {
  const [isOpen, setIsOpen] = useState(false)
  const styles = popover({ size: walletAvailability.isInstalled ? 'xs' : 'sm', layout: 'sections' })

  // Index of names and urls to validate existing items for NewNetworkForm
  const [labelsIndex, urlsIndex] = useMemo(() => {
    return predefinedItems.concat(editableItems).reduce(
      ([labels, urls], { displayName, rpcUrl }) => {
        return [labels.add(displayName), urls.set(rpcUrl, displayName)]
      },
      [new Set<string>(), new Map<string, string>()],
    )
  }, [predefinedItems, editableItems])

  const validateFormProp = (k: keyof Chain, v: string) => {
    // Reject duplicate labels or urls
    if (k === 'displayName') {
      return !labelsIndex.has(v)
    }

    return !urlsIndex.has(v)
  }

  return (
    <Popover.Root
      modal
      open={isOpen}
      lazyMount
      unmountOnExit
      onOpenChange={({ open }) => setIsOpen(open)}
      positioning={{ offset: { mainAxis: 8 } }}
    >
      <Portal>
        <Popover.Positioner data-testid="network-selector-popover">
          <Popover.Content className={cx(styles.content)}>
            <Popover.Arrow className={styles.arrow}>
              <Popover.ArrowTip className={styles.arrowTip} />
            </Popover.Arrow>
            {connected ? (
              <>
                <div className={styles.section}>
                  <span className={styles.sectionLabel}>Select Network</span>
                  <NetworkList
                    predefinedItems={predefinedItems}
                    editableItems={editableItems}
                    disabled={disabled}
                    selectedItem={selectedItem}
                    {...listProps}
                  />
                </div>
                <div className={styles.section}>
                  <span className={styles.sectionLabel}>Add Custom Network</span>
                  <NewNetworkForm
                    disabled={disabled}
                    onAdd={onAdd}
                    onValidate={validateFormProp}
                    chainIdProvider={chainIdProvider}
                  />
                </div>
              </>
            ) : (
              <WalletConnectPrompt
                onConnectRequest={onConnectRequest}
                walletAvailability={walletAvailability}
                disabled={disabled}
              />
            )}
          </Popover.Content>
        </Popover.Positioner>
      </Portal>
      <NetworkLabel value={selectedItem?.displayName} connected={connected} />
    </Popover.Root>
  )
}
