export interface CallMessageInput {
  caller: string
  send: string
  pkg_path: string
  func: string
  args: string[]
}

export function createCallMessage(input: CallMessageInput) {
  const msg: Record<string, unknown> = {
    caller: input.caller,
    send: input.send,
    pkg_path: input.pkg_path,
    func: input.func,
    args: input.args.length > 0 ? input.args : null,
  }

  return {
    type: '/vm.m_call',
    value: msg,
  }
}
