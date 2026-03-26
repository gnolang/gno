import React from 'react'
import { PiCaretDownFill } from 'react-icons/pi'

import { Popover } from '@gnostudio/react'

import { css, cx } from '@/styled-system/css'
import { hstack, stack } from '@/styled-system/patterns'
import { link, text } from '@/styled-system/recipes'

interface Props {
  value?: string
  connected?: boolean
}

export const NetworkLabel: React.FC<Props> = ({ value, connected }) => {
  return (
    <div className={stack({ gap: '0' })} data-testid="network-selector-label">
      <span className={text({ transform: 'uppercase', size: 'xs' })}>Network</span>
      <Popover.Trigger data-testid="network-selector-trigger">
        <span className={cx(hstack({ gap: '1' }))}>
          <span
            data-testid="network-selector-name"
            className={cx(link(), text({ ellipsis: true, align: 'left' }), css({ width: '128px' }))}
          >
            {connected ? (value ?? 'Not Selected') : 'Not Connected'}
          </span>
          <PiCaretDownFill />
        </span>
      </Popover.Trigger>
    </div>
  )
}
