import React from 'react'
import { PiCaretDownFill } from 'react-icons/pi'

import { Popover } from '@gnostudio/react'

import { observer } from 'mobx-react-lite'

import { useStore } from '@/contexts'
import { css, cx } from '@/styled-system/css'
import { divider, hstack, visuallyHidden } from '@/styled-system/patterns'
import { link } from '@/styled-system/recipes'

import { EditableProjectName } from '../editable-project-name'
import { type PageType } from '../layout'
import { ConnectedNetworkSelector } from '../network-selector'
import { ProjectsPopoverContainer } from '../projects-popover'
import { SettingsPopover } from '../settings-popover'
import { ThemeSwitcher } from '../theme-switcher'

interface Props {
  type: PageType
}

const ProjectsToolbar: React.FC = observer(() => {
  const store = useStore()

  const handleSave = () => {
    store.projects.save()
  }

  return (
    <div className={hstack({ gap: '3' })}>
      {store.projects.hasActive && <EditableProjectName project={store.projects.active as any} showIcon />}

      <div className={hstack({ gap: '1' })}>
        <button
          onClick={handleSave}
          className={cx(css({ paw: 'Click+Header+SaveProject' }), link())}
          disabled={!store.projects.hasUnsavedChanges}
        >
          {store.projects.hasActive ? 'Save' : 'Save as draft'}
        </button>

        <ProjectsPopoverContainer>
          <Popover.Trigger>
            <span className={visuallyHidden()}>My Projects</span>
            <PiCaretDownFill className={css({ paw: 'Click+Header+Projects' })} />
          </Popover.Trigger>
        </ProjectsPopoverContainer>
      </div>
    </div>
  )
})

export const HeaderSettings: React.FC<Props> = observer(({ type }) => {
  return (
    <>
      {type === 'play' && (
        <>
          <ProjectsToolbar />
          <ConnectedNetworkSelector />
        </>
      )}
      <div className={hstack({ gap: '0' })}>
        <ThemeSwitcher />
        {type === 'play' && (
          <>
            <hr className={divider({ orientation: 'vertical', height: '4' })} />
            <SettingsPopover />
          </>
        )}
      </div>
    </>
  )
})
