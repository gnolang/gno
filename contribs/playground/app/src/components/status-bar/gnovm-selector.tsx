import React from 'react'
import { PiCaretDownFill } from 'react-icons/pi'

import { Menu } from '@gnostudio/react'
import { useWorker } from '@gnostudio/wasm'

import { useMutation } from '@tanstack/react-query'
import { observer } from 'mobx-react-lite'

import { useGnoVMQuery } from '@/hooks/use-gnovm-query'
import { hstack } from '@/styled-system/patterns'
import { menu } from '@/styled-system/recipes'

interface Props {
  allowChange?: boolean
}

export const GnoVMSelector: React.FC<Props> = observer(({ allowChange }) => {
  const menuClasses = menu()
  const worker = useWorker()

  const [versions, version] = useGnoVMQuery()

  const setVersion = useMutation({
    mutationKey: ['gnoVMWasmStore.setVersion'],
    mutationFn: (version: string) => worker.setGnoVersion(version),
    onSuccess: () => version.refetch(),
  })

  return (
    <Menu.Root onSelect={(e) => setVersion.mutate(e.value)}>
      <Menu.Trigger className={hstack({ gap: 1.5 })} disabled={!allowChange}>
        <span>GnoVM</span>
        {version.data as string}
        {allowChange && <PiCaretDownFill />}
      </Menu.Trigger>

      <Menu.Positioner>
        <Menu.Content className={menuClasses.content}>
          {versions.data?.map((v: string) => (
            <Menu.Item key={v} value={v} className={menuClasses.item}>
              {v}
            </Menu.Item>
          ))}
        </Menu.Content>
      </Menu.Positioner>
    </Menu.Root>
  )
})
