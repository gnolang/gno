import { useWorker } from '@gnostudio/wasm'

import { parseErrorOutput } from '../../common/gofmt'
import type { FormatResult } from '../types'

export function useFormatter() {
  const worker = useWorker()

  return async (content: string): Promise<FormatResult> => {
    try {
      const result = await worker.gofmt(content)
      return {
        text: result.out,
      }
    } catch (err) {
      const errors = parseErrorOutput((err as Error).message)
      return {
        markers: errors,
      }
    }
  }
}
