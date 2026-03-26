import { useWorker, type Worker } from '@gnostudio/wasm'

import { useQueries, useQuery } from '@tanstack/react-query'

const getVersionsQuery = (worker: Worker) => ({
  queryKey: ['gnoVMWasmStore.versions'],
  queryFn: () => worker.getGnoVersions(),
})

const getVersionQuery = (worker: Worker) => ({
  queryKey: ['gnoVMWasmStore.version'],
  queryFn: () => worker.getGnoVersion(),
})

export function useGnoVMVersionQuery() {
  const worker = useWorker()

  return useQuery(getVersionQuery(worker))
}

export function useGnoVMQuery() {
  const worker = useWorker()

  return useQueries({ queries: [getVersionsQuery(worker), getVersionQuery(worker)] })
}
