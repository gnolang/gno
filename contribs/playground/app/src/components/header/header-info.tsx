import React, { memo, type ReactNode } from 'react'
import { PiX } from 'react-icons/pi'

import { hstack } from '@/styled-system/patterns'

interface Props {
  children: ReactNode
  clearInfo: () => void
}

const HeaderInfoComponent: React.FC<Props> = ({ children, clearInfo }: Props) => {
  if (!children) {
    return null
  }
  return (
    <div
      className={hstack({
        gap: '8',
        justify: 'center',
        px: '8',
        py: '2',
        bg: 'gray.950',
        position: 'relative',
        color: '{colors.white}',
      })}
      role="status"
      aria-live="polite"
    >
      <span
        id="info-message"
        role="alert"
        className={hstack({
          justify: 'center',
          width: '100%',
        })}
      >
        {children}
      </span>

      <button
        onClick={clearInfo}
        aria-label="Close info message"
        aria-controls="info-message"
        style={{
          position: 'absolute',
          right: '8px',
        }}
      >
        <PiX />
      </button>
    </div>
  )
}

export const HeaderInfo = memo(HeaderInfoComponent)
