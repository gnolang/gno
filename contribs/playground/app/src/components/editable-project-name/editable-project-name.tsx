import React, { useEffect } from 'react'
import { PiCheck, PiPencilSimple } from 'react-icons/pi'

import { Editable } from '@gnostudio/react'

import { observer } from 'mobx-react-lite'

import { useStore } from '@/contexts'
import { type ProjectType } from '@/store/projects'
import { css, cx } from '@/styled-system/css'
import { hstack } from '@/styled-system/patterns'
import { editable, input } from '@/styled-system/recipes'

interface Props {
  project: ProjectType
  showIcon?: boolean
}

export const EditableProjectName: React.FC<Props> = observer(({ project, showIcon }) => {
  const store = useStore()
  const editableStyles = editable()
  const inputStyles = input({ body: 'strict', size: 'xs' })
  const [value, setValue] = React.useState(project.title)

  const handleCommit = (value: string) => {
    const sanitized = value.trim()

    if (sanitized !== project.title) {
      store.projects.update(project.id, { title: sanitized })
    }
  }

  useEffect(() => {
    setValue(project.title ?? '')
  }, [project])

  return (
    <Editable.Root
      aria-label="Project name"
      value={value}
      placeholder={project.title || 'Untitled Project'}
      activationMode="dblclick"
      className={editableStyles.root}
      onValueChange={(details) => setValue(details.value)}
      onValueCommit={(details) => handleCommit(details.value)}
    >
      <Editable.Context>
        {(state) => (
          <div className={hstack({ gap: '1' })}>
            {showIcon && <PiPencilSimple />}
            <Editable.Area data-testid="editable-project-name" className={editableStyles.area}>
              <Editable.Input className={cx(editableStyles.input, inputStyles.root)} />
              <Editable.Preview className={cx('pawpal-event-name=Click+Project+Name', editableStyles.preview)}>
                <span className={css({ textDecoration: 'underline' })}>{project.title || 'Untitled Project'}</span>
              </Editable.Preview>
            </Editable.Area>

            <Editable.Control className={editableStyles.control}>
              {state.editing && (
                <Editable.SubmitTrigger
                  title="Confirm"
                  className={cx(css({ paw: 'Click+Project+UpdateName' }), editableStyles.submitTrigger)}
                >
                  <PiCheck />
                </Editable.SubmitTrigger>
              )}
            </Editable.Control>
          </div>
        )}
      </Editable.Context>
    </Editable.Root>
  )
})
