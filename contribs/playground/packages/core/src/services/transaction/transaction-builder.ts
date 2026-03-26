import { createCallMessage, type CallMessageInput } from './transaction-types'
import { createAddPkgMessage, type AddPkgInput } from './transaction-types/add-pkg'
import { createRunMessage, type RunMessageInput } from './transaction-types/run'

interface Message {
  type: string
  value: Record<string, unknown>
}

export interface TransactionDocument {
  messages: Message[]
  gasFee: number
  gasWanted: number
  memo: string | undefined
}

/**
 * Mapping between transaction message type and gnokey command.
 *
 * Used to construct command in `buildCommand`.
 *
 * @see https://docs.adena.app/integrations/transactions/sign-and-send-a-transaction#vm.m_run (for 'addpkg' and 'run')
 */
const commandMessageMap = {
  '/vm.m_addpkg': 'addpkg',
  '/vm.m_call': 'call',
  '/vm.m_run': 'run',
} as const

export class TransactionBuilder {
  private readonly messages: Message[] = []
  private gasFee: number
  private gasWanted: number
  private memo: string | undefined

  constructor() {
    this.messages = []
    this.gasFee = 0
    this.gasWanted = 0
    this.memo = 'Deployed through play.gno.land'
  }

  addPkg(pkg: AddPkgInput) {
    const message = createAddPkgMessage(pkg)
    this.messages.push(message)
    return this
  }

  runPkg(input: RunMessageInput) {
    const message = createRunMessage(input)
    this.messages.push(message)
    return this
  }

  call(input: CallMessageInput) {
    this.messages.push(createCallMessage(input))
    return this
  }

  setGas({ gasFee, gasWanted }: { gasFee: number; gasWanted: number }) {
    this.gasFee = gasFee
    this.gasWanted = gasWanted
    return this
  }

  setMemo(memo: string) {
    this.memo = memo
    return this
  }

  buildCommand({ rpcUrl, networkId }: { rpcUrl?: string; networkId?: string } = {}) {
    const commands: string[] = []

    for (const message of this.messages) {
      const action = commandMessageMap[message.type as keyof typeof commandMessageMap]

      const params: Record<string, boolean | string | string[]> = {
        'gas-fee': `${this.gasFee.toString()}ugnot`,
        'gas-wanted': this.gasWanted.toString(),
        broadcast: true,
      }

      if (this.memo) params.memo = this.memo
      if (networkId) params.chainid = networkId
      if (rpcUrl) params.remote = rpcUrl

      let address = ''

      const cmd = `gnokey maketx ${action}`
      let suffix = ''

      switch (action) {
        case 'call':
          address = message.value.caller as string
          params.send = message.value.send as string
          params.pkgpath = message.value.pkg_path as string
          params.func = message.value.func as string
          params.args = (message.value.args as string[]) ?? []
          break
        case 'addpkg':
          address = message.value.creator as string
          params.deposit = message.value.deposit as string
          params.name = (message.value.package as any).name as string
          params.pkgpath = (message.value.package as any).path as string
          params.pkgdir = './'
          break
        case 'run':
          suffix = ' path_to_a_script.gno'
          break
      }

      const parseQuotes = (str = '') => str.replace(/(["'`])/g, '\\$1')
      if (suffix !== '') {
        suffix = ` ${suffix}`
      }

      const cmdParams = Object.entries(params)
        .map(([key, value]) => {
          if (Array.isArray(value)) {
            return value.map((item) => `-${key}="${parseQuotes(item)}"`).join(' ')
          }

          if (typeof value === 'boolean') {
            return `-${key}=${String(value)}`
          }

          return `-${key}="${parseQuotes(value)}"`
        })
        .join(' ')

      commands.push(`${cmd} ${address} ${cmdParams}${suffix}`)
    }

    return commands
  }

  build(): TransactionDocument {
    const transaction = {
      messages: this.messages,
      gasFee: this.gasFee,
      gasWanted: this.gasWanted,
      memo: this.memo,
    }

    return transaction
  }
}
