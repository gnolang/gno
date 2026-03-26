declare module '*.mdx' {
  let MDXComponent: (props: any) => JSX.Element
  export default MDXComponent
}

interface Window {
  showOpenFilePicker: (options?: ShowOpenFilePickerOptions) => Promise<FileSystemFileHandle[]>
}

interface ShowOpenFilePickerOptions {
  /**
   * A boolean value that indicates whether the user can select multiple files. Default is false.
   */
  multiple?: boolean

  /**
   * An array of file types that the file picker's file type control is limited to.
   */
  types?: FilePickerAcceptType[]

  /**
   * A boolean that, when set to true, suggests that the user can choose directories
   * instead of files. Default is false.
   */
  excludeAcceptAllOption?: boolean

  /**
   * A hint to the browser for which file picker to display. "open" indicates a file opener,
   * and "save" indicates a file saver. The default is "open".
   */
  startIn?: FileSystemHandle | string
}

interface FilePickerAcceptType {
  /**
   * Descriptive label for the file type.
   */
  description?: string

  /**
   * An array of MIME types or file extensions that can be selected.
   */
  accept: Record<string, string[]>
}
