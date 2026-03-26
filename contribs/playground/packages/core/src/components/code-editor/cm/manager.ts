import { type EditorState } from '@codemirror/state'

/**
 * Manages states of opened files and responsible for keeping scroll and cursor state per file view.
 */
export class EditorStateManager {
  private readonly fileStates = new Map<string, EditorState>()

  replaceFileState(filePath: string, state: EditorState) {
    this.fileStates.set(filePath, state)
  }

  deleteFileState(filePath: string) {
    this.fileStates.delete(filePath)
  }

  getFileState(filePath: string) {
    return this.fileStates.get(filePath)
  }

  clear() {
    this.fileStates.clear()
  }
}
