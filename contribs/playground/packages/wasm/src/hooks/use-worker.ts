import { createWorker } from '../workers'

const instance = createWorker()

export function useWorker() {
  // TODO: move worker singleton management outside of component.
  // Calling destructor here breaks gno studio and playground due to multiple re-renders killing a singleton.

  // useEffect(() => {
  //   const currentWorker = worker.current
  //
  //   return () => currentWorker[Comlink.releaseProxy]()
  // }, [])

  return instance
}
