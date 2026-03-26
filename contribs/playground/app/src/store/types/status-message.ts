import { types, type Instance } from 'mobx-state-tree'

export enum MessageType {
  /**
   * Informal message.
   */
  Info = 'info',

  /**
   * Represents ongoing loading or processing status.
   */
  Progress = 'progress',

  /**
   * Represents a warning message.
   */
  Warning = 'warning',

  /**
   * Represents failure message.
   */
  Error = 'error',
}

export const StatusMessageModel = types.model({
  type: types.maybeNull(types.enumeration<MessageType>('MessageType', Object.values(MessageType))),
  tag: types.maybeNull(types.string),
  text: types.maybeNull(types.string),
})

export type StatusMessage = Instance<typeof StatusMessageModel>
