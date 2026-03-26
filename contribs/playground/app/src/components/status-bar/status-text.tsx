import React from 'react'
import { PiInfoBold, PiWarningBold, PiWarningCircleBold } from 'react-icons/pi'

import { observer } from 'mobx-react-lite'

import { useStore } from '@/contexts'
import { statusText } from '@/styled-system/recipes'

import { MessageType } from '../../store/types/status-message'

const iconsMapping: Partial<Record<MessageType, React.ComponentType>> = {
  [MessageType.Warning]: PiWarningBold,
  [MessageType.Error]: PiWarningCircleBold,
  [MessageType.Info]: PiInfoBold,
}

const getIconComponent = (className: string, msgType?: MessageType | null) => {
  if (!msgType) {
    return null
  }

  const IconComponent = iconsMapping[msgType]
  if (!IconComponent) {
    return null
  }

  return (
    <span className={className} aria-hidden>
      <IconComponent />
    </span>
  )
}

export const StatusText: React.FC = observer(() => {
  const store = useStore()
  const { statusMessage } = store.workbench
  if (!statusMessage) {
    return null
  }

  const { type, tag, text } = statusMessage
  const tagVisible = tag && text
  const isProgress = type === MessageType.Progress
  const styles = statusText()

  return (
    <div className={styles.root}>
      {isProgress && <i className={styles.progress} aria-hidden />}
      {getIconComponent(styles.icon, type)}
      {tagVisible && <span className={styles.tag}>{tag}</span>}
      <span className={styles.text}>{text}</span>
    </div>
  )
})
