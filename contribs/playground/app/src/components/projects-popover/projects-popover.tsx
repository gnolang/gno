import React from 'react'
import { PiArrowUpRight, PiTrashSimpleBold } from 'react-icons/pi'

import { dateFromNow } from '@gnostudio/core'
import { Popover, Portal } from '@gnostudio/react'

import { observer } from 'mobx-react-lite'

import { useStore } from '@/contexts'
import { type ProjectType } from '@/store/projects'
import { css, cx } from '@/styled-system/css'
import { center, hstack, stack, visuallyHidden } from '@/styled-system/patterns'
import { button, popover } from '@/styled-system/recipes'

import { EditableProjectName } from '../editable-project-name'

const ProjectsList: React.FC = observer(() => {
  const store = useStore()

  const handleLoad = (project: ProjectType) => {
    store.workbench.loadFromProject(project.id)
  }

  const handleDelete = (project: ProjectType) => {
    store.projects.delete(project.id)
  }

  return (
    <div className={stack({ mt: 3, mx: -3, px: 3, maxH: '300px', overflowY: 'auto', gap: '1' })}>
      {store.projects.list.length === 0 && (
        <p className={css({ color: 'foreground.muted' })}>Your projects will appear here once you save them.</p>
      )}

      <ul>
        {store.projects.list.map((project) => (
          <li key={project.id} className={css({ py: '3', borderBottom: '1px solid', borderColor: 'border' })}>
            <div className={hstack({ justify: 'space-between' })}>
              <div>
                <div className={hstack({ gap: 2 })}>
                  <EditableProjectName project={project} />

                  {store.projects.activeId === project.id && (
                    <span title="Active" className={center({ w: '2', h: '2', bg: 'foreground', rounded: 'full' })} />
                  )}
                </div>

                <p className={css({ fontSize: 'sm', color: 'foreground.muted' })}>
                  Last saved {dateFromNow(project.timestamp)}
                </p>
              </div>

              <div className={hstack({ gap: 2 })}>
                {project.isDraft && <span className={css({ fontSize: 'xs', px: '2', bg: 'primary' })}>Draft</span>}

                <button
                  title="Open project"
                  className={cx(css({ paw: 'Click+Project+Open' }), button({}))}
                  onClick={() => handleLoad(project)}
                >
                  <PiArrowUpRight />
                </button>

                <button
                  className={cx(css({ paw: 'Click+Project+Delete' }), button({ variant: 'outline' }))}
                  onClick={() => handleDelete(project)}
                >
                  <span className={visuallyHidden()}>Delete</span>
                  <PiTrashSimpleBold />
                </button>
              </div>
            </div>
          </li>
        ))}
      </ul>
    </div>
  )
})

interface Props {
  isOpen?: boolean
  onClose?: () => void
  children: React.ReactNode
}

export const ProjectsPopoverContainer: React.FC<Props> = observer(({ children, isOpen, onClose }) => {
  const store = useStore()
  const popoverStyles = popover({ size: 'md' })

  const handleNewProject = () => {
    if (store.resetWorkbench()) {
      store.projects.saveDraft()
    }
  }

  return (
    <Popover.Root
      modal
      open={isOpen}
      lazyMount
      unmountOnExit
      onOpenChange={({ open }) => !open && onClose?.()}
      positioning={{ offset: { mainAxis: 8 } }}
    >
      <Portal>
        <Popover.Positioner data-testid="projects-popover">
          <Popover.Content className={cx(popoverStyles.content)}>
            <Popover.Arrow className={popoverStyles.arrow}>
              <Popover.ArrowTip className={popoverStyles.arrowTip} />
            </Popover.Arrow>

            <Popover.Title className={popoverStyles.title}>
              <span>My Projects</span>
              <div className={css({ flex: '1' })} />

              {store.projects.hasActive && (
                <button
                  className={cx(
                    css({ paw: 'Click+Popover+NewProject' }),
                    button({ variant: 'ghost' }),
                    css({ alignSelf: 'flex-end' }),
                  )}
                  onClick={handleNewProject}
                >
                  + New project
                </button>
              )}
            </Popover.Title>

            <ProjectsList />
          </Popover.Content>
        </Popover.Positioner>
      </Portal>

      {children}
    </Popover.Root>
  )
})
