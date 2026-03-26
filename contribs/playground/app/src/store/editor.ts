import {
  EditorEventType,
  EditorMessageSeverity,
  keybindingsStore,
  type CursorPosition,
  type EditorController,
  type EditorEvent,
  type SelectionRange,
  type Size,
} from '@gnostudio/core'

import { reaction } from 'mobx'
import { addDisposer, getRoot, types } from 'mobx-state-tree'

import { type RootStore } from '.'
import { MessageType } from './types/status-message'

const logLevelMap: Partial<Record<EditorMessageSeverity, MessageType | null>> = {
  [EditorMessageSeverity.Log]: null,
  [EditorMessageSeverity.Info]: MessageType.Info,
  [EditorMessageSeverity.Warning]: MessageType.Warning,
  [EditorMessageSeverity.Error]: MessageType.Error,
}

export const Editor = types
  .model({})
  .volatile(() => ({
    isLoading: false,
    editorCtrl: null as EditorController | null,
    position: { column: 0, lineNumber: 0 },
    highlightedRange: null as SelectionRange | null,
    keybindingsStore,
  }))
  .views((self) => ({
    get highlightedLine() {
      return self.highlightedRange?.startLineNumber
    },
  }))
  .actions((self) => ({
    setPosition(position: CursorPosition) {
      self.position = position
    },

    setHighlightedRange(range: SelectionRange | null) {
      self.highlightedRange = range
    },
    setIsLoading(isBusy: boolean) {
      self.isLoading = isBusy
    },
  }))
  .actions((self) => ({
    formatDocument() {
      self.editorCtrl?.formatDocument()
    },

    changeLayout(size: Size) {
      self.editorCtrl?.resize?.(size)
    },

    setEditor(ctrl: EditorController) {
      self.editorCtrl?.dispose()
      const root = getRoot<RootStore>(self)

      self.editorCtrl = ctrl
      ctrl.onGutterClick = (range) => {
        ctrl.highlightLine(range.startLineNumber)
      }
      ctrl.onCursorPositionChange = (pos) => {
        self.setPosition(pos)
      }
      ctrl.onLoadingStateChange = (isLoading) => {
        self.setIsLoading(isLoading)
      }

      // Apply keymap
      self.keybindingsStore.attachInputListener(ctrl)
      void ctrl.setInputMode?.(self.keybindingsStore.currentInputMode)
      root.boot()
    },

    handleEditorEvent(event: EditorEvent) {
      const root = getRoot<RootStore>(self)
      switch (event.type) {
        case EditorEventType.None:
          root.workbench.setStatusMessage(null)
          break

        case EditorEventType.Notification: {
          if (event.severity === EditorMessageSeverity.Debug) {
            // swallow debug logs
            return
          }

          root.workbench.setStatusMessage({
            type: logLevelMap[event.severity] ?? null,
            tag: event.tag ?? null,
            text: event.message,
          })
          break
        }

        case EditorEventType.Progress: {
          if (event.finished) {
            root.workbench.setStatusMessage(null)
            return
          }

          const message = event.message ?? root.workbench.statusMessage?.text
          root.workbench.setStatusMessage({
            type: MessageType.Progress,
            tag: event.tag ?? null,
            text: message ?? null,
          })
          break
        }
      }
    },

    highlightLine(line: number) {
      self.editorCtrl?.highlightLine(line)
    },

    afterCreate() {
      const highlightDisposer = reaction(
        () => self.highlightedRange,
        () => self.editorCtrl?.removeHighlight(),
      )

      addDisposer(self, highlightDisposer)
    },
  }))
