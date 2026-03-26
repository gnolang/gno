import * as Comlink from 'comlink'

import type { RunParams, WorkerHandler } from './go/types'
import GoWorker from './go/worker?worker&inline'

type ComlinkProxy = ReturnType<typeof Comlink.wrap<WorkerHandler>> & {
  terminate: () => void
}

export interface Worker extends Omit<ComlinkProxy, 'gnotest' | 'gnorun' | 'gnorepl'> {
  // Override signatures for methods that have transferable decorator.

  gnotest: WorkerHandler['gnotest']
  gnorun: WorkerHandler['gnorun']
  gnorepl: WorkerHandler['gnorepl']
}

/**
 * Wraps a Gno-related method that accepts port for TTY communication to automatically transfer passed MessagePort.
 */
const argsToTransferable = <T extends RunParams>(cb: (args: T) => Promise<number>) => {
  return (args: T) => cb(Comlink.transfer(args, [args.stdioPort]))
}

/**
 * Creates a new Go worker.
 */
export const createWorker = (): Worker => {
  const worker = new GoWorker()
  const wrap = Comlink.wrap<WorkerHandler>(worker)

  // Abstract transferrable objects logic away from user.
  // Spread operator doesn't work for Proxy instances.
  const { gnotest, gnorepl, gnorun } = wrap
  const target = Object.setPrototypeOf(
    {
      gnotest: argsToTransferable(gnotest),
      gnorepl: argsToTransferable(gnorepl),
      gnorun: argsToTransferable(gnorun),
      terminate: () => worker.terminate(),
      [Comlink.releaseProxy]: () => wrap[Comlink.releaseProxy](),
    },
    wrap,
  )

  return target
}
