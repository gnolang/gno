import React from 'react'
import { PiCheckBold, PiTrashFill } from 'react-icons/pi'

import { chainsEqual, type Chain } from '@gnostudio/core'

import { comboList } from '@/styled-system/recipes'

interface Props {
  selectedItem?: Chain | null
  predefinedItems: Chain[]
  editableItems: Chain[]
  disabled?: boolean
  onChange?: (c: Chain) => void
  onRemove?: (c: Chain, i: number) => void
}

interface ListItemProps extends Omit<Props, 'predefinedItems' | 'editableItems' | 'selectedItem'> {
  index: number
  item: Chain
  styles: ReturnType<typeof comboList>
  selected?: boolean
  removable?: boolean
}

const NetworkListItem: React.FC<ListItemProps> = ({
  styles,
  index,
  item,
  removable,
  selected,
  disabled,
  onChange,
  onRemove,
}) => {
  const canRemove = removable && !selected
  return (
    <li
      role="option"
      tabIndex={-1}
      className={styles.item}
      aria-disabled={disabled}
      aria-selected={selected}
      data-value={item.id}
      onClick={() => !disabled && !selected && onChange?.(item)}
    >
      <span className={styles.indicator} aria-hidden={true} hidden={!selected}>
        <PiCheckBold aria-hidden={true} />
      </span>
      <button className={styles.label} data-testid="select-chain-btn">
        {item.displayName}
      </button>
      {canRemove && (
        <span className={styles.actions}>
          <button
            className={styles.action}
            disabled={disabled}
            aria-label="Remove chain"
            data-testid="chain-remove-btn"
            onClick={(e) => {
              e.stopPropagation()
              onRemove?.(item, index)
            }}
          >
            <PiTrashFill aria-hidden={true} />
          </button>
        </span>
      )}
    </li>
  )
}

export const NetworkList: React.FC<Props> = ({
  selectedItem,
  predefinedItems,
  editableItems,
  disabled,
  onChange,
  onRemove,
}) => {
  const styles = comboList()
  return (
    <ul className={styles.container} role="combobox" tabIndex={-1} data-testid="network-list">
      {predefinedItems.map((item, i) => (
        <NetworkListItem
          key={item.rpcUrl}
          index={i}
          item={item}
          styles={styles}
          selected={chainsEqual(item, selectedItem)}
          disabled={disabled}
          onChange={onChange}
        />
      ))}
      {editableItems.map((item, i) => (
        <NetworkListItem
          key={item.rpcUrl}
          index={i}
          item={item}
          styles={styles}
          selected={chainsEqual(item, selectedItem)}
          disabled={disabled}
          onRemove={onRemove}
          onChange={onChange}
          removable
        />
      ))}
    </ul>
  )
}
