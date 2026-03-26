import type { MemPackage } from './types'

export interface AddPkgInput {
  creator: string
  deposit: string
  data: MemPackage
}

export function createAddPkgMessage(input: AddPkgInput) {
  const { data } = input

  const msg = {
    creator: input.creator,
    deposit: input.deposit,
    package: {
      name: data.name,
      path: data.path,
      files: data.files.map((item) => ({
        name: item.name,
        body: item.body,
      })),
    },
  }

  return {
    type: '/vm.m_addpkg',
    value: msg,
  }
}
