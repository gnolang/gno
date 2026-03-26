import { useLocation, useParams } from 'react-router-dom'

import { useMutation } from '@tanstack/react-query'

import { useStore } from '@/contexts'

export function useLoadWorkspace() {
  const store = useStore()
  const location = useLocation()
  const params = useParams<{ id?: string }>()

  return useMutation({
    mutationKey: ['LOAD_WORKSPACE', params.id, location.hash],
    mutationFn: async () => {
      if (params.id) {
        store.workbench.loadFromCloud(params.id)
      } else if (location.hash) {
        const hash = location.hash.slice(1)
        store.workbench.loadFromSerializedHash(hash)
      }
    },
  })
}

export function useSaveToCloudMutation() {
  const store = useStore()

  return useMutation<{ uri: string }>({
    mutationKey: ['CREATE_SESSION'],
    mutationFn: async () => {
      return store.workbench.saveToCloud()
    },
  })
}
