import { types, type Instance } from 'mobx-state-tree'

import { type EditorController, type InputMode } from '../components'

const KeybindingsModel = types
  .model({})
  .volatile(() => ({
    currentLayout: undefined as InputMode | undefined,
    keyBuffer: '',
    mode: '',
  }))
  .views((self) => ({
    get currentInputMode() {
      return self.currentLayout ?? 'classic'
    },
  }))
  .actions((self) => ({
    setMode(mode: string) {
      self.mode = mode
    },
    setKeyBuffer(keyBuffer: string) {
      self.keyBuffer = keyBuffer
    },
  }))
  .actions((self) => {
    return {
      /**
       * Attaches store state to EditorController to listen for input mode changes.
       *
       * Designed for cases when editor component already accepts input method already passed as prop
       * and store only needs to attach vim/emacs-mode specific hooks.
       */
      attachInputListener(ctrl: EditorController) {
        ctrl.onInputModeChange = (listener) => {
          switch (listener?.mode) {
            case 'vim':
              listener.onModeChange((mode) => {
                self.setMode(mode)
              })
              listener.onKeyPress((key) => {
                const newKeyBuffer = key === ':' ? ':' : keybindingsStore.keyBuffer + key
                self.setKeyBuffer(newKeyBuffer)
              })
              listener.onCommandDone(() => self.setKeyBuffer(''))
              listener.onDispose(() => {
                self.setMode('')
                self.setKeyBuffer('')
              })
              break
            case 'emacs':
              listener.onDidMarkChange((e) => self.setMode(e ? 'Mark set' : 'Mark unset'))
              listener.onDidChangeKey((e) => self.setMode(e))
              break
            default:
              return
          }

          listener?.attach?.()
        }
      },
      setInputMode(ctrl: EditorController, layout: InputMode) {
        self.currentLayout = layout
        void ctrl.setInputMode?.(layout)
      },
    }
  })

export type KeybindingsModelType = Instance<typeof KeybindingsModel>

export const keybindingsStore = KeybindingsModel.create()
export const useKeybindingsStore = () => keybindingsStore
