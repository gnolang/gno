import type { MemPackage } from './types'

/**
 * Payload for package run transaction.
 *
 * See: https://docs.adena.app/integrations/transactions/sign-and-send-a-transaction#vm.m_run
 */
export interface RunMessageInput {
  caller: string
  send: string
  package: MemPackage
}

export function createRunMessage(input: RunMessageInput) {
  const value: Record<string, unknown> = {
    caller: input.caller,
    send: input.send,
    package: input.package,
  }

  return {
    type: '/vm.m_run',
    value,
  }
}
